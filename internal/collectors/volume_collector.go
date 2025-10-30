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
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
	"filesystem-exporter/internal/utils"
	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

type FilesystemCollector struct {
	config  *config.Config
	metrics *metrics.FilesystemRegistry
	app     *app.App
}

func NewFilesystemCollector(cfg *config.Config, registry *metrics.FilesystemRegistry, app *app.App) *FilesystemCollector {
	return &FilesystemCollector{
		config:  cfg,
		metrics: registry,
		app:     app,
	}
}

// Stop stops the collector
func (fc *FilesystemCollector) Stop() {
	// No cleanup needed for this collector
}

func (fc *FilesystemCollector) Start(ctx context.Context) {
	go fc.run(ctx)
}

func (fc *FilesystemCollector) run(ctx context.Context) {
	// Create individual tickers for each filesystem
	tickers := make(map[string]*time.Ticker)

	defer func() {
		for _, ticker := range tickers {
			ticker.Stop()
		}
	}()

	// Start individual tickers for each filesystem
	for _, filesystem := range fc.config.Filesystems {
		interval := fc.config.GetFilesystemInterval(filesystem)
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		tickers[filesystem.Name] = ticker

		// Initial collection for this filesystem
		fc.collectSingleFilesystem(ctx, filesystem)

		// Start goroutine for this filesystem
		go func(fs config.FilesystemConfig) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					fc.collectSingleFilesystem(ctx, fs)
				}
			}
		}(filesystem)
	}

	// Wait for context cancellation
	<-ctx.Done()
	slog.Info("Filesystem collector stopped")
}

func (fc *FilesystemCollector) collectSingleFilesystem(ctx context.Context, filesystem config.FilesystemConfig) {
	startTime := time.Now()
	collectionType := "filesystem"
	interval := fc.config.GetFilesystemInterval(filesystem)

	slog.Info("Starting filesystem metrics collection", "filesystem", filesystem.Name)

	// Create span for collection cycle
	tracer := fc.app.GetTracer()

	var collectorSpan *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		collectorSpan = tracer.NewCollectorSpan(ctx, "filesystem-collector", "collect-filesystem")

		collectorSpan.SetAttributes(
			attribute.String("filesystem.name", filesystem.Name),
			attribute.String("filesystem.mount_point", filesystem.MountPoint),
			attribute.String("filesystem.device", filesystem.Device),
			attribute.Int("filesystem.interval", interval),
		)
		defer collectorSpan.End()
	}

	// Get span context for retry if tracing is enabled
	var retryCtx context.Context
	if collectorSpan != nil {
		retryCtx = collectorSpan.Context()
	} else {
		retryCtx = ctx
	}

	// Retry with exponential backoff
	err := fc.retryWithBackoff(retryCtx, func() error {
		// Pass the span context through to collection
		var collectCtx context.Context
		if collectorSpan != nil {
			collectCtx = collectorSpan.Context()
		} else {
			collectCtx = ctx
		}

		return fc.collectFilesystemUsage(collectCtx, filesystem, collectorSpan)
	}, 3, 2*time.Second)
	if err != nil {
		slog.Error("Failed to collect filesystem metrics after retries", "filesystem", filesystem.Name, "error", err)

		if collectorSpan != nil {
			collectorSpan.RecordError(err, attribute.String("filesystem.name", filesystem.Name))
		}

		fc.metrics.CollectionFailedCounter.With(prometheus.Labels{
			"collector": collectionType,
			"group":     filesystem.Name,
			"interval":  strconv.Itoa(interval),
		}).Inc()

		return
	}

	fc.metrics.CollectionSuccess.With(prometheus.Labels{
		"type":             collectionType,
		"group":            filesystem.Name,
		"interval_seconds": strconv.Itoa(interval),
	}).Inc()
	// Expose configured interval as a numeric gauge for PromQL arithmetic
	fc.metrics.CollectionIntervalGauge.With(prometheus.Labels{
		"group": filesystem.Name,
		"type":  collectionType,
	}).Set(float64(interval))

	duration := time.Since(startTime).Seconds()
	fc.metrics.CollectionDuration.With(prometheus.Labels{
		"group":            filesystem.Name,
		"interval_seconds": strconv.Itoa(interval),
		"type":             collectionType,
	}).Set(duration)
	fc.metrics.CollectionTimestampGauge.With(prometheus.Labels{
		"group":            filesystem.Name,
		"interval_seconds": strconv.Itoa(interval),
		"type":             collectionType,
	}).Set(float64(time.Now().Unix()))

	if collectorSpan != nil {
		collectorSpan.SetAttributes(
			attribute.Float64("collection.duration_seconds", duration),
		)
		collectorSpan.AddEvent("collection_completed",
			attribute.String("filesystem.name", filesystem.Name),
			attribute.Float64("duration_seconds", duration),
		)
	}

	slog.Info("Filesystem metrics collection completed", "filesystem", filesystem.Name, "duration", duration)
}

// retryWithBackoff implements exponential backoff retry logic
func (fc *FilesystemCollector) retryWithBackoff(ctx context.Context, operation func() error, maxRetries int, initialDelay time.Duration) error {
	return utils.RetryWithBackoff(ctx, operation, maxRetries, initialDelay)
}

func (fc *FilesystemCollector) collectFilesystemUsage(ctx context.Context, filesystem config.FilesystemConfig, collectorSpan *tracing.CollectorSpan) error {
	tracer := fc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context
	)

	if tracer != nil && tracer.IsEnabled() && collectorSpan != nil {
		spanCtx = collectorSpan.Context()
		span = tracer.NewCollectorSpan(spanCtx, "filesystem-collector", "collect-filesystem-usage")

		span.SetAttributes(
			attribute.String("filesystem.name", filesystem.Name),
			attribute.String("filesystem.mount_point", filesystem.MountPoint),
		)
		defer span.End()
	} else {
		spanCtx = ctx
	}

	// Validate mount point
	if err := fc.validateMountPoint(spanCtx, filesystem.MountPoint, span); err != nil {
		if span != nil {
			span.RecordError(err, attribute.String("operation", "validate_mount_point"))
		}

		return fmt.Errorf("mount point validation failed for %s: %w", filesystem.MountPoint, err)
	}

	// Sanitize mount point for command execution
	sanitizedMountPoint := fc.sanitizePath(filesystem.MountPoint)

	// Execute df command with tracing
	output, err := fc.executeDfCommand(spanCtx, sanitizedMountPoint, span)
	if err != nil {
		if span != nil {
			span.RecordError(err, attribute.String("operation", "df_command"))
		}

		return fmt.Errorf("failed to execute df command: %w", err)
	}

	// Parse df output with tracing
	sizeKB, availableKB, err := fc.parseDfOutput(spanCtx, output, span)
	if err != nil {
		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_df_output"))
		}

		return err
	}

	// Convert KB to bytes
	sizeBytes := sizeKB * 1024
	availableBytes := availableKB * 1024
	usedBytes := sizeBytes - availableBytes
	usedRatio := float64(usedBytes) / float64(sizeBytes)

	// Update metrics with tracing
	fc.updateFilesystemMetrics(spanCtx, filesystem, sizeBytes, availableBytes, usedRatio, span)

	if span != nil {
		span.SetAttributes(
			attribute.Int64("filesystem.size_bytes", sizeBytes),
			attribute.Int64("filesystem.available_bytes", availableBytes),
			attribute.Int64("filesystem.used_bytes", usedBytes),
			attribute.Float64("filesystem.used_ratio", usedRatio),
		)
		span.AddEvent("metrics_collected",
			attribute.String("filesystem.name", filesystem.Name),
			attribute.Int64("size_bytes", sizeBytes),
		)
	}

	slog.Debug("Filesystem metrics collected",
		"filesystem", filesystem.Name,
		"mount_point", filesystem.MountPoint,
		"size_bytes", sizeBytes,
		"available_bytes", availableBytes,
		"used_ratio", usedRatio,
	)

	return nil
}

// validateMountPoint ensures the mount point is safe to use with df command
func (fc *FilesystemCollector) validateMountPoint(ctx context.Context, mountPoint string, parentSpan *tracing.CollectorSpan) error {
	tracer := fc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context
	)

	if tracer != nil && tracer.IsEnabled() && parentSpan != nil {
		spanCtx = parentSpan.Context()
		span = tracer.NewCollectorSpan(spanCtx, "filesystem-collector", "validate-mount-point")

		span.SetAttributes(attribute.String("mount_point", mountPoint))
		defer span.End()
	} else {
		spanCtx = ctx
	}

	// Check if mount point exists
	statStart := time.Now()

	if _, err := os.Stat(mountPoint); err != nil {
		if span != nil {
			span.RecordError(err, attribute.String("check", "exists"))
		}

		return fmt.Errorf("mount point does not exist: %s", mountPoint)
	}

	if span != nil {
		span.SetAttributes(
			attribute.Float64("validation.stat_duration_seconds", time.Since(statStart).Seconds()),
			attribute.Bool("validation.exists", true),
		)
	}

	// Ensure mount point is absolute
	if !filepath.IsAbs(mountPoint) {
		if span != nil {
			span.SetAttributes(attribute.Bool("validation.is_absolute", false))
			span.RecordError(fmt.Errorf("mount point must be absolute"), attribute.String("check", "absolute"))
		}

		return fmt.Errorf("mount point must be absolute: %s", mountPoint)
	}

	if span != nil {
		span.SetAttributes(attribute.Bool("validation.is_absolute", true))
	}

	// Basic sanitization - check for dangerous patterns
	dangerousPatterns := []string{"..", "~", "*", "?", "["}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(mountPoint, pattern) {
			if span != nil {
				span.SetAttributes(
					attribute.String("validation.dangerous_pattern", pattern),
					attribute.Bool("validation.is_safe", false),
				)
				span.RecordError(fmt.Errorf("dangerous pattern found"), attribute.String("pattern", pattern))
			}

			return fmt.Errorf("mount point contains dangerous pattern '%s': %s", pattern, mountPoint)
		}
	}

	if span != nil {
		span.SetAttributes(attribute.Bool("validation.is_safe", true))
		span.AddEvent("validation_completed")
	}

	return nil
}

// executeDfCommand executes the df command with tracing
func (fc *FilesystemCollector) executeDfCommand(ctx context.Context, mountPoint string, parentSpan *tracing.CollectorSpan) ([]byte, error) {
	tracer := fc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context
	)

	if tracer != nil && tracer.IsEnabled() && parentSpan != nil {
		spanCtx = parentSpan.Context()
		span = tracer.NewCollectorSpan(spanCtx, "filesystem-collector", "execute-df-command")

		span.SetAttributes(attribute.String("command.mount_point", mountPoint))
		defer span.End()
	} else {
		spanCtx = ctx
	}

	// Create context with timeout (10 seconds max) - use span context if available
	timeoutCtx, cancel := context.WithTimeout(spanCtx, 10*time.Second)
	defer cancel()

	// G204: Subprocess launched with variable - This is safe because mountPoint is validated and sanitized
	cmd := exec.CommandContext(timeoutCtx, "df", mountPoint)

	execStart := time.Now()
	output, err := cmd.Output()
	execDuration := time.Since(execStart)

	if span != nil {
		span.SetAttributes(
			attribute.String("command.name", "df"),
			attribute.String("command.args", mountPoint),
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

// updateFilesystemMetrics updates Prometheus metrics with tracing
func (fc *FilesystemCollector) updateFilesystemMetrics(ctx context.Context, filesystem config.FilesystemConfig, sizeBytes, availableBytes int64, usedRatio float64, parentSpan *tracing.CollectorSpan) {
	tracer := fc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context
	)

	if tracer != nil && tracer.IsEnabled() && parentSpan != nil {
		spanCtx = parentSpan.Context()

		span = tracer.NewCollectorSpan(spanCtx, "filesystem-collector", "update-metrics")
		defer span.End()
	} else {
		spanCtx = ctx
	}

	updateStart := time.Now()

	fc.metrics.VolumeSizeGauge.With(prometheus.Labels{
		"device":      filesystem.Device,
		"mount_point": filesystem.MountPoint,
		"volume":      filesystem.Name,
	}).Set(float64(sizeBytes))
	fc.metrics.VolumeAvailableGauge.With(prometheus.Labels{
		"device":      filesystem.Device,
		"mount_point": filesystem.MountPoint,
		"volume":      filesystem.Name,
	}).Set(float64(availableBytes))
	fc.metrics.VolumeUsedRatioGauge.With(prometheus.Labels{
		"device":      filesystem.Device,
		"mount_point": filesystem.MountPoint,
		"volume":      filesystem.Name,
	}).Set(usedRatio)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("metrics.update_duration_seconds", time.Since(updateStart).Seconds()),
			attribute.Int("metrics.count", 3),
		)
		span.AddEvent("metrics_updated")
	}
}

// parseDfOutput parses the df command output with tracing
func (fc *FilesystemCollector) parseDfOutput(ctx context.Context, output []byte, parentSpan *tracing.CollectorSpan) (sizeKB, availableKB int64, err error) {
	tracer := fc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context
	)

	if tracer != nil && tracer.IsEnabled() && parentSpan != nil {
		spanCtx = parentSpan.Context()
		span = tracer.NewCollectorSpan(spanCtx, "filesystem-collector", "parse-df-output")

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
	} else {
		spanCtx = ctx
	}

	parseStart := time.Now()

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		if span != nil {
			span.RecordError(fmt.Errorf("unexpected df output format"), attribute.Int("lines_count", len(lines)))
		}

		return 0, 0, fmt.Errorf("unexpected df output format")
	}

	if span != nil {
		span.SetAttributes(
			attribute.Int("parse.lines_count", len(lines)),
		)
	}

	// Debug: log the output
	slog.Debug("df output", "lines", lines)

	// Parse the header and data lines
	// Format: Filesystem 1K-blocks Used Available Use% Mounted on
	//         /dev/mapper/cachedev_0
	//         16847009220 14430849176 2416160044  86% /mnt/data
	// OR:     /dev/sdb1          239313084  27770148 211424152  12% /mnt/backup
	// The filesystem name might be on a separate line, so we need to find the line with the stats
	var statsLine string

	format := "unknown"

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		// Check if this line contains the stats (should have 5-6 fields)
		parts := strings.Fields(line)
		if len(parts) >= 4 {
			// Try to parse the first field as a number (size in KB)
			if _, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				statsLine = line
				format = "multi-line"

				break
			}
		}
	}

	if statsLine == "" {
		// If we couldn't find a line starting with a number, try the last non-empty line
		for i := len(lines) - 1; i >= 1; i-- {
			line := strings.TrimSpace(lines[i])
			if line != "" {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					// Check if this line has a number in the right position
					// For single-line format: /dev/sdb1 239313084 27770148 211424152 12% /mnt/backup
					// For multi-line format: 16847009220 14430849176 2416160044 86% /mnt/data
					if len(parts) >= 5 {
						// Try to parse the second field as a number (size in KB)
						if _, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
							statsLine = line
							format = "single-line"

							break
						}
					}
				}
			}
		}
	}

	if statsLine == "" {
		if span != nil {
			span.RecordError(fmt.Errorf("could not find stats line"), attribute.Int("lines_searched", len(lines)-1))
		}

		return 0, 0, fmt.Errorf("could not find stats line in df output")
	}

	if span != nil {
		span.SetAttributes(
			attribute.String("parse.format", format),
			attribute.String("parse.stats_line", statsLine),
		)
	}

	parts := strings.Fields(statsLine)
	if len(parts) < 4 {
		if span != nil {
			span.RecordError(fmt.Errorf("unexpected df output format"), attribute.Int("parts_count", len(parts)))
		}

		return 0, 0, fmt.Errorf("unexpected df output format")
	}

	// Parse size and available space (in 1K blocks, convert to bytes)
	// Handle both formats:
	// Single-line: /dev/usb1p1 239313084 27770148 211424152 12% /volumeUSB1/usbshare
	// Multi-line:  16847009220 14430849176 2416160044 86% /mnt/data
	var parseErr error

	// Try to parse the first field as a number
	if sizeKB, parseErr = strconv.ParseInt(parts[0], 10, 64); parseErr == nil {
		// Multi-line format: size is in parts[0], available in parts[2]
		if len(parts) >= 3 {
			availableKB, parseErr = strconv.ParseInt(parts[2], 10, 64)
		} else {
			parseErr = fmt.Errorf("insufficient fields in multi-line format, expected at least 3, got %d", len(parts))
		}
	} else {
		// Single-line format: size is in parts[1], available in parts[3]
		if len(parts) >= 4 {
			sizeKB, parseErr = strconv.ParseInt(parts[1], 10, 64)
			if parseErr != nil {
				parseErr = fmt.Errorf("failed to parse size: %w", parseErr)
			} else {
				availableKB, parseErr = strconv.ParseInt(parts[3], 10, 64)
				if parseErr != nil {
					parseErr = fmt.Errorf("failed to parse available space in single-line format: %w", parseErr)
				}
			}
		} else {
			parseErr = fmt.Errorf("insufficient fields in single-line format, expected at least 4, got %d", len(parts))
		}
	}

	if parseErr != nil {
		if span != nil {
			span.RecordError(parseErr, attribute.String("parse.format", format))
		}

		return 0, 0, parseErr
	}

	if span != nil {
		span.SetAttributes(
			attribute.Int64("parse.size_kb", sizeKB),
			attribute.Int64("parse.available_kb", availableKB),
			attribute.Float64("parse.duration_seconds", time.Since(parseStart).Seconds()),
		)
		span.AddEvent("parse_completed")
	}

	return sizeKB, availableKB, nil
}

// sanitizePath removes any potentially dangerous characters for command execution
func (fc *FilesystemCollector) sanitizePath(path string) string {
	// For now, just return the path as-is since we've already validated it
	// In a more robust implementation, you might want to do additional sanitization
	return path
}
