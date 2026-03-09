package storage

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigDir returns the OS-appropriate configuration directory for gistsync
func GetConfigDir() (string, error) {
	// 1. Honor XDG_CONFIG_HOME if set
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		dir := filepath.Join(xdg, "gistsync")
		return dir, os.MkdirAll(dir, 0755)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// 2. For Mac/Linux, prefer ~/.config/gistsync over Application Support
	if runtime.GOOS != "windows" {
		dir := filepath.Join(home, ".config", "gistsync")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
		return dir, nil
	}

	// 3. For Windows, use standard UserConfigDir (%AppData%)
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
