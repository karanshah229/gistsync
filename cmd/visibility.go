package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/pkg/ui"
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
			ui.Error("PublicPrivateConflict", nil)
			os.Exit(1)
		}

		if !publicFlag && !privateFlag {
			ui.Error("PublicPrivateMissing", nil)
			os.Exit(1)
		}

		isPublic := publicFlag

		state, err := core.LoadState()
		if err != nil {
			ui.Error("LoadStateFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		provider := providers.NewGitHubProvider()
		engine := core.NewEngine(state, provider)

		targetName := "private"
		if isPublic {
			targetName = "public"
		}

		ui.Print("ChangingVisibility", map[string]interface{}{"Path": path, "Visibility": targetName})
		if !isPublic {
			ui.Info("PrivateConversionNote", nil)
		}

		if err := engine.SetVisibility(path, isPublic); err != nil {
			ui.Error("VisibilityChangeError", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		ui.Success("VisibilityChangeSuccess", map[string]interface{}{"Visibility": targetName})
	},
}

func init() {
	visibilityCmd.Flags().Bool("public", false, "Make the gist public")
	visibilityCmd.Flags().Bool("private", false, "Make the gist private")
	rootCmd.AddCommand(visibilityCmd)
}
