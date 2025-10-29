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

	// Retry with exponential backoff
	err := dc.retryWithBackoff(func() error {
		return dc.collectDirectoryGroup(ctx, groupName, group, collectionType)
	}, 3, 2*time.Second)
	if err != nil {
		slog.Error("Failed to collect directory group metrics after retries", "group", groupName, "error", err)

		if collectorSpan != nil {
			collectorSpan.RecordError(err, attribute.String("directory.group", groupName))
		}

		dc.metrics.CollectionFailedCounter.With(prometheus.Labels{
			"collector": collectionType,
			"group":     groupName,
			"interval":  strconv.Itoa(interval),
		}).Inc()

		return
	}

	dc.metrics.CollectionSuccessCounter.With(prometheus.Labels{
		"collector": collectionType,
		"group":     groupName,
		"interval":  strconv.Itoa(interval),
	}).Inc()
	// Expose configured interval as a numeric gauge for PromQL arithmetic
	dc.metrics.CollectionIntervalGauge.With(prometheus.Labels{
		"group": groupName,
		"type":  collectionType,
	}).Set(float64(interval))

	duration := time.Since(startTime).Seconds()
	dc.metrics.CollectionDurationGauge.With(prometheus.Labels{
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
func (dc *DirectoryCollector) retryWithBackoff(operation func() error, maxRetries int, initialDelay time.Duration) error {
	var lastErr error

	delay := initialDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt < maxRetries {
				slog.Warn("Operation failed, retrying", "attempt", attempt+1, "max_retries", maxRetries, "delay", delay, "error", err)
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
			}
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries+1, lastErr)
}

func (dc *DirectoryCollector) collectDirectoryGroup(ctx context.Context, groupName string, group config.DirectoryGroup, collectionType string) error {
	return dc.collectDirectorySizes(ctx, groupName, group, collectionType)
}

func (dc *DirectoryCollector) collectDirectorySizes(ctx context.Context, groupName string, group config.DirectoryGroup, collectionType string) error {
	if group.SubdirectoryLevels == 0 {
		// Only collect the base directory
		return dc.collectSingleDirectoryFile(ctx, groupName, group.Path, collectionType, 0)
	}

	// Collect subdirectories recursively
	return dc.collectSubdirectories(ctx, groupName, group, collectionType)
}

func (dc *DirectoryCollector) collectSingleDirectoryFile(ctx context.Context, groupName, path, collectionType string, subdirectoryLevel int) error {
	// Validate and sanitize path
	if err := dc.validatePath(path); err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "validation",
		}).Inc()

		return fmt.Errorf("path validation failed for %s: %w", path, err)
	}

	// Create context with timeout (6 minutes max) - use background context to avoid cancellation propagation
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	// Track lock waiting time
	lockWaitStart := time.Now()

	dc.duMutex.Lock()

	lockWaitDuration := time.Since(lockWaitStart)

	defer dc.duMutex.Unlock()

	// Record lock wait duration
	dc.metrics.DuLockWaitDurationGauge.With(prometheus.Labels{
		"group": groupName,
		"path":  path,
	}).Set(lockWaitDuration.Seconds())

	slog.Debug("Acquired du mutex lock", "path", path, "group", groupName, "wait_duration_ms", lockWaitDuration.Milliseconds())

	// Use du with performance optimizations for large directories
	// -s: summarize only
	// -x: don't cross filesystem boundaries (faster)
	//nolint:contextcheck // Intentionally use background context to prevent cancellation propagation
	cmd := exec.CommandContext(timeoutCtx, "du", "-s", "-x", path)

	output, err := cmd.Output()
	if err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "du",
		}).Inc()

		return fmt.Errorf("failed to execute du command for %s: %w", path, err)
	}

	// Parse du output: "size\tpath"
	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "du",
		}).Inc()

		return fmt.Errorf("unexpected du output format for %s", path)
	}

	sizeKB, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "du",
		}).Inc()

		return fmt.Errorf("failed to parse directory size for %s: %w", path, err)
	}

	// Convert KB to bytes
	sizeBytes := sizeKB * 1024

	// Update metrics
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

	slog.Debug("Directory size collected",
		"group", groupName,
		"directory", path,
		"size_bytes", sizeBytes,
		"collection_type", collectionType,
	)

	return nil
}

func (dc *DirectoryCollector) collectSubdirectories(ctx context.Context, groupName string, group config.DirectoryGroup, collectionType string) error {
	// Validate and sanitize base path
	if err := dc.validatePath(group.Path); err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "validation",
		}).Inc()

		return fmt.Errorf("base path validation failed for %s: %w", group.Path, err)
	}

	// First, collect the base directory (level 0)
	if err := dc.collectSingleDirectoryFile(ctx, groupName, group.Path, collectionType, 0); err != nil {
		slog.Warn("Failed to collect base directory", "path", group.Path, "error", err)
		// Continue with subdirectories even if base directory fails
	}

	// Walk the directory tree up to the specified level
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

		// Collect directories up to the specified level (inclusive)
		if depth <= group.SubdirectoryLevels && d.IsDir() {
			// Collect directory size
			if err := dc.collectSingleDirectoryFile(ctx, groupName, path, collectionType, depth); err != nil {
				slog.Warn("Failed to collect subdirectory", "path", path, "error", err)
				// Continue with other directories, don't fail the entire collection
			}
		} else if depth > group.SubdirectoryLevels {
			// Skip deeper directories
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		dc.metrics.DirectoriesFailedCounter.With(prometheus.Labels{
			"group":  groupName,
			"reason": "walk",
		}).Inc()

		return fmt.Errorf("failed to walk directory tree for %s: %w", group.Path, err)
	}

	return nil
}

// validatePath ensures the path is safe to use with du command
func (dc *DirectoryCollector) validatePath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Ensure path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	// Basic sanitization - check for dangerous patterns
	dangerousPatterns := []string{"..", "~", "*", "?", "["}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(path, pattern) {
			return fmt.Errorf("path contains dangerous pattern '%s': %s", pattern, path)
		}
	}

	return nil
}
