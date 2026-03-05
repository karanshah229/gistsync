package cmd

import (
	_ "embed"
	"os"

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
}

// SetVersion allows main.go to inject the embedded VERSION content
func SetVersion(v string) {
	if Version == "" {
		Version = v
	}
	rootCmd.Version = Version
}
