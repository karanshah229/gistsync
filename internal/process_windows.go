//go:build windows

package internal

import (
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

// killProcessEscalating attempts to kill a process on Windows.
// Windows doesn't have standard POSIX signals like SIGTERM/SIGKILL in the same way,
// so we fall back to the standard library's Kill() which is forceful.
func killProcessEscalating(pid int32) error {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}
	return proc.Kill()
}

// isZombie checks if a process is in a zombie state.
// Zombies aren't really a thing on Windows in the Unix sense.
func isZombie(p *process.Process) bool {
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
