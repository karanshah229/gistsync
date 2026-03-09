package internal

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/karanshah229/gistsync/internal/storage"
)

// Config represents user-level application settings
type Config struct {
	WatchInterval   int               `json:"watch_interval_seconds"`
	WatchDebounce   int               `json:"watch_debounce_ms"`
	LogLevel        string            `json:"log_level"`
	DefaultProvider string            `json:"default_provider"`
	Autostart       bool              `json:"autostart"`
	Providers       map[string]string `json:"providers"` // Store provider status
}

// ConfigOption defines metadata for a configuration field
type ConfigOption struct {
	Key          string
	Default      interface{}
	Description  string
	Prompt       string
}

// GetConfigOptions returns the list of configurable options
func GetConfigOptions() []ConfigOption {
	return []ConfigOption{
		{
			Key:         "WatchInterval",
			Default:     60,
			Description: "Interval in seconds to poll for remote changes",
			Prompt:      "Watch Interval (seconds)",
		},
		{
			Key:         "WatchDebounce",
			Default:     500,
			Description: "Delay in milliseconds before syncing local changes",
			Prompt:      "Watch Debounce (ms)",
		},
		{
			Key:         "LogLevel",
			Default:     "info",
			Description: "Logging verbosity (debug, info, warn, error)",
			Prompt:      "Log Level",
		},
		{
			Key:         "Autostart",
			Default:     true,
			Description: "Automatically start gistsync at login",
			Prompt:      "Enable Autostart",
		},
	}
}

// DefaultConfig returns the default settings
func DefaultConfig() *Config {
	return &Config{
		WatchInterval:   60,
		WatchDebounce:   500,
		LogLevel:        "info",
		DefaultProvider: "github",
		Autostart:       true,
		Providers:       make(map[string]string),
	}
}

// GetConfigFilePath returns the absolute path to the config.json file
func GetConfigFilePath() (string, error) {
	dir, err := storage.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// IsConfigPresent checks if the config file exists on disk
func IsConfigPresent() bool {
	path, err := GetConfigFilePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// LoadConfig reads the config file from disk. It does NOT return defaults if the file is missing.
func LoadConfig() (*Config, error) {
	path, err := GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if config.Providers == nil {
		config.Providers = make(map[string]string)
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

	return storage.WriteAtomic(path, data)
}
