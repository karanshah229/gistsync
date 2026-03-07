package cmd

import (
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/core"
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
		fmt.Println("🔍 Testing Provider: GitHub")
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
		fmt.Println("🔍 Testing Provider: GitLab")
		testProvider(p)
	},
}

func testProvider(p core.Provider) {
	ok, msg, err := p.Verify()
	if err != nil {
		fmt.Printf("❌ Error during verification: %v\n", err)
		os.Exit(1)
	}

	if ok {
		fmt.Printf("✅ Success: %s\n", msg)
	} else {
		fmt.Printf("❌ Failure: %s\n", msg)
		os.Exit(1)
	}

	// Also check rate limit
	remaining, reset, err := p.CheckRateLimit()
	if err == nil {
		fmt.Printf("📊 Rate Limit: %d remaining (resets at %v)\n", remaining, reset.Format("15:04:05"))
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
	fmt.Println("ℹ️  Provider Setup Information")
	fmt.Println("\n🔹 GitHub")
	fmt.Println("   1. Install GitHub CLI: https://cli.github.com/")
	fmt.Println("   2. Authenticate: run 'gh auth login'")
	fmt.Println("   3. Verify: run 'gistsync provider github test'")

	fmt.Println("\n🔹 GitLab")
	fmt.Println("   (Implementation pending. Support for GitLab personal access tokens coming soon.)")
	fmt.Println("   3. Verify: run 'gistsync provider gitlab test'")
}

func init() {
	githubCmd.AddCommand(githubTestCmd)
	gitlabCmd.AddCommand(gitlabTestCmd)
	
	providerCmd.AddCommand(githubCmd)
	providerCmd.AddCommand(gitlabCmd)
	providerCmd.AddCommand(providerInfoCmd)
	
	rootCmd.AddCommand(providerCmd)
}
