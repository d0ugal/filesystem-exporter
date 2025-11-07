package queue

import (
	"context"
	"time"

	"filesystem-exporter/internal/state"
	"github.com/d0ugal/promexporter/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Job represents a collection job
type Job struct {
	ID        string
	Type      string // "filesystem" or "directory"
	Name      string
	Path      string
	Timeout   time.Duration
	Interval  time.Duration
	CreatedAt time.Time
	Context   context.Context // Context with trace span
}

// Queue represents a job queue
type Queue struct {
	jobs   chan Job
	state  *state.Tracker
	tracer *tracing.Tracer
	name   string // "filesystem" or "directory"
}

// NewQueue creates a new queue
func NewQueue(name string, bufferSize int, stateTracker *state.Tracker, tracer *tracing.Tracer) *Queue {
	return &Queue{
		jobs:   make(chan Job, bufferSize),
		state:  stateTracker,
		tracer: tracer,
		name:   name,
	}
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	ctx, span := q.startSpan(ctx, "queue.enqueue", trace.WithAttributes(
		attribute.String("queue.name", q.name),
		attribute.String("job.id", job.ID),
		attribute.String("job.type", job.Type),
		attribute.String("job.name", job.Name),
		attribute.String("job.path", job.Path),
		attribute.Float64("job.timeout_seconds", job.Timeout.Seconds()),
		attribute.Float64("job.interval_seconds", job.Interval.Seconds()),
	))
	defer span.End()

	// Preserve the parent context from scheduler - the span we created above is already
	// a child of the scheduler span, so we use ctx (which has our span) as the job context
	job.Context = ctx
	job.CreatedAt = time.Now()

	select {
	case q.jobs <- job:
		// Update queue depth
		depth := len(q.jobs)
		q.state.SetQueueDepth(ctx, q.name, depth)

		span.SetAttributes(
			attribute.Int("queue.depth_after", depth),
			attribute.Float64("queue.wait_time_seconds", 0),
		)
		span.AddEvent("job_queued")

		return nil
	case <-ctx.Done():
		err := ctx.Err()
		span.RecordError(err)
		span.SetStatus(codes.Error, "context cancelled")

		return err
	}
}

// Dequeue removes and returns a job from the queue
func (q *Queue) Dequeue(ctx context.Context) (Job, error) {
	ctx, span := q.startSpan(ctx, "queue.dequeue", trace.WithAttributes(
		attribute.String("queue.name", q.name),
	))
	defer span.End()

	waitStart := time.Now()

	select {
	case job := <-q.jobs:
		waitDuration := time.Since(waitStart)

		// Update queue depth
		depth := len(q.jobs)
		q.state.SetQueueDepth(ctx, q.name, depth)

		span.SetAttributes(
			attribute.String("job.id", job.ID),
			attribute.String("job.type", job.Type),
			attribute.String("job.name", job.Name),
			attribute.Float64("queue.wait_time_seconds", waitDuration.Seconds()),
			attribute.Int("queue.depth_after", depth),
		)
		span.AddEvent("job_dequeued")

		return job, nil
	case <-ctx.Done():
		err := ctx.Err()
		span.RecordError(err)
		span.SetStatus(codes.Error, "context cancelled")

		return Job{}, err
	}
}

// Size returns the current queue size
func (q *Queue) Size(ctx context.Context) int {
	_, span := q.startSpan(ctx, "queue.size", trace.WithAttributes(
		attribute.String("queue.name", q.name),
	))
	defer span.End()

	size := len(q.jobs)
	span.SetAttributes(attribute.Int("queue.size", size))

	return size
}

// Channel returns the underlying channel (for use in select statements)
func (q *Queue) Channel() <-chan Job {
	return q.jobs
}

// startSpan is a helper to start an OTEL span
func (q *Queue) startSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if q.tracer != nil && q.tracer.IsEnabled() {
		return q.tracer.StartSpan(ctx, name, opts...)
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ctx, span
	}

	return ctx, span
}
