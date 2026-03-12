package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/karanshah229/gistsync/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Start a background watcher to automatically sync changes",
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			ui.Error("InitializationFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		w := watcher.NewWatcher(manager)
		if err := w.Start(); err != nil {
			ui.Error("WatcherFailedToStart", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
