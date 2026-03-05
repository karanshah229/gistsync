package core

import "time"

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
	// CheckRateLimit returns remaining requests, reset time, and error
	CheckRateLimit() (int, time.Time, error)
}
