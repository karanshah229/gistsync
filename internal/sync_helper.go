package internal

import (
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/core"
)

// SyncAll iterates through all mappings in the state and performs a sync for each
func SyncAll(engine *core.Engine) {
	if len(engine.State.Mappings) == 0 {
		fmt.Println("No mappings found to sync. Use 'gistsync sync <path>' to start tracking files.")
		return
	}

	fmt.Printf("🔄 Syncing all %d mappings...\n", len(engine.State.Mappings))
	for _, m := range engine.State.Mappings {
		fmt.Printf("Syncing %s...\n", m.LocalPath)
		var err error
		if m.IsFolder {
			err = engine.SyncDir(m.LocalPath)
		} else {
			err = engine.SyncFile(m.LocalPath)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Sync failed for %s: %v\n", m.LocalPath, err)
		} else {
			fmt.Printf("✅ %s synced successfully\n", m.LocalPath)
		}
	}
}
