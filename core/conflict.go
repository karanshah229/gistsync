package core

import "fmt"

type SyncAction string

const (
	ActionNoop     SyncAction = "NOOP"
	ActionPush     SyncAction = "PUSH"
	ActionPull     SyncAction = "PULL"
	ActionConflict SyncAction = "CONFLICT"
)

// DetermineAction decides which sync action to take based on 3-way hash comparison
func DetermineAction(localHash, remoteHash, lastSyncedHash string) SyncAction {
	if localHash == remoteHash {
		return ActionNoop
	}

	if localHash == lastSyncedHash {
		// Local hasn't changed, but remote has
		return ActionPull
	}

	if remoteHash == lastSyncedHash {
		// Remote hasn't changed, but local has
		return ActionPush
	}

	// Both have changed from the last sync point
	return ActionConflict
}

// ConflictError represents a sync conflict
type ConflictError struct {
	LocalHash      string
	RemoteHash     string
	LastSyncedHash string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("Conflict detected!\n  Local:      %s\n  Remote:     %s\n  LastSynced: %s\n(Both local and remote have changed from the last sync point)",
		e.LocalHash, e.RemoteHash, e.LastSyncedHash)
}
