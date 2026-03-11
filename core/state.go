package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/karanshah229/gistsync/internal/logger"
	"github.com/karanshah229/gistsync/internal/storage"
)

// LoadState loads the current state from the config directory
func LoadState() (*State, error) {
	path, err := storage.GetStateFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file %s: %w", path, err)
	}

	state := &State{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return state, nil
}

// Save saves the current state to the config directory atomically
func (s *State) Save() error {
	path, err := storage.GetStateFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	err = storage.WriteAtomic(path, data)
	if err == nil {
		logger.Checkpoint("State saved successfully")
	}
	return err
}
// AddMapping adds a mapping and saves state safely.
// It ensures the receiver's state is updated to match the saved content.
func (s *State) AddMapping(m Mapping) error {
	return s.WithLock(func(state *State) error {
		// Simple overwrite if local path exists
		for i, mapping := range state.Mappings {
			if mapping.LocalPath == m.LocalPath {
				state.Mappings[i] = m
				return nil
			}
		}
		state.Mappings = append(state.Mappings, m)
		return nil
	})
}

// GetMapping searches for a mapping by local path
func (s *State) GetMapping(localPath string) *Mapping {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return nil
	}
	for i := range s.Mappings {
		if s.Mappings[i].LocalPath == absPath {
			return &s.Mappings[i]
		}
	}
	return nil
}


// WithLock executes the given function with a file lock held.
// It loads the fresh state from disk, runs the function, saves the result,
// and synchronizes the caller's state object (s).
func (s *State) WithLock(fn func(freshState *State) error) error {
	path, err := storage.GetStateFilePath()
	if err != nil {
		return err
	}

	return storage.WithFileLock(path, func() error {
		freshState, err := LoadState()
		if err != nil {
			return err
		}

		if err := fn(freshState); err != nil {
			return err
		}

		if err := freshState.Save(); err != nil {
			return err
		}

		// Synchronize the caller's state object with the fresh state
		if s != nil {
			s.Mappings = freshState.Mappings
			s.Version = freshState.Version
		}
		return nil
	})
}
