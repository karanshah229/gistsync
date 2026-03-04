package core

import (
	"testing"
)

func TestDetermineAction(t *testing.T) {
	tests := []struct {
		name           string
		local          string
		remote         string
		last           string
		expectedAction SyncAction
	}{
		{"No-op", "A", "A", "A", ActionNoop},
		{"Push", "B", "A", "A", ActionPush},
		{"Pull", "A", "B", "A", ActionPull},
		{"Conflict", "B", "C", "A", ActionConflict},
		{"No-op match but different from last", "B", "B", "A", ActionNoop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := DetermineAction(tt.local, tt.remote, tt.last)
			if actual != tt.expectedAction {
				t.Errorf("%s: Expected %s, got %s", tt.name, tt.expectedAction, actual)
			}
		})
	}
}
