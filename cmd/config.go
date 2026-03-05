package cmd

import (
	"fmt"
	"strconv"

	"github.com/karanshah229/gistsync/internal"
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

		fmt.Printf("watch_interval_seconds: %d\n", config.WatchInterval)
		fmt.Printf("watch_debounce_ms:      %d\n", config.WatchDebounce)
		fmt.Printf("log_level:              %s\n", config.LogLevel)
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
			fmt.Println(config.WatchInterval)
		case "watch_debounce_ms":
			fmt.Println(config.WatchDebounce)
		case "log_level":
			fmt.Println(config.LogLevel)
		default:
			return fmt.Errorf("unknown config key: %s", key)
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

		fmt.Printf("Successfully set %s to %s\n", key, val)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
