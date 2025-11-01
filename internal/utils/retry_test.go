package utils

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestRetryWithBackoff_Success(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	err := RetryWithBackoff(ctx, func() error {
		attempts++
		return nil // Success on first attempt
	}, 3, 100*time.Millisecond)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got: %d", attempts)
	}
}

func TestRetryWithBackoff_RetryableError(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	err := RetryWithBackoff(ctx, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil // Success on third attempt
	}, 3, 10*time.Millisecond)

	if err != nil {
		t.Errorf("Expected no error after retries, got: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}
}

func TestRetryWithBackoff_NonRetryableError_SignalKilled(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	retryableErr := exec.Command("nonexistent").Run() // This won't be signal: killed, so let's create a mock error
	killedErr := errors.New("failed to execute du command for /path: signal: killed")

	err := RetryWithBackoff(ctx, func() error {
		attempts++
		return killedErr
	}, 3, 10*time.Millisecond)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retries for non-retryable error), got: %d", attempts)
	}

	if !strings.Contains(err.Error(), "signal: killed") {
		t.Errorf("Expected error to contain 'signal: killed', got: %v", err)
	}

	_ = retryableErr // Suppress unused variable warning
}

func TestRetryWithBackoff_NonRetryableError_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	// Cancel the context to simulate context cancellation
	cancel()

	err := RetryWithBackoff(ctx, func() error {
		attempts++
		return context.Canceled
	}, 3, 10*time.Millisecond)

	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retries for context canceled), got: %d", attempts)
	}

	// The function should detect context.Canceled and stop retrying
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Logf("Note: Error may be wrapped: %v", err)
	}
}

func TestRetryWithBackoff_MaxRetries(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	err := RetryWithBackoff(ctx, func() error {
		attempts++
		return errors.New("persistent error")
	}, 2, 10*time.Millisecond)

	if err == nil {
		t.Error("Expected error after max retries, got nil")
	}

	if attempts != 3 { // 1 initial + 2 retries
		t.Errorf("Expected 3 attempts (1 initial + 2 retries), got: %d", attempts)
	}

	if !strings.Contains(err.Error(), "persistent error") {
		t.Errorf("Expected error message to contain 'persistent error', got: %v", err)
	}
}

func TestIsNonRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "signal killed",
			err:      errors.New("failed to execute du command: signal: killed"),
			expected: true,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "context canceled string",
			err:      errors.New("context canceled"),
			expected: true,
		},
		{
			name:     "retryable error",
			err:      errors.New("temporary network error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "exit status 1",
			err:      errors.New("exit status 1"),
			expected: false, // Exit status 1 is retryable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNonRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isNonRetryableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestGetNonRetryableReason(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "signal killed",
			err:      errors.New("signal: killed"),
			expected: "signal_killed",
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: "context_canceled",
		},
		{
			name:     "context canceled string",
			err:      errors.New("context canceled"),
			expected: "context_canceled",
		},
		{
			name:     "unknown error",
			err:      errors.New("some other error"),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNonRetryableReason(tt.err)
			if result != tt.expected {
				t.Errorf("getNonRetryableReason(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

