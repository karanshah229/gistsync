package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/karanshah229/gistsync/providers"
	"github.com/karanshah229/gistsync/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Start a background watcher to automatically sync changes",
	Run: func(cmd *cobra.Command, args []string) {
		state, err := core.LoadState()
		if err != nil {
			ui.Error("LoadStateFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		provider := providers.NewGitHubProvider()
		engine := core.NewEngine(state, provider)

		config, err := internal.LoadConfig()
		if err != nil {
			ui.Error("LoadConfigFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		w := watcher.NewWatcher(engine, config)
		if err := w.Start(); err != nil {
			ui.Error("WatcherFailedToStart", map[string]interface{}{"Err": err})
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
