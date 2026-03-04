package cmd

import (
	"fmt"
	"os"

	"github.com/karan/gistsync/core"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [path]",
	Short: "Stop tracking a file or directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		
		state, err := core.LoadState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
			os.Exit(1)
		}

		absPath, _ := core.GetAbsPath(path) // I should add a helper or just use filepath.Abs
		
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
			fmt.Printf("Path %s is not being tracked.\n", path)
			return
		}

		state.Mappings = newMappings
		if err := state.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving state: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Stopped tracking %s\n", path)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
