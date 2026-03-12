package core

import (
	"fmt"

	"github.com/karanshah229/gistsync/internal/domain"
)

// DetermineAction decides which sync action to take based on 3-way hash comparison
func DetermineAction(localHash, remoteHash, lastSyncedHash string) domain.SyncAction {
	if localHash == remoteHash {
		return domain.ActionNoop
	}

	if localHash == lastSyncedHash {
		// Local hasn't changed, but remote has
		return domain.ActionPull
	}

	if remoteHash == lastSyncedHash {
		// Remote hasn't changed, but local has
		return domain.ActionPush
	}

	// Both have changed from the last sync point
	return domain.ActionConflict
}

// ConflictError represents a sync conflict
type ConflictError struct {
	LocalHash      string
	RemoteHash     string
	LastSyncedHash string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("Conflict detected!\n  Local:      %s\n  Remote:     %s\n  LastSynced: %s\n(Both local and remote have changed from the last sync point)",
		truncHash(e.LocalHash), truncHash(e.RemoteHash), truncHash(e.LastSyncedHash))
}

// truncHash safely truncates a hash to 8 chars for display
func truncHash(h string) string {
	if len(h) > 8 {
		return h[:8]
	}
	return h
}
