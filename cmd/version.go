package cmd

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var versionFileContent string

// version is the primary version string, can be overwritten by ldflags
var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gistsync",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gistsync version %s\n", version)
	},
}

func init() {
	// Fallback to embedded VERSION file if not set by ldflags
	if version == "" {
		version = strings.TrimSpace(versionFileContent)
	}
	rootCmd.AddCommand(versionCmd)
}
