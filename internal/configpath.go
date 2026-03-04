package internal

import (
	"os"
	"path/filepath"
)

// GetConfigDir returns the OS-appropriate configuration directory for gistsync
func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	
	dir := filepath.Join(configDir, "gistsync")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	
	return dir, nil
}

// GetStateFilePath returns the absolute path to the state.json file
func GetStateFilePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}
