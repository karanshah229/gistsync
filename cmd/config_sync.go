package cmd

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/internal/storage"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/karanshah229/gistsync/providers"
	"github.com/spf13/cobra"
)

var configSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync your configuration folder to a gist provider",
	Run: func(cmd *cobra.Command, args []string) {
		configDir, err := storage.GetConfigDir()
		if err != nil {
			ui.Error("ConfigDirFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		state, err := core.LoadState()
		if err != nil {
			ui.Error("LoadStateFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		config, err := internal.LoadConfig()
		if err != nil {
			ui.Error("LoadConfigFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		mapping := state.GetMapping(configDir)
		var providerName string

		// 1. Resolve Provider
		if mapping != nil {
			providerName = mapping.Provider
		} else {
			// Not synced yet, resolve provider from Flag -> Config -> Prompt
			providerFlag, _ := cmd.Flags().GetString("provider")
			if providerFlag != "" {
				providerName = providerFlag
			} else if config.DefaultProvider != "" {
				providerName = config.DefaultProvider
			} else {
				// Prompt for provider
				gh := providers.NewGitHubProvider()
				gl := providers.NewGitLabProvider()

				ghOk, _, _ := gh.Verify()
				glOk, _, _ := gl.Verify()

				connected := []string{}
				if ghOk {
					connected = append(connected, "github")
				}
				if glOk {
					connected = append(connected, "gitlab")
				}

				if len(connected) == 0 {
					ui.Error("NoProvidersConnected", nil)
					return
				}

				if len(connected) == 1 {
					providerName = connected[0]
				} else {
					pSelect := &survey.Select{
						Message: "Select provider to sync config with:",
						Options: connected,
					}
					if err := survey.AskOne(pSelect, &providerName); err != nil {
						ui.Error("SelectionFailed", map[string]interface{}{"Err": err})
						return
					}
				}
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

		// 2. Perform Sync
		if mapping == nil {
			ui.Print("InitialSyncStart", map[string]interface{}{"Path": configDir})
			// Always private for config sync
			err = engine.InitialSyncWithVisibility(configDir, false)
			if err != nil {
				ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
				os.Exit(1)
			}

			// Reload state
			state, _ = core.LoadState()
			mapping = state.GetMapping(configDir)
			if mapping != nil {
				ui.Success("InitialSyncSuccess", map[string]interface{}{
					"Visibility": "private",
					"Path":       configDir,
					"ID":         mapping.RemoteID,
				})
			}
		} else {
			action, err := engine.SyncDir(configDir)
			if err != nil {
				if _, ok := err.(*core.ConflictError); ok {
					ui.Error("ConflictDetected", map[string]interface{}{"Err": err})
					os.Exit(1)
				}
				ui.Error("SyncFailedWithErr", map[string]interface{}{"Err": err})
				os.Exit(1)
			}

			switch action {
			case core.ActionNoop:
				ui.Print("SyncNoop", map[string]interface{}{"Path": configDir})
			case core.ActionPush:
				ui.Success("SyncPushed", map[string]interface{}{"Path": configDir})
			case core.ActionPull:
				ui.Success("SyncPulled", map[string]interface{}{"Path": configDir})
			}
		}
	},
}

func init() {
	configSyncCmd.Flags().String("provider", "", "Provider to use for initial sync (github, gitlab)")
}
