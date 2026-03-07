package core

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
