package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Show the sync status of a file or directory",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		path := ""
		if len(args) > 0 {
			path = args[0]
		}

		if err := manager.Status(path, ""); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
