package i18n

import (
	"testing"
)

func TestT(t *testing.T) {
	// Simple translation
	if res := T("Initializing", nil); res != "Initializing gistsync..." {
		t.Errorf("Expected 'Initializing gistsync...', got '%s'", res)
	}

	// Template data
	data := map[string]interface{}{"Msg": "Test Connection"}
	if res := T("GitHubConnected", data); res != "GitHub: Connected (Test Connection)" {
		t.Errorf("Expected 'GitHub: Connected (Test Connection)', got '%s'", res)
	}

	// Missing key fallback
	if res := T("NonExistentKey", nil); res != "!NonExistentKey!" {
		t.Errorf("Expected '!NonExistentKey!', got '%s'", res)
	}
}

func TestSetLanguage(t *testing.T) {
	// This might be hard to test without another locale file, 
	// but we can check if it at least doesn't panic.
	SetLanguage("en")
	if res := T("Initializing", nil); res != "Initializing gistsync..." {
		t.Errorf("Fallback failed after SetLanguage")
	}
}
