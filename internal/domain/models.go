package domain

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"time"
)

// ComputeHash computes the SHA256 hash of a byte slice
func ComputeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// ComputeFileHash computes the SHA256 hash of a file at the given path
func ComputeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing %s: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// ComputeAggregateHash computes a single hash from multiple files (e.g. for a directory)
func ComputeAggregateHash(files []File) string {
	if len(files) == 0 {
		return ""
	}
	h := sha256.New()
	for _, f := range files {
		h.Write([]byte(f.Path))
		h.Write([]byte(f.Hash))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// SyncAction represents the result of a sync check
type SyncAction string

const (
	ActionNoop     SyncAction = "NOOP"
	ActionPush     SyncAction = "PUSH"
	ActionPull     SyncAction = "PULL"
	ActionConflict SyncAction = "CONFLICT"
)

// File represents a single file to be synced
type File struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
	Hash    string `json:"hash"`
}

// Mapping represents a sync relationship between a local path and a remote gist
type Mapping struct {
	LocalPath      string `json:"local_path"`
	RemoteID       string `json:"remote_id"`
	Provider       string `json:"provider"`
	IsFolder       bool   `json:"is_folder"`
	Public         bool   `json:"public"`
	LastSyncedHash string `json:"last_synced_hash"`
}

// State represents the local state of all mappings
type State struct {
	Version  string    `json:"version"`
	Mappings []Mapping `json:"mappings"`
}

// GetMapping returns the mapping for a given local path, if it exists
func (s *State) GetMapping(path string) *Mapping {
	for i := range s.Mappings {
		if s.Mappings[i].LocalPath == path {
			return &s.Mappings[i]
		}
	}
	return nil
}

// GistInfo represents summary information about a remote gist
type GistInfo struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
	Files       []string  `json:"files"`
}

// Provider defines the interface for sync backends
type Provider interface {
	Create(files []File, public bool) (string, error)
	Update(remoteID string, files []File) error
	Fetch(remoteID string) ([]File, error)
	Delete(remoteID string) error
	List() ([]GistInfo, error)
	CheckRateLimit() (int, time.Time, error)
	Verify() (bool, string, error)
}
