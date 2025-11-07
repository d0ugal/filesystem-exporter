package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
	"filesystem-exporter/internal/queue"
	"filesystem-exporter/internal/state"
	"github.com/d0ugal/promexporter/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Scheduler manages scheduling of collection jobs
type Scheduler struct {
	config  *config.Config
	metrics *metrics.FilesystemRegistry
	state   *state.Tracker

	// Queues
	filesystemQueue *queue.Queue
	directoryQueue  *queue.Queue

	// Tickers for filesystems
	filesystemTickers map[string]*time.Ticker
	filesystemMutex   sync.RWMutex

	// Tickers for directories
	directoryTickers map[string]*time.Ticker
	directoryMutex   sync.RWMutex

	// Track running jobs per item
	filesystemRunning map[string]bool
	directoryRunning  map[string]bool
	runningMutex      sync.RWMutex

	tracer             trace.Tracer
	promexporterTracer *tracing.Tracer
}

// NewScheduler creates a new scheduler
func NewScheduler(
	cfg *config.Config,
	m *metrics.FilesystemRegistry,
	s *state.Tracker,
	fsQueue *queue.Queue,
	dirQueue *queue.Queue,
	tracer *tracing.Tracer,
) *Scheduler {
	var otelTracer trace.Tracer
	if tracer != nil && tracer.IsEnabled() {
		// Use the tracer's StartSpan method which returns otel spans
		// We'll use a wrapper approach - store the promexporter tracer
		// and use it in startSpan
		otelTracer = nil // Will use tracer.StartSpan directly
	}

	return &Scheduler{
		config:             cfg,
		metrics:            m,
		state:              s,
		filesystemQueue:    fsQueue,
		directoryQueue:     dirQueue,
		filesystemTickers:  make(map[string]*time.Ticker),
		directoryTickers:   make(map[string]*time.Ticker),
		filesystemRunning:  make(map[string]bool),
		directoryRunning:   make(map[string]bool),
		tracer:             otelTracer,
		promexporterTracer: tracer,
	}
}

// Start initializes the scheduler
func (s *Scheduler) Start(ctx context.Context) {
	ctx, span := s.startSpan(ctx, "scheduler.init", trace.WithAttributes(
		attribute.Int("filesystem_count", len(s.config.Filesystems)),
		attribute.Int("directory_count", len(s.config.Directories)),
	))
	defer span.End()

	slog.Info("Initializing scheduler",
		"filesystems", len(s.config.Filesystems),
		"directories", len(s.config.Directories),
	)

	// Register items in state tracker
	for _, fs := range s.config.Filesystems {
		s.state.RegisterItem(ctx, "filesystem", fs.Name)

		// Set timeout metric
		timeout := s.config.GetFilesystemTimeout(fs)
		s.metrics.CollectionTimeoutSeconds.With(prometheus.Labels{
			"item_name": fs.Name,
			"item_type": "filesystem",
		}).Set(timeout.Seconds())

		// Set interval metric
		interval := s.config.GetFilesystemInterval(fs)
		s.metrics.CollectionIntervalGauge.With(prometheus.Labels{
			"group": fs.Name,
			"type":  "filesystem",
		}).Set(float64(interval))
	}

	for name, dir := range s.config.Directories {
		s.state.RegisterItem(ctx, "directory", name)

		// Set timeout metric
		timeout := s.config.GetDirectoryTimeout(dir)
		s.metrics.CollectionTimeoutSeconds.With(prometheus.Labels{
			"item_name": name,
			"item_type": "directory",
		}).Set(timeout.Seconds())

		// Set interval metric
		interval := s.config.GetDirectoryInterval(dir)
		s.metrics.CollectionIntervalGauge.With(prometheus.Labels{
			"group": name,
			"type":  "directory",
		}).Set(float64(interval))
	}

	// Start filesystem tickers
	for _, fs := range s.config.Filesystems {
		s.startFilesystemTicker(ctx, fs)
	}

	// Start directory tickers
	for name, dir := range s.config.Directories {
		s.startDirectoryTicker(ctx, name, dir)
	}

	span.AddEvent("scheduler_initialized")
	slog.Info("Scheduler initialized")
}

// startFilesystemTicker starts a ticker for a filesystem
func (s *Scheduler) startFilesystemTicker(ctx context.Context, fs config.FilesystemConfig) {
	ctx, span := s.startSpan(ctx, "scheduler.start_filesystem_ticker", trace.WithAttributes(
		attribute.String("filesystem.name", fs.Name),
	))
	defer span.End()

	interval := s.config.GetFilesystemInterval(fs)
	intervalDuration := time.Duration(interval) * time.Second
	timeout := s.config.GetFilesystemTimeout(fs)

	// Validate interval vs timeout
	if intervalDuration < timeout {
		slog.Warn("Filesystem interval is less than timeout",
			"filesystem", fs.Name,
			"interval", intervalDuration,
			"timeout", timeout,
		)
		span.SetAttributes(
			attribute.Bool("scheduler.warning", true),
			attribute.String("scheduler.warning_reason", "interval_less_than_timeout"),
		)
	}

	ticker := time.NewTicker(intervalDuration)

	s.filesystemMutex.Lock()
	s.filesystemTickers[fs.Name] = ticker
	s.filesystemMutex.Unlock()

	// Initial collection - create a root span for it
	spanCtx := context.WithoutCancel(ctx)
	initCtx, initSpan := s.startSpan(spanCtx, "collection.cycle", trace.WithAttributes(
		attribute.String("item.type", "filesystem"),
		attribute.String("item.name", fs.Name),
		attribute.Float64("interval_seconds", intervalDuration.Seconds()),
		attribute.Bool("initial", true),
	))
	s.scheduleFilesystem(initCtx, fs, timeout, intervalDuration)
	// End the cycle span when the job completes (async)
	go s.waitForJobCompletionAndEndSpan(initCtx, initSpan, "filesystem", fs.Name, timeout)

	// Start goroutine for ticker
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Create a new root span for each collection cycle
				cycleCtx := context.WithoutCancel(ctx)
				cycleCtx, cycleSpan := s.startSpan(cycleCtx, "collection.cycle", trace.WithAttributes(
					attribute.String("item.type", "filesystem"),
					attribute.String("item.name", fs.Name),
					attribute.Float64("interval_seconds", intervalDuration.Seconds()),
				))
				s.scheduleFilesystem(cycleCtx, fs, timeout, intervalDuration)
				// End the cycle span when the job completes (async)
				go s.waitForJobCompletionAndEndSpan(cycleCtx, cycleSpan, "filesystem", fs.Name, timeout)
			}
		}
	}()

	span.AddEvent("filesystem_ticker_started")
}

// startDirectoryTicker starts a ticker for a directory
func (s *Scheduler) startDirectoryTicker(ctx context.Context, name string, dir config.DirectoryGroup) {
	ctx, span := s.startSpan(ctx, "scheduler.start_directory_ticker", trace.WithAttributes(
		attribute.String("directory.name", name),
	))
	defer span.End()

	interval := s.config.GetDirectoryInterval(dir)
	intervalDuration := time.Duration(interval) * time.Second
	timeout := s.config.GetDirectoryTimeout(dir)

	// Validate interval vs timeout
	if intervalDuration < timeout {
		slog.Warn("Directory interval is less than timeout",
			"directory", name,
			"interval", intervalDuration,
			"timeout", timeout,
		)
		span.SetAttributes(
			attribute.Bool("scheduler.warning", true),
			attribute.String("scheduler.warning_reason", "interval_less_than_timeout"),
		)
	}

	ticker := time.NewTicker(intervalDuration)

	s.directoryMutex.Lock()
	s.directoryTickers[name] = ticker
	s.directoryMutex.Unlock()

	// Initial collection - create a root span for it
	initCtx := context.WithoutCancel(ctx)
	initCtx, initSpan := s.startSpan(initCtx, "collection.cycle", trace.WithAttributes(
		attribute.String("item.type", "directory"),
		attribute.String("item.name", name),
		attribute.Float64("interval_seconds", intervalDuration.Seconds()),
		attribute.Bool("initial", true),
	))
	s.scheduleDirectory(initCtx, name, dir, timeout, intervalDuration)
	// End the cycle span when the job completes (async)
	go s.waitForJobCompletionAndEndSpan(initCtx, initSpan, "directory", name, timeout)

	// Start goroutine for ticker
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Create a new root span for each collection cycle
				cycleCtx := context.WithoutCancel(ctx)
				cycleCtx, cycleSpan := s.startSpan(cycleCtx, "collection.cycle", trace.WithAttributes(
					attribute.String("item.type", "directory"),
					attribute.String("item.name", name),
					attribute.Float64("interval_seconds", intervalDuration.Seconds()),
				))
				s.scheduleDirectory(cycleCtx, name, dir, timeout, intervalDuration)
				// End the cycle span when the job completes (async)
				go s.waitForJobCompletionAndEndSpan(cycleCtx, cycleSpan, "directory", name, timeout)
			}
		}
	}()

	span.AddEvent("directory_ticker_started")
}

// scheduleFilesystem schedules a filesystem collection job
func (s *Scheduler) scheduleFilesystem(ctx context.Context, fs config.FilesystemConfig, timeout time.Duration, interval time.Duration) {
	ctx, span := s.startSpan(ctx, "scheduler.schedule", trace.WithAttributes(
		attribute.String("item.type", "filesystem"),
		attribute.String("item.name", fs.Name),
	))
	defer span.End()

	// Check if already running
	s.runningMutex.RLock()
	running := s.filesystemRunning[fs.Name]
	s.runningMutex.RUnlock()

	if running {
		// Check state tracker as well
		if s.state.IsRunning(ctx, "filesystem", fs.Name) {
			slog.Warn("Skipping filesystem collection - previous job still running",
				"filesystem", fs.Name,
			)
			s.metrics.CollectionSkippedCounter.With(prometheus.Labels{
				"queue_type": "filesystem",
				"item_name":  fs.Name,
				"reason":     "previous_job_running",
			}).Inc()
			span.SetAttributes(
				attribute.Bool("scheduler.skipped", true),
				attribute.String("scheduler.skip_reason", "previous_job_running"),
			)
			span.AddEvent("job_skipped")

			return
		}
	}

	// Mark as running (will be cleared when job completes)
	s.runningMutex.Lock()
	s.filesystemRunning[fs.Name] = true
	s.runningMutex.Unlock()

	// Clear running flag when job completes (async check)
	go func() {
		// Wait for job to complete by checking state
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		timeoutChan := time.After(timeout + 5*time.Second) // Wait a bit longer than timeout

		for {
			select {
			case <-timeoutChan:
				s.runningMutex.Lock()
				delete(s.filesystemRunning, fs.Name)
				s.runningMutex.Unlock()

				return
			case <-ticker.C:
				if !s.state.IsRunning(ctx, "filesystem", fs.Name) {
					s.runningMutex.Lock()
					delete(s.filesystemRunning, fs.Name)
					s.runningMutex.Unlock()

					return
				}
			}
		}
	}()

	// Create job
	job := queue.Job{
		ID:       fmt.Sprintf("%s-%s-%d", "filesystem", fs.Name, time.Now().Unix()),
		Type:     "filesystem",
		Name:     fs.Name,
		Path:     fs.MountPoint,
		Timeout:  timeout,
		Interval: interval,
		Context:  ctx,
	}

	// Enqueue
	if err := s.filesystemQueue.Enqueue(ctx, job); err != nil {
		slog.Error("Failed to enqueue filesystem job", "filesystem", fs.Name, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.runningMutex.Lock()
		s.filesystemRunning[fs.Name] = false
		s.runningMutex.Unlock()

		return
	}

	span.SetAttributes(
		attribute.String("job.id", job.ID),
		attribute.Float64("job.timeout_seconds", timeout.Seconds()),
		attribute.Float64("job.interval_seconds", interval.Seconds()),
	)
	span.AddEvent("job_scheduled")
}

// scheduleDirectory schedules a directory collection job
func (s *Scheduler) scheduleDirectory(ctx context.Context, name string, dir config.DirectoryGroup, timeout time.Duration, interval time.Duration) {
	ctx, span := s.startSpan(ctx, "scheduler.schedule", trace.WithAttributes(
		attribute.String("item.type", "directory"),
		attribute.String("item.name", name),
	))
	defer span.End()

	// Check if already running
	s.runningMutex.RLock()
	running := s.directoryRunning[name]
	s.runningMutex.RUnlock()

	if running {
		// Check state tracker as well
		if s.state.IsRunning(ctx, "directory", name) {
			slog.Warn("Skipping directory collection - previous job still running",
				"directory", name,
			)
			s.metrics.CollectionSkippedCounter.With(prometheus.Labels{
				"queue_type": "directory",
				"item_name":  name,
				"reason":     "previous_job_running",
			}).Inc()
			span.SetAttributes(
				attribute.Bool("scheduler.skipped", true),
				attribute.String("scheduler.skip_reason", "previous_job_running"),
			)
			span.AddEvent("job_skipped")

			return
		}
	}

	// Mark as running (will be cleared when job completes)
	s.runningMutex.Lock()
	s.directoryRunning[name] = true
	s.runningMutex.Unlock()

	// Clear running flag when job completes (async check)
	go func() {
		// Wait for job to complete by checking state
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		timeoutChan := time.After(timeout + 5*time.Second) // Wait a bit longer than timeout

		for {
			select {
			case <-timeoutChan:
				s.runningMutex.Lock()
				delete(s.directoryRunning, name)
				s.runningMutex.Unlock()

				return
			case <-ticker.C:
				if !s.state.IsRunning(ctx, "directory", name) {
					s.runningMutex.Lock()
					delete(s.directoryRunning, name)
					s.runningMutex.Unlock()

					return
				}
			}
		}
	}()

	// Create job
	job := queue.Job{
		ID:       fmt.Sprintf("%s-%s-%d", "directory", name, time.Now().Unix()),
		Type:     "directory",
		Name:     name,
		Path:     dir.Path,
		Timeout:  timeout,
		Interval: interval,
		Context:  ctx,
	}

	// Enqueue
	if err := s.directoryQueue.Enqueue(ctx, job); err != nil {
		slog.Error("Failed to enqueue directory job", "directory", name, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.runningMutex.Lock()
		s.directoryRunning[name] = false
		s.runningMutex.Unlock()

		return
	}

	span.SetAttributes(
		attribute.String("job.id", job.ID),
		attribute.Float64("job.timeout_seconds", timeout.Seconds()),
		attribute.Float64("job.interval_seconds", interval.Seconds()),
	)
	span.AddEvent("job_scheduled")
}

// ClearRunning clears the running flag for an item
func (s *Scheduler) ClearRunning(queueType string, itemName string) {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	switch queueType {
	case "filesystem":
		delete(s.filesystemRunning, itemName)
	case "directory":
		delete(s.directoryRunning, itemName)
	}
}

// waitForJobCompletionAndEndSpan waits for a job to complete and then ends the cycle span
func (s *Scheduler) waitForJobCompletionAndEndSpan(ctx context.Context, span trace.Span, itemType, itemName string, timeout time.Duration) {
	// Wait for job to complete by checking state
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutChan := time.After(timeout + 5*time.Second) // Wait a bit longer than timeout

	for {
		select {
		case <-timeoutChan:
			span.SetAttributes(
				attribute.Bool("cycle.timeout", true),
			)
			span.SetStatus(codes.Error, "job completion timeout")
			span.End()

			return
		case <-ticker.C:
			if !s.state.IsRunning(ctx, itemType, itemName) {
				// Job completed successfully
				span.SetAttributes(
					attribute.Bool("cycle.completed", true),
				)
				span.SetStatus(codes.Ok, "job completed")
				span.End()

				return
			}
		}
	}
}

// startSpan is a helper to start an OTEL span
func (s *Scheduler) startSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if s.promexporterTracer != nil && s.promexporterTracer.IsEnabled() {
		return s.promexporterTracer.StartSpan(ctx, name, opts...)
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ctx, span
	}

	return ctx, span
}
