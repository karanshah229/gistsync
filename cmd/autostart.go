package cmd

import (
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/internal"
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
			fmt.Fprintf(os.Stderr, "❌ Failed to enable autostart: %v\n", err)
			os.Exit(1)
		}
		
		// Also update config
		config, _ := internal.LoadConfig()
		if config != nil {
			config.Autostart = true
			internal.SaveConfig(config)
		}
		
		fmt.Println("✅ Autostart enabled successfully!")
	},
}

var autostartDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable gistsync autostart at login",
	Run: func(cmd *cobra.Command, args []string) {
		if err := internal.UninstallAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to disable autostart: %v\n", err)
			os.Exit(1)
		}
		
		// Also update config
		config, _ := internal.LoadConfig()
		if config != nil {
			config.Autostart = false
			internal.SaveConfig(config)
		}
		
		fmt.Println("✅ Autostart disabled successfully!")
	},
}

var autostartStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if autostart is enabled",
	Run: func(cmd *cobra.Command, args []string) {
		enabled, err := internal.IsAutostartEnabled()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to check autostart status: %v\n", err)
			os.Exit(1)
		}
		
		if enabled {
			fmt.Println("✅ Autostart is currently ENABLED")
		} else {
			fmt.Println("❌ Autostart is currently DISABLED")
		}
	},
}

func init() {
	autostartCmd.AddCommand(autostartEnableCmd)
	autostartCmd.AddCommand(autostartDisableCmd)
	autostartCmd.AddCommand(autostartStatusCmd)
	rootCmd.AddCommand(autostartCmd)
}
