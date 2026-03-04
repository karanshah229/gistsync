package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/karan/gistsync/internal"
)

// LoadState loads the current state from the config directory
func LoadState() (*State, error) {
	path, err := internal.GetStateFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state if file doesn't exist
			return &State{
				Version:  1,
				Mappings: []Mapping{},
			}, nil
		}
		return nil, err
	}

	state := &State{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}

	return state, nil
}

// SaveState saves the current state to the config directory
func (s *State) Save() error {
	path, err := internal.GetStateFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddMapping adds a mapping and saves state
func (s *State) AddMapping(m Mapping) error {
	// Simple overwrite if local path exists
	for i, mapping := range s.Mappings {
		if mapping.LocalPath == m.LocalPath {
			s.Mappings[i] = m
			return s.Save()
		}
	}
	s.Mappings = append(s.Mappings, m)
	return s.Save()
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
