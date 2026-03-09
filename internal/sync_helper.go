package internal

import (
	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/pkg/ui"
)

// SyncAll iterates through all mappings in the state and performs a sync for each
func SyncAll(engine *core.Engine) {
	if len(engine.State.Mappings) == 0 {
		ui.Print("NoMappingsFound", nil)
		return
	}

	ui.Print("SyncingAllMappings", map[string]interface{}{"Count": len(engine.State.Mappings)})
	for _, m := range engine.State.Mappings {
		ui.Print("SyncingPath", map[string]interface{}{"Path": m.LocalPath})
		var err error
		if m.IsFolder {
			err = engine.SyncDir(m.LocalPath)
		} else {
			err = engine.SyncFile(m.LocalPath)
		}

		if err != nil {
			ui.Error("SyncFailed", map[string]interface{}{"Path": m.LocalPath, "Err": err})
		} else {
			ui.Success("SyncSuccess", map[string]interface{}{"Path": m.LocalPath})
		}
	}
}
