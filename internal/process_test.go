package internal

import (
	"os"
	"testing"
)

func TestListProcesses(t *testing.T) {
	procs, err := ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	// Since we are running this test via 'go test', there should be at least one process
	// that matches "gistsync" (the test binary itself often contains the package name)
	// or at least we can verify it doesn't crash and returns a valid slice.
	
	currentPID := int32(os.Getpid())
	for _, p := range procs {
		if p.PID == currentPID {
			// Found the current process, good.
			break
		}
	}

	// Note: 'go test' might name the binary something like 'internal.test'
	// so it might not be found by our "gistsync" string filter.
	// But ListProcesses should at least return a non-nil slice.
	if procs == nil {
		t.Error("Expected non-nil process list")
	}
}

func TestProcessFiltering(t *testing.T) {
	// This is hard to test without a mockable process source.
	// However, we can at least verify the logic of ListProcesses 
	// doesn't error out on a standard system.
	_, err := ListProcesses()
	if err != nil {
		t.Errorf("ListProcesses failed: %v", err)
	}
}
