package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/karanshah229/gistsync/core"
)

func TestRepairConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dummy file to point to
	validPath := filepath.Join(tempDir, "valid.txt")
	os.WriteFile(validPath, []byte("test"), 0644)

	// Create a path that looks like a home dir path for another OS
	var foreignHomePath string
	if runtime.GOOS == "windows" {
		foreignHomePath = "/Users/testuser/documents/file.txt"
	} else {
		foreignHomePath = `C:\Users\testuser\documents\file.txt`
	}

	// Create a path that looks like a home dir path for THIS OS but doesn't exist
	home, _ := os.UserHomeDir()
	missingHomePath := filepath.Join(home, "nonexistent_gistsync_test_file")

	state := &core.State{
		Mappings: []core.Mapping{
			{LocalPath: validPath, RemoteID: "123"},       // VALID
			{LocalPath: foreignHomePath, RemoteID: "456"}, // REPAIRED OR MISSING (depending on if we can match)
			{LocalPath: missingHomePath, RemoteID: "789"}, // MISSING
			{LocalPath: "/tmp/not-home/file", RemoteID: "000"}, // SKIPPED
		},
	}

	results, err := RepairConfig(state)
	if err != nil {
		t.Fatalf("RepairConfig failed: %v", err)
	}

	expectedStatuses := map[string]string{
		validPath:       "VALID",
		foreignHomePath: "MISSING", // It won't exist in the current OS temp dir/home
		missingHomePath: "MISSING",
		"/tmp/not-home/file": "SKIPPED",
	}

	// Adjust expectation for foreignHomePath: 
	// The regex is: `^(/Users/[^/]+/|/home/[^/]+/|[a-zA-Z]:\\Users\\[^\\]+\\)`
	// If it matches, it becomes REPAIRED if the NEW path exists, or MISSING if it doesn't.
	
	foundForeign := false
	for _, res := range results {
		if res.OldPath == foreignHomePath {
			foundForeign = true
			if res.Status != "MISSING" && res.Status != "REPAIRED" {
				t.Errorf("Foreign path %s got status %s, expected MISSING or REPAIRED", foreignHomePath, res.Status)
			}
		}
		if status, ok := expectedStatuses[res.OldPath]; ok {
			if res.OldPath != foreignHomePath && res.Status != status {
				t.Errorf("Path %s got status %s, expected %s", res.OldPath, res.Status, status)
			}
		}
	}
	
	if !foundForeign {
		t.Errorf("Foreign path %s not found in results", foreignHomePath)
	}
}
