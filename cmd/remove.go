package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [path]",
	Short: "Stop tracking a file or directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		
		absPath, _ := filepath.Abs(path)
		
		state, err := core.LoadState()
		if err != nil {
			ui.Error("LoadStateFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		err = state.WithLock(func(state *core.State) error {
			newMappings := []core.Mapping{}
			found := false
			for _, m := range state.Mappings {
				if m.LocalPath == absPath {
					found = true
					continue
				}
				newMappings = append(newMappings, m)
			}

			if !found {
				return fmt.Errorf("path %s is not being tracked", path)
			}

			state.Mappings = newMappings
			return nil
		})

		if err != nil {
			ui.Error("RemovalError", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		ui.Success("RemovalSuccess", map[string]interface{}{"Path": path})
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
