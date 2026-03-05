package cmd

import (
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/providers"
	"github.com/spf13/cobra"
)

var visibilityCmd = &cobra.Command{
	Use:   "visibility [path]",
	Short: "Change the visibility of a tracked path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		publicFlag, _ := cmd.Flags().GetBool("public")
		privateFlag, _ := cmd.Flags().GetBool("private")

		if publicFlag && privateFlag {
			fmt.Println("Error: cannot specify both --public and --private")
			os.Exit(1)
		}

		if !publicFlag && !privateFlag {
			fmt.Println("Error: must specify either --public or --private")
			os.Exit(1)
		}

		isPublic := publicFlag

		state, err := core.LoadState()
		if err != nil {
			fmt.Printf("Error loading state: %v\n", err)
			os.Exit(1)
		}

		provider := providers.NewGitHubProvider()
		engine := core.NewEngine(state, provider)

		targetName := "private"
		if isPublic {
			targetName = "public"
		}

		fmt.Printf("Changing visibility of %s to %s...\n", path, targetName)
		if !isPublic {
			fmt.Println("Note: Converting Public to Private requires recreating the gist. The Gist ID will change.")
		}

		if err := engine.SetVisibility(path, isPublic); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully changed visibility to %s\n", targetName)
	},
}

func init() {
	visibilityCmd.Flags().Bool("public", false, "Make the gist public")
	visibilityCmd.Flags().Bool("private", false, "Make the gist private")
	rootCmd.AddCommand(visibilityCmd)
}
