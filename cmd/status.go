package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/karanshah229/gistsync/core"
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
			fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
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
			fmt.Println("No files are being tracked.")
			return
		}

		for _, p := range paths {
			status, err := engine.Status(p)
			if err != nil {
				fmt.Printf("%s: ERROR (%v)\n", p, err)
				continue
			}
			
			abs, _ := filepath.Abs(p)
			fmt.Printf("%s: %s\n", abs, status)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
