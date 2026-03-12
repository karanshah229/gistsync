package cmd

import (
	"fmt"
	"strconv"

	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/karanshah229/gistsync/pkg/i18n"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gistsync user configurations",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			return err
		}

		ui.Print("ConfigList", map[string]interface{}{
			"Interval": manager.Config.WatchInterval,
			"Debounce": manager.Config.WatchDebounce,
			"Level":    manager.Config.LogLevel,
			"Provider": manager.Config.DefaultProvider,
			"Autostart": manager.Config.Autostart,
		})
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a specific configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			return err
		}

		key := args[0]
		switch key {
		case "watch_interval_seconds":
			ui.Print("ConfigVal", map[string]interface{}{"Val": manager.Config.WatchInterval})
		case "watch_debounce_ms":
			ui.Print("ConfigVal", map[string]interface{}{"Val": manager.Config.WatchDebounce})
		case "log_level":
			ui.Print("ConfigVal", map[string]interface{}{"Val": manager.Config.LogLevel})
		default:
			return fmt.Errorf("%s", i18n.T("UnknownConfigKeyHint", map[string]interface{}{"Key": key}))
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			return err
		}

		key := args[0]
		val := args[1]

		switch key {
		case "watch_interval_seconds":
			v, err := strconv.Atoi(val)
			if err != nil || v <= 0 {
				return fmt.Errorf("%s", i18n.T("InvalidPositiveInt", map[string]interface{}{"Key": key, "Val": val}))
			}
			manager.Config.WatchInterval = v
		case "watch_debounce_ms":
			v, err := strconv.Atoi(val)
			if err != nil || v <= 0 {
				return fmt.Errorf("%s", i18n.T("InvalidPositiveInt", map[string]interface{}{"Key": key, "Val": val}))
			}
			manager.Config.WatchDebounce = v
		case "log_level":
			if val != "info" && val != "debug" && val != "warn" && val != "error" {
				return fmt.Errorf("%s", i18n.T("InvalidLogLevel", map[string]interface{}{"Val": val}))
			}
			manager.Config.LogLevel = val
		default:
			return fmt.Errorf("%s", i18n.T("UnknownConfigKeyHint", map[string]interface{}{"Key": key}))
		}

		if err := internal.SaveConfig(manager.Config); err != nil {
			return err
		}

		ui.Success("ConfigSetSuccess", map[string]interface{}{"Key": key, "Val": val})
		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configSyncCmd)
	rootCmd.AddCommand(configCmd)
}
