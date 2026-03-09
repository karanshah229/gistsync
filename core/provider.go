package core

import "time"

// GistInfo represents summary information about a remote gist
type GistInfo struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
	Files       []string  `json:"files"`
}

// Provider defines the interface for sync backends
type Provider interface {
	// Create creates a new remote gist with the given files
	Create(files []File, public bool) (string, error)
	// Update updates an existing remote gist
	Update(remoteID string, files []File) error
	// Fetch retrieves files from a remote gist
	Fetch(remoteID string) ([]File, error)
	// Delete deletes a remote gist
	Delete(remoteID string) error
	// List returns all gists owned by the user
	List() ([]GistInfo, error)
	// CheckRateLimit returns remaining requests, reset time, and error
	CheckRateLimit() (int, time.Time, error)
	// Verify checks if the provider is correctly setup (e.g., auth, dependencies)
	Verify() (bool, string, error)
}
