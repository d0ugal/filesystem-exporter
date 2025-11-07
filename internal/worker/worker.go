package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
	"filesystem-exporter/internal/queue"
	"filesystem-exporter/internal/state"
	"filesystem-exporter/internal/utils"
	"github.com/d0ugal/promexporter/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Worker processes jobs from a queue
type Worker struct {
	queue     *queue.Queue
	metrics   *metrics.FilesystemRegistry
	state     *state.Tracker
	config    *config.Config
	tracer    *tracing.Tracer
	queueType string // "filesystem" or "directory"
}

// NewWorker creates a new worker
func NewWorker(q *queue.Queue, m *metrics.FilesystemRegistry, s *state.Tracker, cfg *config.Config, tracer *tracing.Tracer, queueType string) *Worker {
	return &Worker{
		queue:     q,
		metrics:   m,
		state:     s,
		config:    cfg,
		tracer:    tracer,
		queueType: queueType,
	}
}

// Start starts the worker goroutine
func (w *Worker) Start(ctx context.Context) {
	ctx, span := w.startSpan(ctx, "worker.start", trace.WithAttributes(
		attribute.String("worker.queue_type", w.queueType),
	))
	defer span.End()

	go w.run(ctx)

	span.AddEvent("worker_started")
}

// run is the main worker loop
func (w *Worker) run(ctx context.Context) {
	slog.Info("Worker started", "queue_type", w.queueType)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker stopping", "queue_type", w.queueType)
			return
		case job := <-w.queue.Channel():
			w.processJob(ctx, job)
		}
	}
}

// processJob processes a single job
func (w *Worker) processJob(_ context.Context, job queue.Job) {
	// Use job context which has the trace span
	ctx := job.Context

	ctx, span := w.startSpan(ctx, "worker.process_job", trace.WithAttributes(
		attribute.String("worker.queue_type", w.queueType),
		attribute.String("job.id", job.ID),
		attribute.String("job.type", job.Type),
		attribute.String("job.name", job.Name),
		attribute.String("job.path", job.Path),
		attribute.Float64("job.timeout_seconds", job.Timeout.Seconds()),
	))
	defer span.End()

	startTime := time.Now()

	// Track memory usage
	var memStart, memEnd runtime.MemStats
	runtime.ReadMemStats(&memStart)

	// Set running job state
	jobState := &state.JobState{
		ID:        job.ID,
		Type:      job.Type,
		Name:      job.Name,
		Path:      job.Path,
		StartedAt: startTime,
		//nolint:contextcheck // Context is from job, not inherited
		TraceID: trace.SpanFromContext(ctx).SpanContext().TraceID().String(),
	}
	//nolint:contextcheck // Context is from job, not inherited
	w.state.SetRunningJob(ctx, w.queueType, jobState)

	slog.Info("Processing job",
		"queue_type", w.queueType,
		"job_id", job.ID,
		"job_name", job.Name,
		"job_path", job.Path,
		"trace_id", jobState.TraceID,
	)

	var err error

	switch job.Type {
	case "filesystem":
		//nolint:contextcheck // Context is from job, not inherited
		err = w.processFilesystem(ctx, job)
	case "directory":
		//nolint:contextcheck // Context is from job, not inherited
		err = w.processDirectory(ctx, job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	duration := time.Since(startTime)

	runtime.ReadMemStats(&memEnd)

	// Calculate resource usage
	// Note: CPU time measurement requires process stats which we'll get from command execution
	// For now, we'll track it per command execution
	cpuUserSecs := 0.0
	cpuSystemSecs := 0.0
	// memAllocated can be negative if GC runs during job execution, so clamp to 0
	var memAllocatedDelta int64
	if memEnd.Alloc > memStart.Alloc {
		//nolint:gosec // Safe: uint64 subtraction with bounds checking
		memAllocatedDelta = int64(memEnd.Alloc - memStart.Alloc)
	}

	memAllocated := memAllocatedDelta
	if memAllocated < 0 {
		memAllocated = 0
	}
	// TotalAlloc is cumulative and always increases, so this should be positive
	var memPeak int64
	if memEnd.TotalAlloc > memStart.TotalAlloc {
		//nolint:gosec // Safe: uint64 subtraction with bounds checking
		memPeak = int64(memEnd.TotalAlloc - memStart.TotalAlloc)
	}

	if memPeak < 0 {
		memPeak = 0
	}

	// Clear running job state
	//nolint:contextcheck // Context is from job, not inherited
	w.state.ClearRunningJob(ctx, w.queueType, job.ID, duration)

	// Note: Scheduler will clear its running flag when it detects the job is complete
	// via the state tracker check

	// Update resource metrics
	//nolint:contextcheck // Context is from job, not inherited
	w.updateResourceMetrics(ctx, job, duration, cpuUserSecs, cpuSystemSecs, memAllocated, memPeak)

	// Update span attributes
	span.SetAttributes(
		attribute.Float64("job.duration_seconds", duration.Seconds()),
		attribute.Float64("job.cpu_user_seconds", cpuUserSecs),
		attribute.Float64("job.cpu_system_seconds", cpuSystemSecs),
		attribute.Int64("job.memory_allocated_bytes", memAllocated),
		attribute.Int64("job.memory_peak_bytes", memPeak),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		w.metrics.CollectionFailedCounter.WithLabelValues(
			job.Name,
			strconv.Itoa(int(job.Interval.Seconds())),
			job.Type,
		).Inc()

		slog.Error("Job failed",
			"queue_type", w.queueType,
			"job_id", job.ID,
			"job_name", job.Name,
			"error", err,
			"duration", duration,
			"trace_id", jobState.TraceID,
		)

		return
	}

	span.SetStatus(codes.Ok, "job completed successfully")

	w.metrics.CollectionSuccess.WithLabelValues(
		job.Name,
		strconv.Itoa(int(job.Interval.Seconds())),
		job.Type,
	).Inc()

	w.metrics.CollectionDuration.WithLabelValues(
		job.Name,
		strconv.Itoa(int(job.Interval.Seconds())),
		job.Type,
	).Set(duration.Seconds())

	// Update collection timestamp for alerting compatibility
	w.metrics.CollectionTimestampGauge.WithLabelValues(
		job.Name,
		strconv.Itoa(int(job.Interval.Seconds())),
		job.Type,
	).Set(float64(time.Now().Unix()))

	slog.Info("Job completed",
		"queue_type", w.queueType,
		"job_id", job.ID,
		"job_name", job.Name,
		"duration", duration,
		"cpu_user", cpuUserSecs,
		"cpu_system", cpuSystemSecs,
		"memory_allocated", memAllocated,
		"trace_id", jobState.TraceID,
	)

	// Warning if duration > 50% of interval
	if duration > job.Interval/2 {
		slog.Warn("Job duration exceeds 50% of interval",
			"queue_type", w.queueType,
			"job_name", job.Name,
			"duration", duration,
			"interval", job.Interval,
			"ratio", float64(duration)/float64(job.Interval),
		)
		span.SetAttributes(
			attribute.Bool("job.slow_warning", true),
			attribute.Float64("job.duration_ratio", float64(duration)/float64(job.Interval)),
		)
	}
}

// processFilesystem processes a filesystem collection job
func (w *Worker) processFilesystem(ctx context.Context, job queue.Job) error {
	ctx, span := w.startSpan(ctx, "filesystem.collect", trace.WithAttributes(
		attribute.String("filesystem.name", job.Name),
		attribute.String("filesystem.mount_point", job.Path),
	))
	defer span.End()

	// Find filesystem config
	var fsConfig *config.FilesystemConfig

	for _, fs := range w.config.Filesystems {
		if fs.Name == job.Name {
			fsConfig = &fs
			break
		}
	}

	if fsConfig == nil {
		err := fmt.Errorf("filesystem config not found: %s", job.Name)
		span.RecordError(err)

		return err
	}

	// Execute df command
	output, err := w.executeDfCommand(ctx, job.Path)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("df command failed: %w", err)
	}

	// Parse df output
	sizeKB, availableKB, err := w.parseDfOutput(ctx, output)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("parse df output failed: %w", err)
	}

	// Convert to bytes
	sizeBytes := sizeKB * 1024
	availableBytes := availableKB * 1024
	usedBytes := sizeBytes - availableBytes
	usedRatio := float64(usedBytes) / float64(sizeBytes)

	// Update metrics
	w.updateFilesystemMetrics(ctx, fsConfig, sizeBytes, availableBytes, usedRatio)

	span.SetAttributes(
		attribute.Int64("filesystem.size_bytes", sizeBytes),
		attribute.Int64("filesystem.available_bytes", availableBytes),
		attribute.Int64("filesystem.used_bytes", usedBytes),
		attribute.Float64("filesystem.used_ratio", usedRatio),
	)
	span.AddEvent("filesystem_collected")

	return nil
}

// processDirectory processes a directory collection job
func (w *Worker) processDirectory(ctx context.Context, job queue.Job) error {
	ctx, span := w.startSpan(ctx, "directory.collect", trace.WithAttributes(
		attribute.String("directory.name", job.Name),
		attribute.String("directory.path", job.Path),
	))
	defer span.End()

	// Find directory config
	dirConfig, exists := w.config.Directories[job.Name]
	if !exists {
		err := fmt.Errorf("directory config not found: %s", job.Name)
		span.RecordError(err)

		return err
	}

	span.SetAttributes(
		attribute.Int("directory.subdirectory_levels", dirConfig.SubdirectoryLevels),
	)

	// Validate path
	if err := w.validatePath(ctx, job.Path); err != nil {
		span.RecordError(err)
		return err
	}

	// Default subdirectory_levels to 0 if not specified
	subdirectoryLevels := dirConfig.SubdirectoryLevels
	if subdirectoryLevels < 0 {
		subdirectoryLevels = 0
	}

	// Collect directory and subdirectories based on subdirectory_levels
	if subdirectoryLevels == 0 {
		// Just collect the directory itself
		sizeKB, err := w.executeDuCommand(ctx, job.Path, job.Timeout)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("du command failed: %w", err)
		}

		// Convert to bytes
		sizeBytes := sizeKB * 1024

		// Update metrics
		w.updateDirectoryMetrics(ctx, job.Name, job.Path, sizeBytes, 0)

		span.SetAttributes(
			attribute.Int64("directory.size_bytes", sizeBytes),
			attribute.Int64("directory.size_kb", sizeKB),
		)
	} else {
		// Collect directory and all subdirectories up to specified depth
		subdirSizes, err := w.executeDuCommandWithDepth(ctx, job.Path, subdirectoryLevels, job.Timeout)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("du command with depth failed: %w", err)
		}

		// Update metrics for each subdirectory found
		for path, sizeKB := range subdirSizes {
			sizeBytes := sizeKB * 1024
			// Calculate subdirectory level (depth from base path)
			level := w.calculateSubdirectoryLevel(job.Path, path)
			w.updateDirectoryMetrics(ctx, job.Name, path, sizeBytes, level)
		}

		span.SetAttributes(
			attribute.Int("directory.subdirectories_collected", len(subdirSizes)),
		)
	}

	span.AddEvent("directory_collected")

	return nil
}

// executeDfCommand executes the df command
func (w *Worker) executeDfCommand(ctx context.Context, mountPoint string) ([]byte, error) {
	ctx, span := w.startSpan(ctx, "command.df", trace.WithAttributes(
		attribute.String("command.mount_point", mountPoint),
	))
	defer span.End()

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "df", mountPoint)

	execStart := time.Now()
	output, err := cmd.Output()
	execDuration := time.Since(execStart)

	span.SetAttributes(
		attribute.Float64("command.duration_seconds", execDuration.Seconds()),
		attribute.Int("command.output_size_bytes", len(output)),
	)

	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			span.SetAttributes(attribute.String("command.error_type", "timeout"))
			slog.Error("df command timed out", "mount_point", mountPoint, "duration", execDuration)
		}

		span.RecordError(err)

		return nil, err
	}

	span.AddEvent("command_completed")

	return output, nil
}

// executeDuCommand executes the du command
func (w *Worker) executeDuCommand(ctx context.Context, path string, timeout time.Duration) (int64, error) {
	ctx, span := w.startSpan(ctx, "command.du", trace.WithAttributes(
		attribute.String("command.path", path),
		attribute.Float64("command.timeout_seconds", timeout.Seconds()),
	))
	defer span.End()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "du", "-s", "-x", path)

	// Set I/O priority
	if err := utils.SetupCommandWithIOPriority(cmd); err != nil {
		slog.Debug("Failed to set I/O priority", "error", err)
	}

	execStart := time.Now()
	output, err := cmd.Output()
	execDuration := time.Since(execStart)

	span.SetAttributes(
		attribute.Float64("command.duration_seconds", execDuration.Seconds()),
		attribute.Int("command.output_size_bytes", len(output)),
	)

	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			span.SetAttributes(attribute.String("command.error_type", "timeout"))
			slog.Error("du command timed out", "path", path, "duration", execDuration, "timeout", timeout)
		}

		span.RecordError(err)

		return 0, err
	}

	// Parse output
	sizeKB, err := w.parseDuOutput(ctx, output)
	if err != nil {
		span.RecordError(err)
		return 0, err
	}

	span.AddEvent("command_completed")

	return sizeKB, nil
}

// executeDuCommandWithDepth executes du with --max-depth to collect subdirectories
// Returns a map of path -> size in KB
func (w *Worker) executeDuCommandWithDepth(ctx context.Context, path string, maxDepth int, timeout time.Duration) (map[string]int64, error) {
	ctx, span := w.startSpan(ctx, "command.du_depth", trace.WithAttributes(
		attribute.String("command.path", path),
		attribute.Int("command.max_depth", maxDepth),
		attribute.Float64("command.timeout_seconds", timeout.Seconds()),
	))
	defer span.End()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Use du with --max-depth to get all subdirectories in one pass
	// -x: don't cross filesystem boundaries
	// --max-depth: maximum depth to traverse (0 = base dir only, 1 = base + direct subdirs, etc.)
	// Note: We don't use -s (summarize) here because it conflicts with --max-depth
	cmd := exec.CommandContext(timeoutCtx, "du", "-x", fmt.Sprintf("--max-depth=%d", maxDepth), path)

	// Set I/O priority
	if err := utils.SetupCommandWithIOPriority(cmd); err != nil {
		slog.Debug("Failed to set I/O priority", "error", err)
	}

	execStart := time.Now()
	output, err := cmd.Output()
	execDuration := time.Since(execStart)

	span.SetAttributes(
		attribute.Float64("command.duration_seconds", execDuration.Seconds()),
		attribute.Int("command.output_size_bytes", len(output)),
	)

	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			span.SetAttributes(attribute.String("command.error_type", "timeout"))
			slog.Error("du command with depth timed out", "path", path, "max_depth", maxDepth, "duration", execDuration, "timeout", timeout)
		}

		span.RecordError(err)

		return nil, err
	}

	// Parse output to extract all subdirectory sizes
	subdirSizes, err := w.parseDuOutputWithDepth(ctx, output, path)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("command.subdirectories_found", len(subdirSizes)),
	)
	span.AddEvent("command_completed")

	return subdirSizes, nil
}

// parseDuOutputWithDepth parses du output with depth information
// du --max-depth outputs lines like: "1024\t/path/to/dir"
// Returns a map of path -> size in KB
func (w *Worker) parseDuOutputWithDepth(ctx context.Context, output []byte, basePath string) (map[string]int64, error) {
	_, span := w.startSpan(ctx, "parse.du_output_depth", trace.WithAttributes(
		attribute.Int("output.size_bytes", len(output)),
	))
	defer span.End()

	result := make(map[string]int64)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// du output format: "SIZE\tPATH"
		// Split on tab (du uses tab separator)
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			span.SetAttributes(
				attribute.String("parse.error", "invalid line format"),
				attribute.String("parse.line", line),
			)

			continue
		}

		sizeStr := strings.TrimSpace(parts[0])
		dirPath := strings.TrimSpace(parts[1])

		// Parse size (in KB)
		sizeKB, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			span.SetAttributes(
				attribute.String("parse.error", "invalid size"),
				attribute.String("parse.size", sizeStr),
			)

			continue
		}

		// Normalize path (remove trailing slashes for consistency)
		dirPath = strings.TrimRight(dirPath, "/")
		result[dirPath] = sizeKB
	}

	span.SetAttributes(
		attribute.Int("parse.directories_parsed", len(result)),
	)
	span.AddEvent("parse_completed")

	return result, nil
}

// calculateSubdirectoryLevel calculates the depth level of a subdirectory relative to the base path
// Returns 0 for the base directory itself, 1 for direct subdirectories, etc.
func (w *Worker) calculateSubdirectoryLevel(basePath, subdirPath string) int {
	// Normalize paths
	basePath = strings.TrimRight(filepath.Clean(basePath), "/")
	subdirPath = strings.TrimRight(filepath.Clean(subdirPath), "/")

	// If paths are the same, it's level 0 (the base directory)
	if basePath == subdirPath {
		return 0
	}

	// Count path separators in the relative path
	relPath, err := filepath.Rel(basePath, subdirPath)
	if err != nil {
		// If we can't calculate relative path, assume it's a direct subdirectory
		return 1
	}

	// Count the number of path separators to determine depth
	level := strings.Count(relPath, string(filepath.Separator))
	if level == 0 {
		// Same directory or immediate subdirectory
		return 1
	}

	return level + 1
}

// parseDfOutput parses df command output
func (w *Worker) parseDfOutput(ctx context.Context, output []byte) (sizeKB, availableKB int64, err error) {
	_, span := w.startSpan(ctx, "parse.df_output", trace.WithAttributes(
		attribute.Int("output.size_bytes", len(output)),
	))
	defer span.End()

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		err := fmt.Errorf("unexpected df output format: %d lines", len(lines))
		span.RecordError(err)

		return 0, 0, err
	}

	// Find the stats line (second line or later that has numeric values)
	var statsLine string

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 4 {
			// Try to parse first field as number
			if _, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				statsLine = line
				break
			}
		}
	}

	if statsLine == "" {
		// Try last non-empty line
		for i := len(lines) - 1; i >= 1; i-- {
			line := strings.TrimSpace(lines[i])
			if line != "" {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					if _, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
						statsLine = line
						break
					}
				}
			}
		}
	}

	if statsLine == "" {
		err := fmt.Errorf("could not find stats line in df output")
		span.RecordError(err)

		return 0, 0, err
	}

	parts := strings.Fields(statsLine)
	if len(parts) < 4 {
		err := fmt.Errorf("unexpected df output format: %d fields", len(parts))
		span.RecordError(err)

		return 0, 0, err
	}

	// Parse size and available
	if sizeKB, err = strconv.ParseInt(parts[0], 10, 64); err == nil {
		// Multi-line format
		if len(parts) >= 3 {
			availableKB, err = strconv.ParseInt(parts[2], 10, 64)
		}
	} else {
		// Single-line format
		if len(parts) >= 4 {
			sizeKB, err = strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				availableKB, err = strconv.ParseInt(parts[3], 10, 64)
			}
		}
	}

	if err != nil {
		span.RecordError(err)
		return 0, 0, fmt.Errorf("failed to parse df output: %w", err)
	}

	span.SetAttributes(
		attribute.Int64("parse.size_kb", sizeKB),
		attribute.Int64("parse.available_kb", availableKB),
	)

	return sizeKB, availableKB, nil
}

// parseDuOutput parses du command output
func (w *Worker) parseDuOutput(ctx context.Context, output []byte) (int64, error) {
	_, span := w.startSpan(ctx, "parse.du_output", trace.WithAttributes(
		attribute.Int("output.size_bytes", len(output)),
	))
	defer span.End()

	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		err := fmt.Errorf("unexpected du output format")
		span.RecordError(err)

		return 0, err
	}

	sizeKB, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to parse directory size: %w", err)
	}

	span.SetAttributes(attribute.Int64("parse.size_kb", sizeKB))

	return sizeKB, nil
}

// validatePath validates a path
func (w *Worker) validatePath(ctx context.Context, path string) error {
	_, span := w.startSpan(ctx, "validate.path", trace.WithAttributes(
		attribute.String("path", path),
	))
	defer span.End()

	if _, err := os.Stat(path); err != nil {
		span.RecordError(err)
		return fmt.Errorf("path does not exist: %s", path)
	}

	if !filepath.IsAbs(path) {
		err := fmt.Errorf("path must be absolute: %s", path)
		span.RecordError(err)

		return err
	}

	// Basic sanitization
	dangerousPatterns := []string{"..", "~", "*", "?", "["}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(path, pattern) {
			err := fmt.Errorf("path contains dangerous pattern '%s': %s", pattern, path)
			span.RecordError(err)

			return err
		}
	}

	span.AddEvent("validation_completed")

	return nil
}

// updateFilesystemMetrics updates filesystem metrics
func (w *Worker) updateFilesystemMetrics(ctx context.Context, fs *config.FilesystemConfig, sizeBytes, availableBytes int64, usedRatio float64) {
	_, span := w.startSpan(ctx, "worker.update_metrics", trace.WithAttributes(
		attribute.String("metric.type", "filesystem"),
	))
	defer span.End()

	w.metrics.VolumeSizeGauge.WithLabelValues(
		fs.Device,
		fs.MountPoint,
		fs.Name,
	).Set(float64(sizeBytes))

	w.metrics.VolumeAvailableGauge.WithLabelValues(
		fs.Device,
		fs.MountPoint,
		fs.Name,
	).Set(float64(availableBytes))

	w.metrics.VolumeUsedRatioGauge.WithLabelValues(
		fs.Device,
		fs.MountPoint,
		fs.Name,
	).Set(usedRatio)

	span.AddEvent("metrics_updated")
}

// updateDirectoryMetrics updates directory metrics
func (w *Worker) updateDirectoryMetrics(ctx context.Context, groupName, path string, sizeBytes int64, subdirectoryLevel int) {
	_, span := w.startSpan(ctx, "worker.update_metrics", trace.WithAttributes(
		attribute.String("metric.type", "directory"),
	))
	defer span.End()

	w.metrics.DirectorySizeGauge.WithLabelValues(
		groupName,
		path,
		"du",
		strconv.Itoa(subdirectoryLevel),
	).Set(float64(sizeBytes))

	w.metrics.DirectoriesProcessedCounter.WithLabelValues(
		groupName,
		"du",
	).Inc()

	// Update du_lock_wait_duration_seconds for alerting compatibility
	// Set to 0 since new architecture uses separate queues (no global lock contention)
	w.metrics.DuLockWaitDurationGauge.WithLabelValues(
		groupName,
		path,
	).Set(0)

	span.AddEvent("metrics_updated")
}

// updateResourceMetrics updates resource usage metrics
func (w *Worker) updateResourceMetrics(ctx context.Context, job queue.Job, duration time.Duration, cpuUser, cpuSystem float64, memAllocated, memPeak int64) {
	_, span := w.startSpan(ctx, "worker.update_resource_metrics")
	defer span.End()

	// Update per-job metrics
	w.metrics.JobCPUUserSeconds.WithLabelValues(
		job.Type,
		job.Name,
	).Set(cpuUser)

	w.metrics.JobCPUSystemSeconds.WithLabelValues(
		job.Type,
		job.Name,
	).Set(cpuSystem)

	w.metrics.JobMemoryAllocatedBytes.WithLabelValues(
		job.Type,
		job.Name,
	).Set(float64(memAllocated))

	w.metrics.JobMemoryPeakBytes.WithLabelValues(
		job.Type,
		job.Name,
	).Set(float64(memPeak))

	// Update process-level metrics (accumulative)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	w.metrics.ProcessMemoryAllocBytes.Set(float64(memStats.Alloc))
	w.metrics.ProcessMemorySysBytes.Set(float64(memStats.Sys))
	w.metrics.ProcessNumGCTotal.Set(float64(memStats.NumGC))

	span.AddEvent("resource_metrics_updated")
}

// startSpan is a helper to start an OTEL span
func (w *Worker) startSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if w.tracer != nil && w.tracer.IsEnabled() {
		return w.tracer.StartSpan(ctx, name, opts...)
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ctx, span
	}

	return ctx, span
}
