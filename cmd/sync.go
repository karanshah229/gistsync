package cmd

import (
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/providers"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [path]",
	Short: "Sync a file or directory to a gist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		
		state, err := core.LoadState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
			os.Exit(1)
		}

		provider := providers.NewGitHubProvider()
		engine := core.NewEngine(state, provider)

		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error stating path: %v\n", err)
			os.Exit(1)
		}

		if info.IsDir() {
			err = engine.SyncDir(path)
		} else {
			err = engine.SyncFile(path)
		}

		if err != nil {
			if _, ok := err.(*core.ConflictError); ok {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Sync failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Sync successful!")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
