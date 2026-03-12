package cmd

import (
	"os"
	"path/filepath"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/pkg/i18n"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/karanshah229/gistsync/providers"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [path]",
	Short: "Sync a file or directory to a gist (creates a new gist if not already tracked)",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		state, err := core.LoadState()
		if err != nil {
			ui.Error("LoadStateFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		var providerName string
		providerFlag, _ := cmd.Flags().GetString("provider")
		if providerFlag != "" {
			providerName = providerFlag
		} else {
			cfg, _ := internal.LoadConfig()
			if cfg != nil && cfg.DefaultProvider != "" {
				providerName = cfg.DefaultProvider
			} else {
				providerName = "github"
			}
		}

		var provider core.Provider
		switch providerName {
		case "gitlab":
			provider = providers.NewGitLabProvider()
		default:
			provider = providers.NewGitHubProvider()
		}

		engine := core.NewEngine(state, provider)

		if len(args) == 0 {
			internal.SyncAll(engine)
			return
		}

		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			ui.Error("AbsPathFailed", map[string]interface{}{"Path": path, "Err": err})
			os.Exit(1)
		}

		if len(args) == 2 {
			gistID := args[1]
			
			info, err := os.Stat(path)
			if err != nil {
				ui.Error("StatPathFailed", map[string]interface{}{"Path": path, "Err": err})
				os.Exit(1)
			}
			
			ui.Print("ManualMappingStart", map[string]interface{}{
				"Path":     path,
				"ID":       gistID,
				"Provider": providerName,
			})

			mapping := state.GetMapping(absPath)
			if mapping != nil {
				msg := i18n.T("MappingOverwriteConfirm", map[string]interface{}{
					"Path":  path,
					"OldID": mapping.RemoteID,
					"NewID": gistID,
				})
				confirm := ui.Confirm("ConfirmQuestion", map[string]interface{}{
					"Message": msg,
				})
				if !confirm {
					ui.Print("MappingOverwriteCancel", nil)
					return
				}
			}

			err = state.AddMapping(core.Mapping{
				LocalPath:      absPath,
				RemoteID:       gistID,
				Provider:       providerName,
				IsFolder:       info.IsDir(),
				Public:         false,
				LastSyncedHash: "",
			})
			if err != nil {
				ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
				os.Exit(1)
			}
			
			// Reload state
			state, _ = core.LoadState()
			engine = core.NewEngine(state, provider)
		}
		
		// Check if path is already tracked
		mapping := state.GetMapping(absPath)
		if mapping == nil {
			// First-time sync
			public, _ := cmd.Flags().GetBool("public")
			private, _ := cmd.Flags().GetBool("private")

			if public && private {
				ui.Error("PublicPrivateConflict", nil)
				os.Exit(1)
			}

			// Default to private if no flags are specified
			isPublic := public
			
			ui.Print("InitialSyncStart", map[string]interface{}{"Path": absPath})
			err = engine.InitialSyncWithVisibility(absPath, isPublic)
			if err != nil {
				ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
				os.Exit(1)
			}

			// Reload state to get the fresh mapping
			state, reloadErr := core.LoadState()
			if reloadErr != nil {
				ui.Warning("LoadStateFailed", map[string]interface{}{"Err": reloadErr})
			}
			if state != nil {
				mapping = state.GetMapping(absPath)
				if mapping != nil {
					visibility := "private"
					if mapping.Public {
						visibility = "public"
					}
					ui.Success("InitialSyncSuccess", map[string]interface{}{
						"Visibility": visibility,
						"Path":       absPath,
						"ID":         mapping.RemoteID,
					})
				}
			}
			return
		}

		// Regular sync for existing mapping
		info, err := os.Stat(path)
		if err != nil {
			ui.Error("StatPathFailed", map[string]interface{}{"Path": path, "Err": err})
			os.Exit(1)
		}

		var action core.SyncAction
		if info.IsDir() {
			action, err = engine.SyncDir(path)
		} else {
			action, err = engine.SyncFile(path)
		}

		if err != nil {
			if _, ok := err.(*core.ConflictError); ok {
				ui.Error("ConflictDetected", map[string]interface{}{"Err": err})
				os.Exit(1)
			}
			ui.Error("SyncFailedWithErr", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		mapping = state.GetMapping(absPath)
		switch action {
		case core.ActionNoop:
			ui.Print("SyncNoop", map[string]interface{}{"Path": absPath})
		case core.ActionPush:
			ui.Success("SyncPushed", map[string]interface{}{"Path": absPath})
		case core.ActionPull:
			ui.Success("SyncPulled", map[string]interface{}{"Path": absPath})
		}
	},
}

func init() {
	syncCmd.Flags().Bool("public", false, "Create a public gist (for initial sync)")
	syncCmd.Flags().Bool("private", false, "Create a private gist (default for initial sync)")
	syncCmd.Flags().String("provider", "", "Override the default provider (github, gitlab)")
	rootCmd.AddCommand(syncCmd)
}
