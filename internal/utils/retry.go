package utils

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// RetryWithBackoff implements exponential backoff retry logic with OpenTelemetry tracing
func RetryWithBackoff(ctx context.Context, operation func() error, maxRetries int, initialDelay time.Duration) error {
	tracer := otel.Tracer("filesystem-exporter/utils")

	ctx, span := tracer.Start(ctx, "retry_with_backoff")
	defer span.End()

	span.SetAttributes(
		attribute.Int("retry.max_attempts", maxRetries+1),
		attribute.String("retry.initial_delay", initialDelay.String()),
	)

	var lastErr error

	delay := initialDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		attemptCtx, attemptSpan := tracer.Start(ctx, "retry_attempt")
		attemptSpan.SetAttributes(
			attribute.Int("retry.attempt", attempt+1),
			attribute.String("retry.delay", delay.String()),
		)

		if err := operation(); err == nil {
			attemptSpan.SetAttributes(attribute.Bool("retry.success", true))
			attemptSpan.End()
			span.SetAttributes(
				attribute.Int("retry.total_attempts", attempt+1),
				attribute.Bool("retry.final_success", true),
			)

			return nil
		} else {
			lastErr = err
			attemptSpan.SetAttributes(
				attribute.Bool("retry.success", false),
				attribute.String("retry.error", err.Error()),
			)
			attemptSpan.End()

			// Check if this error is non-retryable
			if isNonRetryableError(err) {
				slog.Warn("Operation failed with non-retryable error, skipping retries", "error", err)
				span.SetAttributes(
					attribute.Bool("retry.non_retryable", true),
					attribute.String("retry.non_retryable_reason", getNonRetryableReason(err)),
				)

				break
			}

			if attempt < maxRetries {
				slog.Warn("Operation failed, retrying", "attempt", attempt+1, "max_retries", maxRetries, "delay", delay, "error", err)

				// Create a span for the backoff delay
				_, backoffSpan := tracer.Start(attemptCtx, "retry_backoff_delay")
				backoffSpan.SetAttributes(
					attribute.String("retry.backoff_duration", delay.String()),
					attribute.Int("retry.next_attempt", attempt+2),
				)

				time.Sleep(delay)
				delay *= 2 // Exponential backoff

				backoffSpan.End()
			}
		}
	}

	span.SetAttributes(
		attribute.Int("retry.total_attempts", maxRetries+1),
		attribute.Bool("retry.final_success", false),
		attribute.String("retry.final_error", lastErr.Error()),
	)

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isNonRetryableError checks if an error should not be retried
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Context canceled errors should not be retried
	if errors.Is(err, context.Canceled) || strings.Contains(errStr, "context canceled") {
		return true
	}

	// Signal killed errors should not be retried (usually OOM or timeout)
	if strings.Contains(errStr, "signal: killed") {
		return true
	}

	// Check for exec.ExitError with specific exit codes that shouldn't be retried
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// Exit status 1 from du might indicate a persistent issue (permission, path doesn't exist, etc.)
		// We'll still retry exit status errors, but if it happens repeatedly, it's likely a real problem
		// Currently, we retry all exec.ExitError cases to handle transient issues
		// Future enhancement: could track retry counts per error type to skip persistent failures
		_ = exitErr // Suppress unused variable warning - we may use this in the future
	}

	return false
}

// getNonRetryableReason returns a human-readable reason why an error is non-retryable
func getNonRetryableReason(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	if errors.Is(err, context.Canceled) || strings.Contains(errStr, "context canceled") {
		return "context_canceled"
	}

	if strings.Contains(errStr, "signal: killed") {
		return "signal_killed"
	}

	return "unknown"
}
