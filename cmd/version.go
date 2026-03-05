package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev-concurrency-safe"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gistsync",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gistsync version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
