package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/karanshah229/gistsync/providers"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage and test sync providers",
}

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub provider commands",
}

var githubTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test GitHub connection and authentication",
	Run: func(cmd *cobra.Command, args []string) {
		p := providers.NewGitHubProvider()
		ui.Print("TestingProvider", map[string]interface{}{"Name": "GitHub"})
		testProvider(p)
	},
}

var gitlabCmd = &cobra.Command{
	Use:   "gitlab",
	Short: "GitLab provider commands",
}

var gitlabTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test GitLab connection and authentication",
	Run: func(cmd *cobra.Command, args []string) {
		p := providers.NewGitLabProvider()
		ui.Print("TestingProvider", map[string]interface{}{"Name": "GitLab"})
		testProvider(p)
	},
}

func testProvider(p core.Provider) {
	ok, msg, err := p.Verify()
	if err != nil {
		ui.Error("VerificationError", map[string]interface{}{"Err": err})
		os.Exit(1)
	}

	if ok {
		ui.Success("VerificationSuccess", map[string]interface{}{"Msg": msg})
	} else {
		ui.Error("VerificationFailure", map[string]interface{}{"Msg": msg})
		os.Exit(1)
	}

	// Also check rate limit
	remaining, reset, err := p.CheckRateLimit()
	if err == nil {
		ui.Print("RateLimitInfo", map[string]interface{}{
			"Remaining": remaining,
			"Reset":     reset.Format("15:04:05"),
		})
	}
}

var providerInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show setup and authentication information for providers",
	Run: func(cmd *cobra.Command, args []string) {
		showProviderInfo()
	},
}

func showProviderInfo() {
	ui.Header("ProviderInfoTitle", nil)
	ui.Print("GitHubInfo", nil)
	ui.Print("GitLabInfo", nil)
}

func init() {
	githubCmd.AddCommand(githubTestCmd)
	gitlabCmd.AddCommand(gitlabTestCmd)
	
	providerCmd.AddCommand(githubCmd)
	providerCmd.AddCommand(gitlabCmd)
	providerCmd.AddCommand(providerInfoCmd)
	
	rootCmd.AddCommand(providerCmd)
}
