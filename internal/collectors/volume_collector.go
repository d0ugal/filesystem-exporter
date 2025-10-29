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

	// Retry with exponential backoff
	err := fc.retryWithBackoff(func() error {
		return fc.collectFilesystemUsage(ctx, filesystem)
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
func (fc *FilesystemCollector) retryWithBackoff(operation func() error, maxRetries int, initialDelay time.Duration) error {
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

func (fc *FilesystemCollector) collectFilesystemUsage(ctx context.Context, filesystem config.FilesystemConfig) error {
	// Validate mount point
	if err := fc.validateMountPoint(filesystem.MountPoint); err != nil {
		return fmt.Errorf("mount point validation failed for %s: %w", filesystem.MountPoint, err)
	}

	// Sanitize mount point for command execution
	sanitizedMountPoint := fc.sanitizePath(filesystem.MountPoint)

	// Use df command to get filesystem information
	// Create context with timeout (10 seconds max) - use background context to avoid cancellation propagation
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// G204: Subprocess launched with variable - This is safe because sanitizedMountPoint is validated and sanitized
	//nolint:contextcheck // Intentionally use background context to prevent cancellation propagation
	cmd := exec.CommandContext(timeoutCtx, "df", sanitizedMountPoint)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute df command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("unexpected df output format")
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
							break
						}
					}
				}
			}
		}
	}

	if statsLine == "" {
		return fmt.Errorf("could not find stats line in df output")
	}

	parts := strings.Fields(statsLine)
	if len(parts) < 4 {
		return fmt.Errorf("unexpected df output format")
	}

	// Parse size and available space (in 1K blocks, convert to bytes)
	// Handle both formats:
	// Single-line: /dev/usb1p1 239313084 27770148 211424152 12% /volumeUSB1/usbshare
	// Multi-line:  16847009220 14430849176 2416160044 86% /mnt/data
	var (
		sizeKB, availableKB int64
		parseErr            error
	)

	// Try to parse the first field as a number

	if sizeKB, parseErr = strconv.ParseInt(parts[0], 10, 64); parseErr == nil {
		// Multi-line format: size is in parts[0], available in parts[2]
		if len(parts) >= 3 {
			availableKB, parseErr = strconv.ParseInt(parts[2], 10, 64)
		} else {
			return fmt.Errorf("insufficient fields in multi-line format, expected at least 3, got %d", len(parts))
		}
	} else {
		// Single-line format: size is in parts[1], available in parts[3]
		if len(parts) >= 4 {
			sizeKB, parseErr = strconv.ParseInt(parts[1], 10, 64)
			if parseErr != nil {
				return fmt.Errorf("failed to parse size: %w", parseErr)
			}

			availableKB, parseErr = strconv.ParseInt(parts[3], 10, 64)
			if parseErr != nil {
				return fmt.Errorf("failed to parse available space in single-line format: %w", parseErr)
			}
		} else {
			return fmt.Errorf("insufficient fields in single-line format, expected at least 4, got %d", len(parts))
		}
	}

	if parseErr != nil {
		return fmt.Errorf("failed to parse available space: %w", parseErr)
	}

	// Convert KB to bytes
	sizeBytes := sizeKB * 1024
	availableBytes := availableKB * 1024
	usedBytes := sizeBytes - availableBytes
	usedRatio := float64(usedBytes) / float64(sizeBytes)

	// Update metrics
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
func (fc *FilesystemCollector) validateMountPoint(mountPoint string) error {
	// Check if mount point exists
	if _, err := os.Stat(mountPoint); err != nil {
		return fmt.Errorf("mount point does not exist: %s", mountPoint)
	}

	// Ensure mount point is absolute
	if !filepath.IsAbs(mountPoint) {
		return fmt.Errorf("mount point must be absolute: %s", mountPoint)
	}

	// Basic sanitization - check for dangerous patterns
	dangerousPatterns := []string{"..", "~", "*", "?", "["}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(mountPoint, pattern) {
			return fmt.Errorf("mount point contains dangerous pattern '%s': %s", pattern, mountPoint)
		}
	}

	return nil
}

// sanitizePath removes any potentially dangerous characters for command execution
func (fc *FilesystemCollector) sanitizePath(path string) string {
	// For now, just return the path as-is since we've already validated it
	// In a more robust implementation, you might want to do additional sanitization
	return path
}
