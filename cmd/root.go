package cmd

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/internal/logger"
	"github.com/karanshah229/gistsync/pkg/i18n"
	"github.com/spf13/cobra"
)

// Version is the primary version string, can be overwritten by ldflags or set by main.go
var Version string

var rootCmd = &cobra.Command{
	Use:           "gistsync",
	Short:         "gistsync is a provider-agnostic file sync engine using GitHub Gists",
	Long:          `A fast and efficient CLI tool to sync local files and folders to GitHub Gists with 2-way hash-based change detection.`,
	Version:       Version,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("gistsync version {{.Version}}\n")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Initialize logger and reap old logs as early as possible
		cfg, _ := internal.LoadConfig()
		logLevel := "info"
		if cfg != nil {
			logLevel = cfg.LogLevel
		}

		if err := logger.Init(logLevel); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️ Warning: Failed to initialize logger: %v\n", err)
		} else {
			_ = logger.ReapLogs()
			logger.Log.Info("Process started", slog.String("cmd", cmd.Name()))
		}

		// Skip check for commands that don't need config
		skipCommands := map[string]struct{}{
			"init":       {},
			"help":       {},
			"version":    {},
			"completion": {},
			"provider":   {},
			"recover":    {},
		}

		for c := cmd; c != nil; c = c.Parent() {
			if _, ok := skipCommands[c.Name()]; ok {
				return nil
			}
		}

		// Also skip if no subcommand was provided (cmd is rootCmd)
		if cmd == rootCmd {
			return nil
		}

		// Strictly check if config and state are present and valid
		_, configErr := internal.LoadConfig()
		_, stateErr := core.LoadState()

		if configErr != nil || stateErr != nil {
			return fmt.Errorf("%s", i18n.T("ConfigMissingError", nil))
		}
		return nil
	}
}

// SetVersion allows main.go to inject the embedded VERSION content
func SetVersion(v string) {
	if Version == "" {
		Version = v
	}
	rootCmd.Version = Version
}
