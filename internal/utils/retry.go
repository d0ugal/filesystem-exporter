package utils

import (
	"context"
	"fmt"
	"log/slog"
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
