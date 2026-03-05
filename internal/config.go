package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents user-level application settings
type Config struct {
	WatchInterval int    `json:"watch_interval_seconds"`
	WatchDebounce int    `json:"watch_debounce_ms"`
	LogLevel      string `json:"log_level"`
}

// DefaultConfig returns the default settings
func DefaultConfig() *Config {
	return &Config{
		WatchInterval: 60,
		WatchDebounce: 500,
		LogLevel:      "info",
	}
}

// GetConfigFilePath returns the absolute path to the config.json file
func GetConfigFilePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig reads the config file from disk or returns defaults
func LoadConfig() (*Config, error) {
	path, err := GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return DefaultConfig(), nil // Rollback to defaults on corruption
	}

	return config, nil
}

// SaveConfig writes the config file to disk atomically
func SaveConfig(config *Config) error {
	path, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return WriteAtomic(path, data)
}
