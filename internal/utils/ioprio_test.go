//go:build linux
// +build linux

package utils

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestIOPRIO_PRIO_VALUE(t *testing.T) {
	tests := []struct {
		name  string
		class int
		data  int
		want  int
	}{
		{
			name:  "idle class",
			class: IOPRIO_CLASS_IDLE,
			data:  0,
			want:  IOPRIO_CLASS_IDLE << IOPRIO_CLASS_SHIFT,
		},
		{
			name:  "idle class with data",
			class: IOPRIO_CLASS_IDLE,
			data:  4,
			want:  (IOPRIO_CLASS_IDLE << IOPRIO_CLASS_SHIFT) | 4,
		},
		{
			name:  "best effort class",
			class: 2,
			data:  0,
			want:  2 << IOPRIO_CLASS_SHIFT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IOPRIO_PRIO_VALUE(tt.class, tt.data)
			if got != tt.want {
				t.Errorf("IOPRIO_PRIO_VALUE(%d, %d) = %d, want %d", tt.class, tt.data, got, tt.want)
			}
		})
	}
}

func TestSetIOIdlePriority(t *testing.T) {
	// This test requires Linux and appropriate permissions
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test: requires Linux")
	}

	// Try to set I/O priority - this may fail if we don't have permissions
	// but we should at least test that the function doesn't crash
	err := SetIOIdlePriority()
	
	// On systems without permissions, we'll get an error, which is expected
	// We just want to ensure the function doesn't panic
	if err != nil {
		t.Logf("SetIOIdlePriority returned error (expected if no permissions): %v", err)
	}
}

func TestSetIOIdlePriorityForProcess(t *testing.T) {
	// This test requires Linux
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test: requires Linux")
	}

	// Test with current process PID
	pid := os.Getpid()
	err := SetIOIdlePriorityForProcess(pid)

	// May fail if we don't have permissions, but shouldn't panic
	if err != nil {
		t.Logf("SetIOIdlePriorityForProcess returned error (expected if no permissions): %v", err)
	}

	// Test with invalid PID (should fail)
	err = SetIOIdlePriorityForProcess(-1)
	if err == nil {
		t.Log("SetIOIdlePriorityForProcess with invalid PID didn't return error")
	} else {
		t.Logf("SetIOIdlePriorityForProcess with invalid PID correctly returned error: %v", err)
	}
}

func TestSetupCommandWithIOPriority(t *testing.T) {
	// This test requires Linux
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test: requires Linux")
	}

	cmd := exec.Command("echo", "test")
	err := SetupCommandWithIOPriority(cmd)

	if err != nil {
		t.Errorf("SetupCommandWithIOPriority returned error: %v", err)
	}

	if cmd.SysProcAttr == nil {
		t.Error("SetupCommandWithIOPriority should set SysProcAttr")
	}
}

func TestSetIOIdlePriority_NonLinux(t *testing.T) {
	// This test can't actually run on non-Linux, but we can verify
	// the logic is correct by checking the runtime check

	// Since we're on Linux, we can't test the non-Linux path directly
	// but we can verify the function handles it correctly in the code
	if runtime.GOOS != "linux" {
		err := SetIOIdlePriority()
		if err != nil {
			t.Errorf("SetIOIdlePriority should return nil on non-Linux, got: %v", err)
		}
	}
}

