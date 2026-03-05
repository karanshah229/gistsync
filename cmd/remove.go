package cmd

import (
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [path]",
	Short: "Stop tracking a file or directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		
		absPath, _ := core.GetAbsPath(path) // I should add a helper or just use filepath.Abs
		
		err := core.WithLock(func(state *core.State) error {
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
			return nil // WithLock will call Save()
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Stopped tracking %s\n", path)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
