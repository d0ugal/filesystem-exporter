package collectors

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
	"filesystem-exporter/internal/utils"

	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

type DirectoryCollector struct {
	config  *config.Config
	metrics *metrics.FilesystemRegistry
	app     *app.App
	duMutex sync.Mutex // Mutex to ensure only one du operation at a time
}

func NewDirectoryCollector(cfg *config.Config, registry *metrics.FilesystemRegistry, app *app.App) *DirectoryCollector {
	return &DirectoryCollector{
		config:  cfg,
		metrics: registry,
		app:     app,
	}
}

// Stop stops the collector
func (dc *DirectoryCollector) Stop() {
	// No cleanup needed for this collector
}

func (dc *DirectoryCollector) Start(ctx context.Context) {
	go dc.run(ctx)
}

func (dc *DirectoryCollector) run(ctx context.Context) {
	// Create individual tickers for each directory
	tickers := make(map[string]*time.Ticker)

	defer func() {
		for _, ticker := range tickers {
			ticker.Stop()
		}
	}()

	// Start individual tickers for each directory
	for groupName, group := range dc.config.Directories {
		interval := dc.config.GetDirectoryInterval(group)
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		tickers[groupName] = ticker

		// Initial collection for this directory
		dc.collectSingleDirectory(ctx, groupName, group)

		// Start goroutine for this directory
		go func(name string, dir config.DirectoryGroup) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					dc.collectSingleDirectory(ctx, name, dir)
				}
			}
		}(groupName, group)
	}

	// Wait for context cancellation
	<-ctx.Done()
	slog.Info("Directory collector stopped")
}

func (dc *DirectoryCollector) collectSingleDirectory(ctx context.Context, groupName string, group config.DirectoryGroup) {
	startTime := time.Now()
	collectionType := "directory"
	interval := dc.config.GetDirectoryInterval(group)

	slog.Info("Starting directory metrics collection", "group", groupName)

	// Create span for collection cycle
	tracer := dc.app.GetTracer()

	var collectorSpan *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		collectorSpan = tracer.NewCollectorSpan(ctx, "directory-collector", "collect-directory")

		collectorSpan.SetAttributes(
			attribute.String("directory.group", groupName),
			attribute.String("directory.path", group.Path),
			attribute.Int("directory.interval", interval),
		)
		defer collectorSpan.End()
	}

	// Get span context for retry if tracing is enabled
	//nolint:contextcheck
	var retryCtx context.Context
	if collectorSpan != nil {
		retryCtx = collectorSpan.Context()
	} else {
		retryCtx = ctx
	}

	// Retry with exponential backoff
	err := dc.retryWithBackoff(retryCtx, func() error {
		// Pass the span context through to collection
		//nolint:contextcheck
		var collectCtx context.Context
		if collectorSpan != nil {
			collectCtx = collectorSpan.Context()
		} else {
			collectCtx = ctx
		}

		return dc.collectDirectoryGroup(collectCtx, groupName, group, collectionType)
	}, 3, 2*time.Second)
	if err != nil {
		slog.Error("Failed to collect directory group metrics after retries", "group", groupName, "error", err)

		if collectorSpan != nil {
			collectorSpan.RecordError(err, attribute.String("directory.group", groupName))
		}

		dc.metrics.CollectionFailedCounter.With(prometheus.Labels{
			"type":             collectionType,
			"group":            groupName,
			"interval_seconds": strconv.Itoa(interval),
		}).Inc()

		return
	}

	dc.metrics.CollectionSuccess.With(prometheus.Labels{
		"type":             collectionType,
		"group":            groupName,
		"interval_seconds": strconv.Itoa(interval),
	}).Inc()
	// Expose configured interval as a numeric gauge for PromQL arithmetic
	dc.metrics.CollectionIntervalGauge.With(prometheus.Labels{
		"group": groupName,
		"type":  collectionType,
	}).Set(float64(interval))

	duration := time.Since(startTime).Seconds()
	dc.metrics.CollectionDuration.With(prometheus.Labels{
		"group":            groupName,
		"interval_seconds": strconv.Itoa(interval),
		"type":             collectionType,
	}).Set(duration)
	dc.metrics.CollectionTimestampGauge.With(prometheus.Labels{
		"group":            groupName,
		"interval_seconds": strconv.Itoa(interval),
		"type":             collectionType,
	}).Set(float64(time.Now().Unix()))

	if collectorSpan != nil {
		collectorSpan.SetAttributes(
			attribute.Float64("collection.duration_seconds", duration),
		)
		collectorSpan.AddEvent("collection_completed",
			attribute.String("directory.group", groupName),
			attribute.Float64("duration_seconds", duration),
		)
	}

	slog.Info("Directory metrics collection completed", "group", groupName, "duration", duration)
}

// retryWithBackoff implements exponential backoff retry logic
func (dc *DirectoryCollector) retryWithBackoff(ctx context.Context, operation func() error, maxRetries int, initialDelay time.Duration) error {
	return utils.RetryWithBackoff(ctx, operation, maxRetries, initialDelay)
}

func (dc *DirectoryCollector) collectDirectoryGroup(ctx context.Context, groupName string, group config.DirectoryGroup, collectionType string) error {
	tracer := dc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context //nolint:contextcheck
	)

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "collect-directory-group")

		span.SetAttributes(
			attribute.String("directory.group", groupName),
			attribute.String("directory.path", group.Path),
			attribute.Int("directory.subdirectory_levels", group.SubdirectoryLevels),
		)

		spanCtx = span.Context()
		defer span.End()
	} else {
		spanCtx = ctx
	}

	return dc.collectDirectorySizes(spanCtx, groupName, group, collectionType)
}

func (dc *DirectoryCollector) collectDirectorySizes(ctx context.Context, groupName string, group config.DirectoryGroup, collectionType string) error {
	tracer := dc.app.GetTracer()

	//nolint:contextcheck
	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context
	)

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "collect-directory-sizes")

		spanCtx = span.Context()

		span.SetAttributes(
			attribute.String("directory.group", groupName),
			attribute.Int("directory.subdirectory_levels", group.SubdirectoryLevels),
		)

		defer span.End()
	} else {
		spanCtx = ctx
	}

	if group.SubdirectoryLevels == 0 {
		// Only collect the base directory
		return dc.collectSingleDirectoryFile(spanCtx, groupName, group.Path, collectionType, 0)
	}

	// Collect subdirectories recursively
	return dc.collectSubdirectories(spanCtx, groupName, group, collectionType)
}

//nolint:contextcheck
func (dc *DirectoryCollector) collectSingleDirectoryFile(ctx context.Context, groupName, path, collectionType string, subdirectoryLevel int) error {
	tracer := dc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "collect-single-directory")

		span.SetAttributes(
			attribute.String("directory.group", groupName),
			attribute.String("directory.path", path),
			attribute.Int("directory.subdirectory_level", subdirectoryLevel),
		)
		defer span.End()
	}

	// Validate and sanitize path
	if err := dc.validatePath(ctx, path); err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "validation",
		}).Inc()

		if span != nil {
			span.RecordError(err, attribute.String("operation", "validate_path"))
		}

		return fmt.Errorf("path validation failed for %s: %w", path, err)
	}

	// Acquire lock with tracing
	lockWaitDuration, err := dc.acquireDuLock(ctx)
	if err != nil {
		return err
	}
	defer dc.duMutex.Unlock()

	// Record lock wait duration
	dc.metrics.DuLockWaitDurationGauge.With(prometheus.Labels{
		"group": groupName,
		"path":  path,
	}).Set(lockWaitDuration.Seconds())

	if span != nil {
		span.SetAttributes(
			attribute.Float64("lock.wait_duration_seconds", lockWaitDuration.Seconds()),
		)
	}

	slog.Debug("Acquired du mutex lock", "path", path, "group", groupName, "wait_duration_ms", lockWaitDuration.Milliseconds())

	// Execute du command with tracing
	output, err := dc.executeDuCommand(ctx, path)
	if err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "du",
		}).Inc()

		if span != nil {
			span.RecordError(err, attribute.String("operation", "du_command"))
		}

		return fmt.Errorf("failed to execute du command for %s: %w", path, err)
	}

	// Parse du output with tracing
	sizeKB, err := dc.parseDuOutput(ctx, output)
	if err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "du",
		}).Inc()

		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_du_output"))
		}

		return fmt.Errorf("failed to parse directory size for %s: %w", path, err)
	}

	// Convert KB to bytes
	sizeBytes := sizeKB * 1024

	// Update metrics with tracing
	dc.updateDirectoryMetrics(ctx, groupName, path, sizeBytes, subdirectoryLevel)

	if span != nil {
		span.SetAttributes(
			attribute.Int64("directory.size_bytes", sizeBytes),
			attribute.Int64("directory.size_kb", sizeKB),
		)
		span.AddEvent("directory_collected",
			attribute.String("directory.path", path),
			attribute.Int64("size_bytes", sizeBytes),
		)
	}

	slog.Debug("Directory size collected",
		"group", groupName,
		"directory", path,
		"size_bytes", sizeBytes,
		"collection_type", collectionType,
	)

	return nil
}

func (dc *DirectoryCollector) collectSubdirectories(ctx context.Context, groupName string, group config.DirectoryGroup, collectionType string) error {
	tracer := dc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context
	)

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "collect-subdirectories")

		span.SetAttributes(
			attribute.String("directory.group", groupName),
			attribute.String("directory.base_path", group.Path),
			attribute.Int("directory.subdirectory_levels", group.SubdirectoryLevels),
		)

		spanCtx = span.Context()
		defer span.End()
	} else {
		spanCtx = ctx
	}

	// Validate and sanitize base path
	if err := dc.validatePath(spanCtx, group.Path); err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "validation",
		}).Inc()

		if span != nil {
			span.RecordError(err, attribute.String("operation", "validate_base_path"))
		}

		return fmt.Errorf("base path validation failed for %s: %w", group.Path, err)
	}

	// First, collect the base directory (level 0)
	if err := dc.collectSingleDirectoryFile(spanCtx, groupName, group.Path, collectionType, 0); err != nil {
		slog.Warn("Failed to collect base directory", "path", group.Path, "error", err)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "collect_base_directory"))
		}
		// Continue with subdirectories even if base directory fails
	}

	// Walk the directory tree up to the specified level
	walkStart := time.Now()

	var directoriesWalked, directoriesCollected, directoriesSkipped int

	err := filepath.WalkDir(group.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			slog.Warn("Error accessing path during walk", "path", path, "error", err)
			return nil // Continue walking, don't fail the entire collection
		}

		// Skip the root directory itself (we already collected it above)
		if path == group.Path {
			return nil
		}

		// Skip system directories
		dirName := filepath.Base(path)
		if dirName == "#recycle" || dirName == "@eaDir" || dirName == ".DS_Store" {
			return nil
		}

		// Calculate the depth relative to the base path
		relPath, err := filepath.Rel(group.Path, path)
		if err != nil {
			slog.Warn("Failed to calculate relative path", "path", path, "error", err)
			return nil
		}

		// Calculate depth based on directory components
		// Split the relative path and count the components
		pathComponents := strings.Split(relPath, string(filepath.Separator))

		// Handle special case: if relative path is "." (same directory), depth is 0
		var depth int
		if len(pathComponents) == 1 && pathComponents[0] == "." {
			depth = 0
		} else {
			depth = len(pathComponents) // Each component represents one level of depth
		}

		if !d.IsDir() {
			depth++ // Files are at depth + 1
		}

		directoriesWalked++

		// Collect directories up to the specified level (inclusive)
		if depth <= group.SubdirectoryLevels && d.IsDir() {
			// Collect directory size
			if err := dc.collectSingleDirectoryFile(spanCtx, groupName, path, collectionType, depth); err != nil {
				slog.Warn("Failed to collect subdirectory", "path", path, "error", err)

				if span != nil {
					span.RecordError(err, attribute.String("subdirectory.path", path), attribute.Int("subdirectory.depth", depth))
				}
				// Continue with other directories, don't fail the entire collection
			} else {
				directoriesCollected++
			}
		} else if depth > group.SubdirectoryLevels {
			// Skip deeper directories
			directoriesSkipped++
			return filepath.SkipDir
		}

		return nil
	})
	walkDuration := time.Since(walkStart)

	if err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "walk",
		}).Inc()

		if span != nil {
			span.RecordError(err, attribute.String("operation", "walk_directory_tree"))
		}

		return fmt.Errorf("failed to walk directory tree for %s: %w", group.Path, err)
	}

	if span != nil {
		span.SetAttributes(
			attribute.Int("walk.directories_walked", directoriesWalked),
			attribute.Int("walk.directories_collected", directoriesCollected),
			attribute.Int("walk.directories_skipped", directoriesSkipped),
			attribute.Float64("walk.duration_seconds", walkDuration.Seconds()),
		)
		span.AddEvent("walk_completed")
	}

	return nil
}

// validatePath ensures the path is safe to use with du command
func (dc *DirectoryCollector) validatePath(ctx context.Context, path string) error {
	tracer := dc.app.GetTracer()

	var span *tracing.CollectorSpan
	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "validate-path")

		span.SetAttributes(attribute.String("path", path))
		defer span.End()
	}

	// Check if path exists
	statStart := time.Now()

	if _, err := os.Stat(path); err != nil {
		if span != nil {
			span.RecordError(err, attribute.String("check", "exists"))
		}

		return fmt.Errorf("path does not exist: %s", path)
	}

	if span != nil {
		span.SetAttributes(
			attribute.Float64("validation.stat_duration_seconds", time.Since(statStart).Seconds()),
			attribute.Bool("validation.exists", true),
		)
	}

	// Ensure path is absolute
	if !filepath.IsAbs(path) {
		if span != nil {
			span.SetAttributes(attribute.Bool("validation.is_absolute", false))
			span.RecordError(fmt.Errorf("path must be absolute"), attribute.String("check", "absolute"))
		}

		return fmt.Errorf("path must be absolute: %s", path)
	}

	if span != nil {
		span.SetAttributes(attribute.Bool("validation.is_absolute", true))
	}

	// Basic sanitization - check for dangerous patterns
	dangerousPatterns := []string{"..", "~", "*", "?", "["}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(path, pattern) {
			if span != nil {
				span.SetAttributes(
					attribute.String("validation.dangerous_pattern", pattern),
					attribute.Bool("validation.is_safe", false),
				)
				span.RecordError(fmt.Errorf("dangerous pattern found"), attribute.String("pattern", pattern))
			}

			return fmt.Errorf("path contains dangerous pattern '%s': %s", pattern, path)
		}
	}

	if span != nil {
		span.SetAttributes(attribute.Bool("validation.is_safe", true))
		span.AddEvent("validation_completed")
	}

	return nil
}

// acquireDuLock acquires the du mutex lock with tracing
func (dc *DirectoryCollector) acquireDuLock(ctx context.Context) (time.Duration, error) {
	tracer := dc.app.GetTracer()

	var span *tracing.CollectorSpan
	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "acquire-du-lock")
		defer span.End()
	}

	lockWaitStart := time.Now()

	dc.duMutex.Lock()

	lockWaitDuration := time.Since(lockWaitStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("lock.wait_duration_seconds", lockWaitDuration.Seconds()),
		)
		span.AddEvent("lock_acquired")
	}

	return lockWaitDuration, nil
}

// executeDuCommand executes the du command with tracing
func (dc *DirectoryCollector) executeDuCommand(ctx context.Context, path string) ([]byte, error) {
	tracer := dc.app.GetTracer()

	//nolint:contextcheck
	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context
	)

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "execute-du-command")

		spanCtx = span.Context()

		span.SetAttributes(attribute.String("command.path", path))
		defer span.End()
	}

	// Create context with timeout (6 minutes max) - use span context if available
	timeoutCtx, cancel := context.WithTimeout(spanCtx, 6*time.Minute)
	defer cancel()

	// Use du with performance optimizations for large directories
	// -s: summarize only
	// -x: don't cross filesystem boundaries (faster)
	cmd := exec.CommandContext(timeoutCtx, "du", "-s", "-x", path)

	execStart := time.Now()
	output, err := cmd.Output()
	execDuration := time.Since(execStart)

	if span != nil {
		span.SetAttributes(
			attribute.String("command.name", "du"),
			attribute.String("command.args", fmt.Sprintf("-s -x %s", path)),
			attribute.Float64("command.duration_seconds", execDuration.Seconds()),
			attribute.Int("command.output_size_bytes", len(output)),
		)

		if err != nil {
			span.RecordError(err)
		} else {
			span.AddEvent("command_completed", attribute.Int("output_size_bytes", len(output)))
		}
	}

	return output, err
}

// parseDuOutput parses the du command output with tracing
func (dc *DirectoryCollector) parseDuOutput(ctx context.Context, output []byte) (int64, error) {
	tracer := dc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "parse-du-output")

		span.SetAttributes(
			attribute.Int("output.size_bytes", len(output)),
			attribute.String("output.preview", func() string {
				previewLen := 100
				if len(output) < previewLen {
					previewLen = len(output)
				}

				return string(output[:previewLen])
			}()),
		)
		defer span.End()
	}
	// Note: spanCtx is only used when tracing is enabled

	parseStart := time.Now()

	// Parse du output: "size\tpath"
	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		if span != nil {
			span.RecordError(fmt.Errorf("unexpected du output format"), attribute.Int("parts_count", len(parts)))
		}

		return 0, fmt.Errorf("unexpected du output format")
	}

	sizeKB, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		if span != nil {
			span.RecordError(err, attribute.String("field", "size_kb"), attribute.String("value", parts[0]))
		}

		return 0, fmt.Errorf("failed to parse directory size: %w", err)
	}

	if span != nil {
		span.SetAttributes(
			attribute.Int64("parse.size_kb", sizeKB),
			attribute.Float64("parse.duration_seconds", time.Since(parseStart).Seconds()),
		)
		span.AddEvent("parse_completed")
	}

	return sizeKB, nil
}

// updateDirectoryMetrics updates Prometheus metrics with tracing
func (dc *DirectoryCollector) updateDirectoryMetrics(ctx context.Context, groupName, path string, sizeBytes int64, subdirectoryLevel int) {
	tracer := dc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "directory-collector", "update-metrics")
		defer span.End()
	}
	// Note: spanCtx is only used when tracing is enabled

	updateStart := time.Now()

	dc.metrics.DirectorySizeGauge.With(prometheus.Labels{
		"group":              groupName,
		"directory":          path,
		"mode":               "du",
		"subdirectory_level": fmt.Sprintf("%d", subdirectoryLevel),
	}).Set(float64(sizeBytes))
	dc.metrics.DirectoriesProcessedCounter.With(prometheus.Labels{
		"group":  groupName,
		"method": "du",
	}).Inc()

	if span != nil {
		span.SetAttributes(
			attribute.Float64("metrics.update_duration_seconds", time.Since(updateStart).Seconds()),
			attribute.Int("metrics.count", 2),
		)
		span.AddEvent("metrics_updated")
	}
}
