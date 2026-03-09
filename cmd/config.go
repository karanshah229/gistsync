package cmd

import (
	"fmt"
	"strconv"

	"github.com/karanshah229/gistsync/internal"
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
		config, err := internal.LoadConfig()
		if err != nil {
			return err
		}

		ui.Print("ConfigList", map[string]interface{}{
			"Interval": config.WatchInterval,
			"Debounce": config.WatchDebounce,
			"Level":    config.LogLevel,
		})
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a specific configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := internal.LoadConfig()
		if err != nil {
			return err
		}

		key := args[0]
		switch key {
		case "watch_interval_seconds":
			ui.Print("ConfigVal", map[string]interface{}{"Val": config.WatchInterval})
		case "watch_debounce_ms":
			ui.Print("ConfigVal", map[string]interface{}{"Val": config.WatchDebounce})
		case "log_level":
			ui.Print("ConfigVal", map[string]interface{}{"Val": config.LogLevel})
		default:
			return fmt.Errorf("%s", i18n.T("UnknownConfigKey", map[string]interface{}{"Key": key}))
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := internal.LoadConfig()
		if err != nil {
			return err
		}

		key := args[0]
		val := args[1]

		switch key {
		case "watch_interval_seconds":
			v, err := strconv.Atoi(val)
			if err != nil || v <= 0 {
				return fmt.Errorf("value must be a positive integer")
			}
			config.WatchInterval = v
		case "watch_debounce_ms":
			v, err := strconv.Atoi(val)
			if err != nil || v <= 0 {
				return fmt.Errorf("value must be a positive integer")
			}
			config.WatchDebounce = v
		case "log_level":
			if val != "info" && val != "debug" && val != "error" {
				return fmt.Errorf("invalid log level: %s (choose info, debug, or error)", val)
			}
			config.LogLevel = val
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		if err := internal.SaveConfig(config); err != nil {
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
	rootCmd.AddCommand(configCmd)
}
