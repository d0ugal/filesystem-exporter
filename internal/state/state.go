package state

import (
	"context"
	"sync"
	"time"

	"github.com/d0ugal/promexporter/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// JobState represents the state of a job
type JobState struct {
	ID        string
	Type      string // "filesystem" or "directory"
	Name      string
	Path      string
	StartedAt time.Time
	TraceID   string
}

// ItemState represents the state of a monitored item
type ItemState struct {
	Name          string
	Type          string // "filesystem" or "directory"
	LastStartTime time.Time
	LastEndTime   time.Time
	LastDuration  time.Duration
	Running       bool
	RunningJobID  string
}

// Tracker manages the state of jobs and queues
type Tracker struct {
	mu sync.RWMutex

	// Currently running jobs
	runningFilesystem *JobState
	runningDirectory  *JobState

	// Per-item state tracking
	filesystemStates map[string]*ItemState
	directoryStates  map[string]*ItemState

	// Queue depths
	filesystemQueueDepth int
	directoryQueueDepth  int

	// Tracer for OTEL spans
	tracer *tracing.Tracer
}

// NewTracker creates a new state tracker
func NewTracker(tracer *tracing.Tracer) *Tracker {
	return &Tracker{
		filesystemStates: make(map[string]*ItemState),
		directoryStates:  make(map[string]*ItemState),
		tracer:           tracer,
	}
}

// SetRunningJob sets the currently running job for a queue type
func (t *Tracker) SetRunningJob(ctx context.Context, queueType string, job *JobState) {
	_, span := t.startSpan(ctx, "state.set_running_job", trace.WithAttributes(
		attribute.String("queue.type", queueType),
		attribute.String("job.id", job.ID),
		attribute.String("job.name", job.Name),
	))
	defer span.End()

	t.mu.Lock()
	defer t.mu.Unlock()

	switch queueType {
	case "filesystem":
		t.runningFilesystem = job
		if state, exists := t.filesystemStates[job.Name]; exists {
			state.Running = true
			state.RunningJobID = job.ID
			state.LastStartTime = job.StartedAt
		}
	case "directory":
		t.runningDirectory = job
		if state, exists := t.directoryStates[job.Name]; exists {
			state.Running = true
			state.RunningJobID = job.ID
			state.LastStartTime = job.StartedAt
		}
	}

	span.SetAttributes(
		attribute.String("state.queue_type", queueType),
		attribute.String("state.job_id", job.ID),
	)
	span.AddEvent("job_state_updated")
}

// ClearRunningJob clears the currently running job for a queue type
func (t *Tracker) ClearRunningJob(ctx context.Context, queueType string, jobID string, duration time.Duration) {
	_, span := t.startSpan(ctx, "state.clear_running_job", trace.WithAttributes(
		attribute.String("queue.type", queueType),
		attribute.String("job.id", jobID),
		attribute.Float64("job.duration_seconds", duration.Seconds()),
	))
	defer span.End()

	t.mu.Lock()
	defer t.mu.Unlock()

	var jobState *JobState

	switch queueType {
	case "filesystem":
		jobState = t.runningFilesystem
		t.runningFilesystem = nil
	case "directory":
		jobState = t.runningDirectory
		t.runningDirectory = nil
	}

	if jobState != nil {
		if state, exists := t.getItemState(queueType, jobState.Name); exists {
			state.Running = false
			state.RunningJobID = ""
			state.LastEndTime = time.Now()
			state.LastDuration = duration
		}
	}

	span.AddEvent("job_state_cleared")
}

// IsRunning checks if a job is currently running for an item
func (t *Tracker) IsRunning(ctx context.Context, queueType string, itemName string) bool {
	_, span := t.startSpan(ctx, "state.is_running", trace.WithAttributes(
		attribute.String("queue.type", queueType),
		attribute.String("item.name", itemName),
	))
	defer span.End()

	t.mu.RLock()
	defer t.mu.RUnlock()

	state, exists := t.getItemState(queueType, itemName)
	if !exists {
		span.SetAttributes(attribute.Bool("state.exists", false))
		return false
	}

	running := state.Running
	span.SetAttributes(
		attribute.Bool("state.exists", true),
		attribute.Bool("state.running", running),
	)

	return running
}

// SetQueueDepth sets the queue depth for a queue type
func (t *Tracker) SetQueueDepth(ctx context.Context, queueType string, depth int) {
	_, span := t.startSpan(ctx, "state.set_queue_depth", trace.WithAttributes(
		attribute.String("queue.type", queueType),
		attribute.Int("queue.depth", depth),
	))
	defer span.End()

	t.mu.Lock()
	defer t.mu.Unlock()

	switch queueType {
	case "filesystem":
		t.filesystemQueueDepth = depth
	case "directory":
		t.directoryQueueDepth = depth
	}

	span.AddEvent("queue_depth_updated")
}

// GetQueueDepth gets the queue depth for a queue type
func (t *Tracker) GetQueueDepth(ctx context.Context, queueType string) int {
	_, span := t.startSpan(ctx, "state.get_queue_depth", trace.WithAttributes(
		attribute.String("queue.type", queueType),
	))
	defer span.End()

	t.mu.RLock()
	defer t.mu.RUnlock()

	var depth int

	switch queueType {
	case "filesystem":
		depth = t.filesystemQueueDepth
	case "directory":
		depth = t.directoryQueueDepth
	}

	span.SetAttributes(attribute.Int("queue.depth", depth))

	return depth
}

// GetRunningJob gets the currently running job for a queue type
func (t *Tracker) GetRunningJob(ctx context.Context, queueType string) *JobState {
	_, span := t.startSpan(ctx, "state.get_running_job", trace.WithAttributes(
		attribute.String("queue.type", queueType),
	))
	defer span.End()

	t.mu.RLock()
	defer t.mu.RUnlock()

	var job *JobState

	switch queueType {
	case "filesystem":
		job = t.runningFilesystem
	case "directory":
		job = t.runningDirectory
	}

	if job != nil {
		span.SetAttributes(
			attribute.String("job.id", job.ID),
			attribute.String("job.name", job.Name),
		)
	}

	return job
}

// GetItemState gets the state for an item
func (t *Tracker) GetItemState(ctx context.Context, queueType string, itemName string) *ItemState {
	_, span := t.startSpan(ctx, "state.get_item_state", trace.WithAttributes(
		attribute.String("queue.type", queueType),
		attribute.String("item.name", itemName),
	))
	defer span.End()

	t.mu.RLock()
	defer t.mu.RUnlock()

	state, exists := t.getItemState(queueType, itemName)
	if !exists {
		span.SetAttributes(attribute.Bool("state.exists", false))
		return nil
	}

	span.SetAttributes(
		attribute.Bool("state.exists", true),
		attribute.Bool("state.running", state.Running),
		attribute.String("state.last_start", state.LastStartTime.Format(time.RFC3339)),
		attribute.String("state.last_end", state.LastEndTime.Format(time.RFC3339)),
		attribute.Float64("state.last_duration_seconds", state.LastDuration.Seconds()),
	)

	// Return a copy to avoid race conditions
	return &ItemState{
		Name:          state.Name,
		Type:          state.Type,
		LastStartTime: state.LastStartTime,
		LastEndTime:   state.LastEndTime,
		LastDuration:  state.LastDuration,
		Running:       state.Running,
		RunningJobID:  state.RunningJobID,
	}
}

// RegisterItem registers an item for state tracking
func (t *Tracker) RegisterItem(ctx context.Context, queueType string, itemName string) {
	_, span := t.startSpan(ctx, "state.register_item", trace.WithAttributes(
		attribute.String("queue.type", queueType),
		attribute.String("item.name", itemName),
	))
	defer span.End()

	t.mu.Lock()
	defer t.mu.Unlock()

	state := &ItemState{
		Name: itemName,
		Type: queueType,
	}

	switch queueType {
	case "filesystem":
		t.filesystemStates[itemName] = state
	case "directory":
		t.directoryStates[itemName] = state
	}

	span.AddEvent("item_registered")
}

// GetAllStates returns all states (for debugging/status endpoint)
func (t *Tracker) GetAllStates(ctx context.Context) map[string]any {
	_, span := t.startSpan(ctx, "state.get_all_states")
	defer span.End()

	t.mu.RLock()
	defer t.mu.RUnlock()

	states := make(map[string]any)

	// Running jobs
	running := make(map[string]any)
	if t.runningFilesystem != nil {
		running["filesystem"] = map[string]any{
			"id":         t.runningFilesystem.ID,
			"name":       t.runningFilesystem.Name,
			"path":       t.runningFilesystem.Path,
			"started_at": t.runningFilesystem.StartedAt,
			"trace_id":   t.runningFilesystem.TraceID,
		}
	}

	if t.runningDirectory != nil {
		running["directory"] = map[string]any{
			"id":         t.runningDirectory.ID,
			"name":       t.runningDirectory.Name,
			"path":       t.runningDirectory.Path,
			"started_at": t.runningDirectory.StartedAt,
			"trace_id":   t.runningDirectory.TraceID,
		}
	}

	states["running"] = running

	// Queue depths
	states["queue_depths"] = map[string]int{
		"filesystem": t.filesystemQueueDepth,
		"directory":  t.directoryQueueDepth,
	}

	// Item states
	filesystemStates := make(map[string]any)
	for name, state := range t.filesystemStates {
		filesystemStates[name] = map[string]any{
			"name":           state.Name,
			"type":           state.Type,
			"running":        state.Running,
			"running_job_id": state.RunningJobID,
			"last_start":     state.LastStartTime,
			"last_end":       state.LastEndTime,
			"last_duration":  state.LastDuration.Seconds(),
		}
	}

	directoryStates := make(map[string]any)
	for name, state := range t.directoryStates {
		directoryStates[name] = map[string]any{
			"name":           state.Name,
			"type":           state.Type,
			"running":        state.Running,
			"running_job_id": state.RunningJobID,
			"last_start":     state.LastStartTime,
			"last_end":       state.LastEndTime,
			"last_duration":  state.LastDuration.Seconds(),
		}
	}

	states["filesystem_items"] = filesystemStates
	states["directory_items"] = directoryStates

	return states
}

// getItemState is a helper that doesn't require locking (caller must hold lock)
func (t *Tracker) getItemState(queueType string, itemName string) (*ItemState, bool) {
	switch queueType {
	case "filesystem":
		state, exists := t.filesystemStates[itemName]
		return state, exists
	case "directory":
		state, exists := t.directoryStates[itemName]
		return state, exists
	}

	return nil, false
}

// startSpan is a helper to start an OTEL span
func (t *Tracker) startSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t.tracer != nil && t.tracer.IsEnabled() {
		return t.tracer.StartSpan(ctx, name, opts...)
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ctx, span
	}

	return ctx, span
}

// RecordError records an error on the current span
func (t *Tracker) RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
