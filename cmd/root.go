package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gistsync",
	Short: "gistsync is a provider-agnostic file sync engine using GitHub Gists",
	Long:  `A fast and efficient CLI tool to sync local files and folders to GitHub Gists with 2-way hash-based change detection.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Root flags if any
}
