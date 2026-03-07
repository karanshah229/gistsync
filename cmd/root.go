package cmd

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/spf13/cobra"
)

// Version is the primary version string, can be overwritten by ldflags or set by main.go
var Version string

var rootCmd = &cobra.Command{
	Use:     "gistsync",
	Short:   "gistsync is a provider-agnostic file sync engine using GitHub Gists",
	Long:    `A fast and efficient CLI tool to sync local files and folders to GitHub Gists with 2-way hash-based change detection.`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("gistsync version {{.Version}}\n")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip check for commands that don't need config
		if cmd.Name() == "init" || cmd.Name() == "help" || cmd.Name() == "version" || cmd.Name() == "completion" {
			return nil
		}

		// Strictly check if config and state are present and valid
		_, configErr := internal.LoadConfig()
		_, stateErr := core.LoadState()

		if configErr != nil || stateErr != nil {
			return fmt.Errorf("configuration or state is missing or malformed. Please run 'gistsync init' to set up the tool")
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
