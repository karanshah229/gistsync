package internal

import (
	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/pkg/ui"
)

// SyncAll iterates through all mappings in the state and performs a sync for each.
// It returns the number of successful and failed syncs.
func SyncAll(engine *core.Engine) (successCount, failCount int) {
	if len(engine.State.Mappings) == 0 {
		ui.Print("NoMappingsFound", nil)
		return 0, 0
	}

	ui.Print("SyncingAllMappings", map[string]interface{}{"Count": len(engine.State.Mappings)})
	for _, m := range engine.State.Mappings {
		var action core.SyncAction
		var err error
		if m.IsFolder {
			action, err = engine.SyncDir(m.LocalPath)
		} else {
			action, err = engine.SyncFile(m.LocalPath)
		}

		if err != nil {
			ui.Error("SyncFailed", map[string]interface{}{"Path": m.LocalPath, "Err": err})
			failCount++
		} else {
			successCount++
			switch action {
			case core.ActionNoop:
				ui.Print("SyncNoop", map[string]interface{}{"Path": m.LocalPath})
			case core.ActionPush:
				ui.Success("SyncPushed", map[string]interface{}{"Path": m.LocalPath})
			case core.ActionPull:
				ui.Success("SyncPulled", map[string]interface{}{"Path": m.LocalPath})
			default:
				ui.Success("SyncSuccess", map[string]interface{}{"Path": m.LocalPath})
			}
		}
	}

	ui.Print("SyncAllSummary", map[string]interface{}{
		"Success": successCount,
		"Failed":  failCount,
		"Total":   len(engine.State.Mappings),
	})
	return successCount, failCount
}
