package core

import "github.com/karanshah229/gistsync/internal/domain"

// File represents a single file to be synced
type File = domain.File

// Mapping represents a sync relationship between a local path and a remote gist
type Mapping = domain.Mapping

// State represents the local state of all mappings
type State = domain.State

// SyncAction represents the sync action
type SyncAction = domain.SyncAction

const (
	ActionNoop     = domain.ActionNoop
	ActionPush     = domain.ActionPush
	ActionPull     = domain.ActionPull
	ActionConflict = domain.ActionConflict
)
