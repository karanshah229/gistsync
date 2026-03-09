package core

import (
	"encoding/json"
	"os"
	"path/filepath"

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
		return nil, err
	}

	state := &State{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
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

	return storage.WriteAtomic(path, data)
}
// AddMapping adds a mapping and saves state safely
func (s *State) AddMapping(m Mapping) error {
	return WithLock(func(state *State) error {
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


// WithLock executes the given function with a file lock held
func WithLock(fn func(s *State) error) error {
	path, err := storage.GetStateFilePath()
	if err != nil {
		return err
	}

	return storage.WithFileLock(path, func() error {
		state, err := LoadState()
		if err != nil {
			return err
		}

		if err := fn(state); err != nil {
			return err
		}

		return state.Save()
	})
}
