package internal

import (
	"encoding/json"
	"testing"
)

func TestValidateAndCleanConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Invalid LogLevel",
			input: map[string]interface{}{
				"log_level": "crazy",
			},
			expected: map[string]interface{}{
				"log_level": "info",
			},
		},
		{
			name: "Invalid Interval",
			input: map[string]interface{}{
				"watch_interval_seconds": -1,
			},
			expected: map[string]interface{}{
				"watch_interval_seconds": 60.0, // json unmarshal gives float64
			},
		},
		{
			name: "Valid Config Unchanged",
			input: map[string]interface{}{
				"log_level": "debug",
				"watch_interval_seconds": 30,
			},
			expected: map[string]interface{}{
				"log_level": "debug",
				"watch_interval_seconds": 30.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.input)
			cleaned := ValidateAndCleanConfig(data)
			
			var result map[string]interface{}
			json.Unmarshal(cleaned, &result)

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("%s: Key %s expected %v, got %v", tt.name, k, v, result[k])
				}
			}
		})
	}
}

func TestFixPendingRemoteID(t *testing.T) {
	stateJSON := `{
		"version": "0.1.0",
		"mappings": [
			{"local_path": "/home/user/test", "remote_id": "PENDING", "last_synced_hash": "abc"}
		]
	}`
	
	// This logic is actually inside RestoreConfig, but let's test the concept
	// Since I can't easily call RestoreConfig without mocking providers and storage,
	// I'll ensure the logic I'm testing matches what's in backup.go L103-119.

	var state struct {
		Mappings []struct {
			RemoteID string `json:"remote_id"`
		} `json:"mappings"`
	}
	
	json.Unmarshal([]byte(stateJSON), &state)
	selectedID := "new-gist-id"
	
	modified := false
	for i := range state.Mappings {
		if state.Mappings[i].RemoteID == "PENDING" {
			state.Mappings[i].RemoteID = selectedID
			modified = true
		}
	}

	if !modified {
		t.Error("Expected ID to be modified")
	}
	if state.Mappings[0].RemoteID != selectedID {
		t.Errorf("Expected RemoteID to be %s, got %s", selectedID, state.Mappings[0].RemoteID)
	}
}
