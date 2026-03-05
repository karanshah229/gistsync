package cmd

import (
	_ "embed"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var versionFileContent string

// version is the primary version string, can be overwritten by ldflags
var version string

var rootCmd = &cobra.Command{
	Use:     "gistsync",
	Short:   "gistsync is a provider-agnostic file sync engine using GitHub Gists",
	Long:    `A fast and efficient CLI tool to sync local files and folders to GitHub Gists with 2-way hash-based change detection.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Fallback to embedded VERSION file if not set by ldflags
	if version == "" {
		version = strings.TrimSpace(versionFileContent)
	}
	rootCmd.Version = version
	// Custom template to match previous style if desired, or just use default
	rootCmd.SetVersionTemplate("gistsync version {{.Version}}\n")
}
