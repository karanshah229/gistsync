package cmd

import (
	"fmt"
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/spf13/cobra"
)

var repairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair configuration paths (useful after restoring on a different OS)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔧 Repairing configuration paths...")

		state, err := core.LoadState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error loading state: %v\n", err)
			os.Exit(1)
		}

		results, err := internal.RepairConfig(state)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error during repair: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("✅ No mappings found to repair.")
			return
		}

		fmt.Println("\n📋 Repair Results:")
		repairedCount := 0
		for _, res := range results {
			statusIcon := "❓"
			switch res.Status {
			case "VALID":
				statusIcon = "✅"
			case "REPAIRED":
				statusIcon = "🛠️"
				repairedCount++
			case "MISSING":
				statusIcon = "❌"
			case "SKIPPED":
				statusIcon = "⏭️"
			}

			fmt.Printf("  %s %s\n", statusIcon, res.OldPath)
			if res.Status == "REPAIRED" {
				fmt.Printf("     -> %s\n", res.NewPath)
			} else if res.Status == "MISSING" {
				fmt.Printf("     ⚠️  File not found at: %s\n", res.NewPath)
			}
		}

		if repairedCount > 0 {
			fmt.Printf("\n🎉 successfully repaired %d path(s)!\n", repairedCount)
		} else {
			fmt.Println("\n✨ No paths were repaired.")
		}
	},
}

func init() {
	configCmd.AddCommand(repairCmd)
}
