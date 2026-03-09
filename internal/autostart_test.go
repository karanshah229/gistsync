package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type mockPathProvider struct {
	tempDir string
}

func (m mockPathProvider) GetMacPlistPath() (string, error) {
	return filepath.Join(m.tempDir, "mac", macPlistName), nil
}

func (m mockPathProvider) GetLinuxUnitPath() (string, error) {
	return filepath.Join(m.tempDir, "linux", linuxUnitName), nil
}

func (m mockPathProvider) GetWindowsShortcutPath() (string, error) {
	return filepath.Join(m.tempDir, "windows", windowsShortcut), nil
}

func TestIsAutostartEnabled(t *testing.T) {
	tempDir := t.TempDir()
	m := mockPathProvider{tempDir: tempDir}
	// Backup and restore currentPathProvider
	oldProvider := currentPathProvider
	currentPathProvider = m
	defer func() { currentPathProvider = oldProvider }()

	// 1. Test initially disabled
	enabled, err := IsAutostartEnabled()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if enabled {
		t.Error("Expected autostart to be disabled")
	}

	// 2. Mock enable by creating the file
	var path string
	switch runtime.GOOS {
	case "darwin":
		path, _ = m.GetMacPlistPath()
	case "linux":
		path, _ = m.GetLinuxUnitPath()
	case "windows":
		path, _ = m.GetWindowsShortcutPath()
	default:
		t.Skip("Unsupported OS for this test")
	}

	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte("test"), 0644)

	enabled, err = IsAutostartEnabled()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !enabled {
		t.Error("Expected autostart to be enabled")
	}
}

func TestInstallAutostartContent(t *testing.T) {
	tempDir := t.TempDir()
	m := mockPathProvider{tempDir: tempDir}
	oldProvider := currentPathProvider
	currentPathProvider = m
	defer func() { currentPathProvider = oldProvider }()

	exe := "/path/to/gistsync"
	logPath := "/path/to/gistsync.log"

	switch runtime.GOOS {
	case "darwin":
		_ = installMacOS(exe, logPath)
		path, _ := m.GetMacPlistPath()
		data, errRead := os.ReadFile(path)
		if errRead == nil {
			content := string(data)
			if !strings.Contains(content, exe) || !strings.Contains(content, logPath) {
				t.Errorf("Plist content incorrect: %s", content)
			}
		}
	case "linux":
		_ = installLinux(exe, logPath)
		path, _ := m.GetLinuxUnitPath()
		data, errRead := os.ReadFile(path)
		if errRead == nil {
			content := string(data)
			if !strings.Contains(content, exe) || !strings.Contains(content, logPath) {
				t.Errorf("Service file content incorrect: %s", content)
			}
		}
	}
}
