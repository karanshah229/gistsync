package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(os.Stdout)

	f()

	return buf.String()
}

func TestUIFunctions(t *testing.T) {
	tests := []struct {
		name     string
		f        func()
		contains string
	}{
		{
			name:     "Success",
			f:        func() { Success("Ready", nil) },
			contains: "✅ gistsync is ready!",
		},
		{
			name:     "Error",
			f:        func() { Error("ProcessNotRunning", nil) },
			contains: "❌ gistsync is not running.",
		},
		{
			name:     "Info",
			f:        func() { Info("ReadyWithHint", nil) },
			contains: "💡 gistsync is ready!",
		},
		{
			name:     "Warning",
			f:        func() { Warning("BackupFailed", map[string]interface{}{"Err": "test err"}) },
			contains: "⚠️  Backup failed: test err",
		},
		{
			name:     "Header",
			f:        func() { Header("CheckingProviders", nil) },
			contains: "Checking Providers...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(tt.f)
			if !strings.Contains(output, tt.contains) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.contains, output)
			}
		})
	}
}
