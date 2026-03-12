package cmd

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gistsync configurations and state",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if internal.IsConfigPresent() {
			if !ui.Confirm("ConfirmQuestion", map[string]interface{}{"Message": "⚠️  Configuration already exists. Overwrite?"}) {
				ui.Print("Aborted", nil)
				return
			}
		}

		ui.Print("Initializing", nil)

		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		// 1. Check Providers
		ui.Header("CheckingProviders", nil)
		gh, _ := manager.GetProvider("github")
		gl, _ := manager.GetProvider("gitlab")

		ghOk := false
		var ghMsg string
		if gh != nil {
			ghOk, ghMsg, _ = gh.Verify()
		}
		
		glOk := false
		var glMsg string
		if gl != nil {
			glOk, glMsg, _ = gl.Verify()
		}

		connectedProviders := []string{}
		if ghOk {
			ui.Success("GitHubConnected", map[string]interface{}{"Msg": strings.TrimSpace(ghMsg)})
			connectedProviders = append(connectedProviders, "github")
		} else {
			ui.Error("GitHubNotConnected", map[string]interface{}{"Msg": strings.TrimSpace(ghMsg)})
		}

		if glOk {
			ui.Success("GitLabConnected", map[string]interface{}{"Msg": strings.TrimSpace(glMsg)})
			connectedProviders = append(connectedProviders, "gitlab")
		} else {
			ui.Error("GitLabNotConnected", map[string]interface{}{"Msg": strings.TrimSpace(glMsg)})
		}

		if len(connectedProviders) == 0 {
			ui.Info("NoProvidersConnected", nil)
			showProviderInfo()
			return
		}

		// 2. Optional Restore
		wantRestore := ui.Confirm("ConfirmQuestion", map[string]interface{}{"Message": "Would you like to restore configurations from a provider?"})

		if wantRestore {
			var selectedRestoreProvider string
			if len(connectedProviders) == 1 {
				selectedRestoreProvider = connectedProviders[0]
			} else {
				pSelect := &survey.Select{
					Message: "Select provider to restore from:",
					Options: connectedProviders,
				}
				survey.AskOne(pSelect, &selectedRestoreProvider)
				if selectedRestoreProvider == "" {
					ui.Print("Aborted", nil)
					return
				}
			}

			backups, err := manager.ListBackups(selectedRestoreProvider)
			if err != nil {
				ui.Error("RestorationFailed", map[string]interface{}{"Err": err})
			} else if len(backups) == 0 {
				ui.Warning("NoBackupFound", nil)
			} else {
				// For now, use the first backup or implement selection
				// To maintain original behavior, we can let user select if multiple
				selectedID := backups[0].ID
				if len(backups) > 1 {
					options := []string{}
					idMap := make(map[string]string)
					for _, b := range backups {
						label := fmt.Sprintf("%s (Updated: %v)", b.ID, b.UpdatedAt.Format("2006-01-02 15:04:05"))
						options = append(options, label)
						idMap[label] = b.ID
					}
					prompt := &survey.Select{
						Message: "Multiple backups found. Select one to restore:",
						Options: options,
					}
					var selection string
					survey.AskOne(prompt, &selection)
					selectedID = idMap[selection]
				}

				if err := manager.RestoreConfig(selectedRestoreProvider, selectedID); err != nil {
					ui.Error("RestorationFailed", map[string]interface{}{"Err": err})
				} else {
					ui.Success("RestorationSuccess", nil)
					
					// Auto-repair paths
					state, _ := manager.Repo.Load()
					if state != nil {
						results, _, err := internal.RepairConfig(state)
						if err == nil && len(results) > 0 {
							repairedCount := 0
							missingCount := 0
							for _, r := range results {
								if r.Status == "REPAIRED" {
									repairedCount++
								} else if r.Status == "MISSING" {
									missingCount++
								}
							}
							if repairedCount > 0 || missingCount > 0 {
								ui.Print("AutoRepairedPaths", map[string]interface{}{
									"Repaired": repairedCount,
									"Missing":  missingCount,
								})
							}
						}
					}

					wantSync := ui.Confirm("ConfirmQuestion", map[string]interface{}{"Message": "Would you like to run sync now?"})
					if wantSync {
						manager.SyncAll(selectedRestoreProvider)
					}
					ui.Success("Ready", nil)
					return
				}
			}
		}

		config := internal.DefaultConfig()
		restored := false // We can use this to skip backup if we restored
		// (The previous block returned early if restored, so if we're here, restored is false or we didn't restore)
		reader := ui.GetSharedReader()

		if ghOk {
			config.Providers["github"] = "connected"
		} else {
			config.Providers["github"] = "not connected"
		}
		if glOk {
			config.Providers["gitlab"] = "connected"
		} else {
			config.Providers["gitlab"] = "not connected"
		}

		// 3. Select Default Provider
		ui.Header("DefaultProviderTitle", nil)
		ui.Print("DefaultProviderUsage", nil)

		var selectedProvider string
		prompt := &survey.Select{
			Message: "Select your default provider:",
			Options: connectedProviders,
			Default: "github",
		}

		if err := survey.AskOne(prompt, &selectedProvider); err != nil {
			ui.Error("SelectionFailed", map[string]interface{}{"Err": err})
			selectedProvider = "github"
		}
		config.DefaultProvider = selectedProvider
		ui.Success("SelectedProvider", map[string]interface{}{"Provider": selectedProvider})

		// 4. Interactive Config
		ui.Header("ConfigSetupTitle", nil)
		options := internal.GetConfigOptions()
		for _, opt := range options {
			for {
				ui.Printf("ConfigPrompt", map[string]interface{}{"Prompt": opt.Prompt, "Default": opt.Default})
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				if input == "" {
					// Use default
					setField(config, opt.Key, opt.Default)
					break
				}

				if input == "?" || strings.ToLower(input) == "help" {
					ui.Info("ConfigHelp", map[string]interface{}{"Description": opt.Description})
					continue
				}

				// Try to parse input
				if err := updateField(config, opt.Key, input); err != nil {
					ui.Error("InvalidInput", map[string]interface{}{"Err": err})
					continue
				}
				break
			}
		}

		// 5. Save Config
		if err := internal.SaveConfig(config); err != nil {
			ui.Error("SaveConfigFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
		ui.Success("ConfigSaved", nil)
		if config.Autostart {
			ui.Print("EnablingAutostart", nil)
			if err := internal.InstallAutostart(); err != nil {
				ui.Warning("AutostartFailed", map[string]interface{}{"Err": err})
			} else {
				ui.Success("AutostartEnabled", nil)
			}
		}

		// 6. Initialize state.json
		if err := manager.InitializeState(); err != nil {
			ui.Error("StateInitFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
		ui.Success("StateInitialized", nil)

		// 7. Optional Backup
		if !restored {
			if ui.Confirm("ConfirmQuestion", map[string]interface{}{"Message": "Would you like to backup your configuration to the default provider?"}) {
				if err := manager.BackupConfig(config.DefaultProvider); err != nil {
					ui.Error("BackupFailed", map[string]interface{}{"Err": err})
				} else {
					ui.Success("BackupSuccess", nil)
				}
			}
		}

		ui.Success("ReadyWithHint", nil)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// Helper to set field via reflection or simple switch
func setField(config *internal.Config, key string, val interface{}) {
	switch key {
	case "WatchInterval":
		config.WatchInterval = val.(int)
	case "WatchDebounce":
		config.WatchDebounce = val.(int)
	case "LogLevel":
		config.LogLevel = val.(string)
	case "Autostart":
		config.Autostart = val.(bool)
	}
}

// Helper to update field from string input
func updateField(config *internal.Config, key string, input string) error {
	switch key {
	case "WatchInterval":
		v, err := strconv.Atoi(input)
		if err != nil {
			return err
		}
		config.WatchInterval = v
	case "WatchDebounce":
		v, err := strconv.Atoi(input)
		if err != nil {
			return err
		}
		config.WatchDebounce = v
	case "LogLevel":
		allowed := []string{"debug", "info", "warn", "error"}
		if !slices.Contains(allowed, strings.ToLower(input)) {
			return fmt.Errorf("must be one of: %s", strings.Join(allowed, ", "))
		}
		config.LogLevel = strings.ToLower(input)
	case "Autostart":
		v, err := strconv.ParseBool(input)
		if err != nil {
			return err
		}
		config.Autostart = v
	}
	return nil
}
