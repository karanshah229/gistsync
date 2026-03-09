package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var autostartCmd = &cobra.Command{
	Use:   "autostart",
	Short: "Manage autostart at login",
}

var autostartEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable gistsync autostart at login",
	Run: func(cmd *cobra.Command, args []string) {
		if err := internal.InstallAutostart(); err != nil {
			ui.Error("AutostartFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
		
		// Also update config
		config, _ := internal.LoadConfig()
		if config != nil {
			config.Autostart = true
			internal.SaveConfig(config)
		}
		
		ui.Success("AutostartEnabled", nil)
	},
}

var autostartDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable gistsync autostart at login",
	Run: func(cmd *cobra.Command, args []string) {
		if err := internal.UninstallAutostart(); err != nil {
			ui.Error("AutostartDisableFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
		
		// Also update config
		config, _ := internal.LoadConfig()
		if config != nil {
			config.Autostart = false
			internal.SaveConfig(config)
		}
		
		ui.Success("AutostartDisabled", nil)
	},
}

var autostartStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if autostart is enabled",
	Run: func(cmd *cobra.Command, args []string) {
		enabled, err := internal.IsAutostartEnabled()
		if err != nil {
			ui.Error("AutostartStatusCheckFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
		
		if enabled {
			ui.Success("AutostartStatusEnabled", nil)
		} else {
			ui.Error("AutostartStatusDisabled", nil)
		}
	},
}

func init() {
	autostartCmd.AddCommand(autostartEnableCmd)
	autostartCmd.AddCommand(autostartDisableCmd)
	autostartCmd.AddCommand(autostartStatusCmd)
	rootCmd.AddCommand(autostartCmd)
}
