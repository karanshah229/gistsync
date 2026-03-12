package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var visibilityCmd = &cobra.Command{
	Use:   "visibility [path] [public|private]",
	Short: "Change the visibility of a gist",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		path := args[0]
		visibility := args[1]

		var public bool
		switch visibility {
			case "public":
				public = true
			case "private":
				public = false
			default:
				ui.Error("InvalidVisibility", map[string]interface{}{"Value": visibility})
				os.Exit(1)
		}

		if err := manager.SetVisibility(path, public, ""); err != nil {
			ui.Error("VisibilityChangeError", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(visibilityCmd)
}
