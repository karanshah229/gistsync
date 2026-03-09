//go:build !windows

package internal

import (
	"os"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// killProcessEscalating attempts to kill a process gracefully with SIGTERM,
// then forcefully with SIGKILL if it doesn't exit.
func killProcessEscalating(pid int32) error {
	// 1. Try SIGTERM (Graceful)
	_ = syscall.Kill(int(pid), syscall.SIGTERM)

	// 2. Wait a bit
	time.Sleep(100 * time.Millisecond)

	// 3. Check if still alive and hit with SIGKILL if needed
	if err := syscall.Kill(int(pid), 0); err == nil {
		return syscall.Kill(int(pid), syscall.SIGKILL)
	}
	return nil
}

// isZombie checks if a process is in a zombie state
func isZombie(p *process.Process) bool {
	status, err := p.Status()
	if err != nil {
		return false
	}
	for _, s := range status {
		if s == "Z" {
			return true
		}
	}
	return false
}

// isChildOfCurrent checks if a process's parent is the current process
func isChildOfCurrent(p *process.Process) bool {
	ppid, err := p.Ppid()
	if err != nil {
		return false
	}
	return ppid == int32(os.Getpid())
}
