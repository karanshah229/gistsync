package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/providers"
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

		public, _ := cmd.Flags().GetBool("public")
		private, _ := cmd.Flags().GetBool("private")

		if public && private {
			fmt.Fprintf(os.Stderr, "Error: cannot specify both --public and --private\n")
			os.Exit(1)
		}

		// Default to private (isPublic = false) if no flags are specified.
		isPublic := public
		
		err = engine.InitialSyncWithVisibility(absPath, isPublic)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)
			os.Exit(1)
		}

		// Reload state to get the fresh mapping added by engine
		state, err = core.LoadState()
		if err != nil {
			fmt.Printf("Warning: Gist created but failed to reload state: %v\n", err)
			return
		}

		mapping := state.GetMapping(absPath)
		if mapping == nil {
			fmt.Println("Initialization successful, but could not retrieve mapping details.")
			return
		}

		visibility := "private"
		if mapping.Public {
			visibility = "public"
		}
		fmt.Printf("Initialized %s sync for %s (Gist ID: %s)\n", visibility, absPath, mapping.RemoteID)
	},
}

func init() {
	initCmd.Flags().Bool("public", false, "Create a public gist")
	initCmd.Flags().Bool("private", false, "Create a private gist (default)")
	rootCmd.AddCommand(initCmd)
}
