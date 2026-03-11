package cmd

import (
	"os"
	"path/filepath"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/karanshah229/gistsync/providers"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Show the sync status of a file or directory",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		state, err := core.LoadState()
		if err != nil {
			ui.Error("LoadStateFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		provider := providers.NewGitHubProvider()
		engine := core.NewEngine(state, provider)

		var paths []string
		if len(args) > 0 {
			paths = append(paths, args[0])
		} else {
			// Show status for all mappings
			for _, m := range state.Mappings {
				paths = append(paths, m.LocalPath)
			}
		}

		if len(paths) == 0 {
			ui.Print("NoFilesTracked", nil)
			return
		}

		for _, p := range paths {
			status, err := engine.Status(p)
			if err != nil {
				ui.Print("StatusError", map[string]interface{}{"Path": p, "Err": err})
				continue
			}
			
			abs, _ := filepath.Abs(p)
			ui.Print("StatusLine", map[string]interface{}{"Path": abs, "Status": status})
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
