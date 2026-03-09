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

	results := []GistSyncProcess{}
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		// Check if it's a gistsync process
		// We check for both "gistsync" and the full path to the executable if it matches
		if !strings.Contains(strings.ToLower(name), "gistsync") {
			cmdline, _ := p.Cmdline()
			if !strings.Contains(strings.ToLower(cmdline), "gistsync") {
				continue
			}
		}

		// Get details
		ppid, _ := p.Ppid()
		cmdline, _ := p.Cmdline()
		createTime, _ := p.CreateTime()
		cpu, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()

		results = append(results, GistSyncProcess{
			PID:       p.Pid,
			PPID:      ppid,
			Cmdline:   cmdline,
			StartTime: time.Unix(createTime/1000, 0),
			CPU:       cpu,
			Memory:    memInfo.RSS,
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

		p, err := process.NewProcess(pInfo.PID)
		if err != nil {
			continue
		}

		if err := p.Terminate(); err != nil {
			// Try killing if termination fails
			p.Kill()
		}
		killedCount++
	}

	return killedCount, nil
}
