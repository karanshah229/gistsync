package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/karan/gistsync/core"
	"github.com/karan/gistsync/providers"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a file or directory for syncing (alias for initial sync)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		absPath, _ := filepath.Abs(path)

		state, err := core.LoadState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
			os.Exit(1)
		}

		if state.GetMapping(absPath) != nil {
			fmt.Printf("Path %s is already being tracked.\n", absPath)
			return
		}

		provider := providers.NewGitHubProvider()
		engine := core.NewEngine(state, provider)

		// Call the internal initialSync (I might need to export it or just use SyncFile/SyncDir)
		// Since SyncFile/SyncDir call initialSync if no mapping exists, we can just call them.
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if info.IsDir() {
			err = engine.SyncDir(path)
		} else {
			err = engine.SyncFile(path)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)
			os.Exit(1)
		}

		mapping := state.GetMapping(absPath)
		fmt.Printf("Initialized sync for %s (Gist ID: %s)\n", absPath, mapping.RemoteID)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
