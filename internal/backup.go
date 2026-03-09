package internal

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal/storage"
)

// RestoreConfig searches for and restores gistsync configuration from a provider
func RestoreConfig(p core.Provider) (bool, error) {
	fmt.Println("🔍 Searching for backups...")
	gists, err := p.List()
	if err != nil {
		return false, fmt.Errorf("failed to list gists: %w", err)
	}

	var candidates []core.GistInfo
	for _, g := range gists {
		hasConfig := false
		hasState := false
		for _, f := range g.Files {
			if f == "config.json" {
				hasConfig = true
			}
			if f == "state.json" {
				hasState = true
			}
		}
		if hasConfig && hasState {
			candidates = append(candidates, g)
		}
	}

	if len(candidates) == 0 {
		return false, nil // No backup found
	}

	// Sort by updated_at descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].UpdatedAt.After(candidates[j].UpdatedAt)
	})

	var selectedID string
	if len(candidates) == 1 {
		message := fmt.Sprintf("Found backup: %s (Updated: %v). Restore this one?", 
			candidates[0].ID, candidates[0].UpdatedAt.Format("2006-01-02 15:04:05"))
		if !Confirm(message) {
			return false, nil
		}
		selectedID = candidates[0].ID
	} else {
		options := []string{}
		idMap := make(map[string]string)
		for _, c := range candidates {
			label := fmt.Sprintf("%s (Updated: %v)", c.ID, c.UpdatedAt.Format("2006-01-02 15:04:05"))
			options = append(options, label)
			idMap[label] = c.ID
		}

		var selection string
		prompt := &survey.Select{
			Message: "Multiple backups found. Select one to restore:",
			Options: options,
		}
		if err := survey.AskOne(prompt, &selection); err != nil {
			return false, err
		}
		selectedID = idMap[selection]
	}

	fmt.Printf("📥 Downloading backup %s...\n", selectedID)
	files, err := p.Fetch(selectedID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch backup files: %w", err)
	}

	configDir, err := storage.GetConfigDir()
	if err != nil {
		return false, err
	}

	for _, f := range files {
		if f.Path == "config.json" || f.Path == "state.json" {
			// Validate content if it's config.json
			content := f.Content
			if f.Path == "config.json" {
				content = ValidateAndCleanConfig(content)
			}

			target := filepath.Join(configDir, f.Path)
			
			// Fix "PENDING" remote_id in state.json during restoration
			if f.Path == "state.json" {
				var state core.State
				if err := json.Unmarshal(content, &state); err == nil {
					modified := false
					for i := range state.Mappings {
						if state.Mappings[i].RemoteID == "PENDING" {
							state.Mappings[i].RemoteID = selectedID
							modified = true
						}
					}
					if modified {
						if newContent, err := json.MarshalIndent(state, "", "  "); err == nil {
							content = newContent
						}
					}
				}
			}

			if err := storage.WriteAtomic(target, content); err != nil {
				return false, fmt.Errorf("failed to write %s: %w", f.Path, err)
			}
			fmt.Printf("✅ Restored %s\n", f.Path)
		}
	}

	return true, nil
}

// ValidateAndCleanConfig ensures configuration values are valid and cleans up broken ones
func ValidateAndCleanConfig(data []byte) []byte {
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return data // Return as is if unmarshal fails, though we expect valid JSON
	}

	cleaned := false
	
	// Example validation: Check LogLevel
	if val, ok := cfg["log_level"]; ok {
		level, ok := val.(string)
		allowed := []string{"debug", "info", "warn", "error"}
		isValid := false
		if ok {
			for _, a := range allowed {
				if strings.ToLower(level) == a {
					isValid = true
					break
				}
			}
		}
		if !isValid {
			cfg["log_level"] = "info"
			cleaned = true
		}
	}

	// Range checks for intervals
	if val, ok := cfg["watch_interval_seconds"]; ok {
		if v, ok := val.(float64); ok && v <= 0 {
			cfg["watch_interval_seconds"] = 60
			cleaned = true
		}
	}
	if val, ok := cfg["watch_debounce_ms"]; ok {
		if v, ok := val.(float64); ok && v < 0 {
			cfg["watch_debounce_ms"] = 500
			cleaned = true
		}
	}

	if cleaned {
		fmt.Println("⚠️  Some invalid configuration values were reset to defaults.")
		newData, err := json.MarshalIndent(cfg, "", "  ")
		if err == nil {
			return newData
		}
	}

	return data
}
