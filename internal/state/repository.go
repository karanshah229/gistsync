package state

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/internal/domain"
	"github.com/karanshah229/gistsync/internal/logger"
	"github.com/karanshah229/gistsync/internal/storage"
)

// Repository handles persistence of the sync state
type Repository interface {
	Load() (*domain.State, error)
	Save(state *domain.State) error
	WithLock(fn func(state *domain.State) error) error
	GetMapping(absPath string) (*domain.Mapping, error)
	AddMapping(m domain.Mapping) error
}

type fileRepository struct {
	statePath string
}

// NewFileRepository creates a new file-based state repository
func NewFileRepository() (Repository, error) {
	path, err := storage.GetStateFilePath()
	if err != nil {
		return nil, err
	}
	return &fileRepository{statePath: path}, nil
}

func (r *fileRepository) Load() (*domain.State, error) {
	data, err := os.ReadFile(r.statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	state := &domain.State{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return state, nil
}

func (r *fileRepository) Save(state *domain.State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	err = storage.WriteAtomic(r.statePath, data)
	if err == nil {
		logger.Checkpoint("State saved successfully")
	}
	return err
}

func (r *fileRepository) WithLock(fn func(state *domain.State) error) error {
	return storage.WithFileLock(r.statePath, func() error {
		state, err := r.Load()
		if err != nil {
			return err
		}

		if err := fn(state); err != nil {
			return err
		}

		return r.Save(state)
	})
}

func (r *fileRepository) GetMapping(absPath string) (*domain.Mapping, error) {
	state, err := r.Load()
	if err != nil {
		return nil, err
	}
	for i := range state.Mappings {
		if state.Mappings[i].LocalPath == absPath {
			return &state.Mappings[i], nil
		}
	}
	return nil, nil
}

func (r *fileRepository) AddMapping(m domain.Mapping) error {
	return r.WithLock(func(state *domain.State) error {
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
