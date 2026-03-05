package cmd

import (
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/providers"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage and test sync providers",
}

var providerTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the current provider connection and authentication",
	Run: func(cmd *cobra.Command, args []string) {
		// MVP: Hardcoded to GitHub for now as per current project state
		p := providers.NewGitHubProvider()
		
		fmt.Println("🔍 Testing Provider: GitHub")
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
	},
}

func init() {
	providerCmd.AddCommand(providerTestCmd)
	rootCmd.AddCommand(providerCmd)
}
