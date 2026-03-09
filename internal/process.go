package internal

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// GistSyncProcess represents a running instance of gistsync
type GistSyncProcess struct {
	PID       int32
	PPID      int32
	Cmdline   string
	StartTime time.Time
	CPU       float64
	Memory    uint64 // RSS in bytes
}

// ListProcesses returns a list of all running gistsync processes
func ListProcesses() ([]GistSyncProcess, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %v", err)
	}

	selfExe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get self executable path: %v", err)
	}

	results := []GistSyncProcess{}
	for _, p := range procs {
		exe, err := p.Exe()
		if err != nil {
			// Fallback to cmdline/name check if Exe() fails
			name, _ := p.Name()
			cmdline, _ := p.Cmdline()
			isGistSync := strings.Contains(strings.ToLower(name), "gistsync") || 
			              strings.Contains(strings.ToLower(cmdline), "gistsync")
			if !isGistSync {
				continue
			}
		} else {
			if exe != selfExe {
				// Also check if the base name matches in case it's a renamed binary
				name, _ := p.Name()
				if !strings.Contains(strings.ToLower(name), "gistsync") {
					continue
				}
			}
		}

		// Check for zombie or child of this process (to avoid self-kill or stale info)
		if isZombie(p) || isChildOfCurrent(p) {
			continue
		}

		// Get details
		cmdline, _ := p.Cmdline()
		ppid, _ := p.Ppid()
		createTime, _ := p.CreateTime()
		cpu, _ := p.CPUPercent()
		
		var rss uint64
		if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
			rss = memInfo.RSS
		}

		results = append(results, GistSyncProcess{
			PID:       p.Pid,
			PPID:      ppid,
			Cmdline:   cmdline,
			StartTime: time.Unix(createTime/1000, 0),
			CPU:       cpu,
			Memory:    rss,
		})
	}

	return results, nil
}

// KillOtherProcesses terminates all gistsync processes except the current one
func KillOtherProcesses() (int, error) {
	currentPID := int32(os.Getpid())
	procs, err := ListProcesses()
	if err != nil {
		return 0, err
	}

	killedCount := 0
	for _, pInfo := range procs {
		if pInfo.PID == currentPID {
			continue
		}

		if err := killProcessEscalating(pInfo.PID); err == nil {
			killedCount++
		}
	}

	return killedCount, nil
}
