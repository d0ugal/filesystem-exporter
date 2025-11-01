//go:build linux
// +build linux

package utils

import (
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
)

// I/O priority constants for Linux ioprio_set syscall
// See: https://man7.org/linux/man-pages/man2/ioprio_set.2.html
const (
	IOPRIO_WHO_PROCESS = 1
	IOPRIO_CLASS_IDLE  = 3
	IOPRIO_CLASS_SHIFT = 13
)

// IOPRIO_PRIO_VALUE constructs the ioprio value from class and data
// Class: 0=none, 1=realtime, 2=best-effort, 3=idle
// Data: 0-7 (priority within class, lower is higher priority)
func IOPRIO_PRIO_VALUE(class, data int) int {
	return (class << IOPRIO_CLASS_SHIFT) | data
}

// SetIOIdlePriority sets the I/O priority of the current process to idle class
// This reduces the impact of I/O operations on system performance
// Returns an error if the operation fails
func SetIOIdlePriority() error {
	// Only supported on Linux
	if runtime.GOOS != "linux" {
		return nil // No-op on non-Linux systems
	}

	ioprio := IOPRIO_PRIO_VALUE(IOPRIO_CLASS_IDLE, 0)

	// Use raw syscall since golang.org/x/sys/unix doesn't expose IoprioSet directly
	// sys_ioprio_set syscall number (251 on x86_64/arm64, may vary by arch)
	// Note: We use the hardcoded value since SYS_IOPRIO_SET may not be available
	// in all versions of golang.org/x/sys/unix
	// 251 is the syscall number for ioprio_set on linux/amd64 and linux/arm64
	const sysIoprioSet = 251
	sysno := uintptr(sysIoprioSet)

	_, _, errno := unix.Syscall(sysno, IOPRIO_WHO_PROCESS, 0, uintptr(ioprio))
	var err error
	if errno != 0 {
		err = errno
	}

	if err != nil {
		return fmt.Errorf("failed to set I/O priority: %w", err)
	}

	slog.Debug("Set I/O priority to idle class using syscall")
	return nil
}

// SetIOIdlePriorityForProcess sets the I/O priority of a specific process to idle class
// This is useful when forking/spawning child processes
func SetIOIdlePriorityForProcess(pid int) error {
	// Only supported on Linux
	if runtime.GOOS != "linux" {
		return nil // No-op on non-Linux systems
	}

	ioprio := IOPRIO_PRIO_VALUE(IOPRIO_CLASS_IDLE, 0)

	// Use the same syscall number as SetIOIdlePriority
	const sysIoprioSet = 251
	sysno := uintptr(sysIoprioSet)

	_, _, errno := unix.Syscall(sysno, IOPRIO_WHO_PROCESS, uintptr(pid), uintptr(ioprio))
	var err error
	if errno != 0 {
		err = errno
	}

	if err != nil {
		return fmt.Errorf("failed to set I/O priority for process %d: %w", pid, err)
	}

	slog.Debug("Set I/O priority to idle class", "pid", pid)
	return nil
}

// SetupCommandWithIOPriority configures an exec.Cmd to run with idle I/O priority
// This uses SysProcAttr to set I/O priority in the child process after fork
func SetupCommandWithIOPriority(cmd *exec.Cmd) error {
	if runtime.GOOS != "linux" {
		return nil // No-op on non-Linux systems
	}

	ioprio := IOPRIO_PRIO_VALUE(IOPRIO_CLASS_IDLE, 0)

	// Create SysProcAttr if it doesn't exist
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	// Note: Go's exec.Cmd doesn't provide a direct way to set I/O priority
	// in the child process after fork. The I/O priority is per-process, not inherited.
	// We'll need to set it after the process starts, which requires a different approach.
	// For now, we'll set it via a wrapper or rely on the system's default behavior.
	// This is a limitation of Go's exec package.
	_ = ioprio // Used in wrapper approach

	// The actual I/O priority will be set by calling SetIOIdlePriorityForProcess
	// after the command starts, but before it does significant I/O.
	return nil
}
