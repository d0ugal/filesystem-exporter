package coordinator

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
	"filesystem-exporter/internal/queue"
	"filesystem-exporter/internal/scheduler"
	"filesystem-exporter/internal/state"
	"filesystem-exporter/internal/worker"

	"github.com/d0ugal/promexporter/tracing"
	"go.opentelemetry.io/otel/trace"
)

// Coordinator coordinates all components
type Coordinator struct {
	config  *config.Config
	metrics *metrics.FilesystemRegistry
	state   *state.Tracker
	tracer  *tracing.Tracer

	// Queues
	filesystemQueue *queue.Queue
	directoryQueue  *queue.Queue

	// Workers
	filesystemWorker *worker.Worker
	directoryWorker  *worker.Worker

	// Scheduler
	scheduler *scheduler.Scheduler
}

// NewCoordinator creates a new coordinator
func NewCoordinator(cfg *config.Config, m *metrics.FilesystemRegistry, tracer *tracing.Tracer) *Coordinator {
	// Create state tracker
	stateTracker := state.NewTracker(tracer)

	// Create queues
	fsQueue := queue.NewQueue("filesystem", 100, stateTracker, tracer)
	dirQueue := queue.NewQueue("directory", 100, stateTracker, tracer)

	// Create workers
	fsWorker := worker.NewWorker(fsQueue, m, stateTracker, cfg, tracer, "filesystem")
	dirWorker := worker.NewWorker(dirQueue, m, stateTracker, cfg, tracer, "directory")

	// Create scheduler
	sched := scheduler.NewScheduler(cfg, m, stateTracker, fsQueue, dirQueue, tracer)

	return &Coordinator{
		config:           cfg,
		metrics:          m,
		state:            stateTracker,
		tracer:           tracer,
		filesystemQueue:  fsQueue,
		directoryQueue:   dirQueue,
		filesystemWorker: fsWorker,
		directoryWorker:  dirWorker,
		scheduler:        sched,
	}
}

// Start starts all components
func (c *Coordinator) Start(ctx context.Context) {
	ctx, span := c.startSpan(ctx, "coordinator.start")
	defer span.End()

	slog.Info("Starting coordinator")

	// Start workers
	c.filesystemWorker.Start(ctx)
	c.directoryWorker.Start(ctx)

	// Start scheduler
	c.scheduler.Start(ctx)

	// Start goroutine count updater
	go c.updateGoroutineCount(ctx)

	// Start queue depth updater
	go c.updateQueueDepths(ctx)

	span.AddEvent("coordinator_started")
	slog.Info("Coordinator started")
}

// updateGoroutineCount periodically updates goroutine count metric
func (c *Coordinator) updateGoroutineCount(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count := runtime.NumGoroutine()
			c.metrics.GoroutineCountGauge.Set(float64(count))
		}
	}
}

// updateQueueDepths periodically updates queue depth metrics
func (c *Coordinator) updateQueueDepths(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fsDepth := c.filesystemQueue.Size(ctx)
			dirDepth := c.directoryQueue.Size(ctx)

			c.metrics.QueueDepthGauge.WithLabelValues("filesystem").Set(float64(fsDepth))
			c.metrics.QueueDepthGauge.WithLabelValues("directory").Set(float64(dirDepth))

			// Update collection active metric
			fsRunning := c.state.GetRunningJob(ctx, "filesystem")
			dirRunning := c.state.GetRunningJob(ctx, "directory")

			if fsRunning != nil {
				c.metrics.CollectionActiveGauge.WithLabelValues("filesystem").Set(1)
			} else {
				c.metrics.CollectionActiveGauge.WithLabelValues("filesystem").Set(0)
			}

			if dirRunning != nil {
				c.metrics.CollectionActiveGauge.WithLabelValues("directory").Set(1)
			} else {
				c.metrics.CollectionActiveGauge.WithLabelValues("directory").Set(0)
			}
		}
	}
}

// GetState returns the current state
func (c *Coordinator) GetState(ctx context.Context) map[string]any {
	return c.state.GetAllStates(ctx)
}

// startSpan is a helper to start an OTEL span
func (c *Coordinator) startSpan(ctx context.Context, name string, opts ...any) (context.Context, trace.Span) {
	if c.tracer != nil && c.tracer.IsEnabled() {
		return c.tracer.StartSpan(ctx, name)
	}

	return ctx, trace.SpanFromContext(ctx)
}
