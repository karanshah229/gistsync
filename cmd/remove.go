package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [path]",
	Short: "Stop tracking a file or directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		if err := manager.RemovePath(args[0]); err != nil {
			ui.Error("RemovalError", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		ui.Success("RemovalSuccess", map[string]interface{}{"Path": args[0]})
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
