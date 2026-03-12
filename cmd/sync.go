package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [path]",
	Short: "Sync a file or directory to a gist (creates a new gist if not already tracked)",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		providerName, _ := cmd.Flags().GetString("provider")
		
		if len(args) == 0 {
			if err := manager.SyncAll(providerName); err != nil {
				os.Exit(1)
			}
			return
		}

		path := args[0]
		gistID := ""
		if len(args) == 2 {
			gistID = args[1]
		}

		public, _ := cmd.Flags().GetBool("public")
		private, _ := cmd.Flags().GetBool("private")

		if public && private {
			ui.Error("PublicPrivateConflict", nil)
			os.Exit(1)
		}

		// SyncPath handles both new and existing mappings
		if err := manager.SyncPath(path, providerName, gistID, public); err != nil {
			ui.Error("SyncFailedWithErr", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
	},
}

func init() {
	syncCmd.Flags().Bool("public", false, "Create a public gist (for initial sync)")
	syncCmd.Flags().Bool("private", false, "Create a private gist (default for initial sync)")
	syncCmd.Flags().String("provider", "", "Override the default provider (github, gitlab)")
	rootCmd.AddCommand(syncCmd)
}
