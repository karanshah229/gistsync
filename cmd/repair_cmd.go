package cmd

import (
	"os"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var repairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair configuration paths (useful after restoring on a different OS)",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Print("RepairingPaths", nil)

		state, err := core.LoadState()
		if err != nil {
			ui.Error("LoadStateFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		results, err := internal.RepairConfig(state)
		if err != nil {
			ui.Error("RepairFailed", map[string]interface{}{"Err": err})
			os.Exit(1)
		}

		if len(results) == 0 {
			ui.Success("NoRepairNeeded", nil)
			return
		}

		ui.Header("RepairResultsTitle", nil)
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

			ui.Printf("RepairResultLine", map[string]interface{}{"Status": statusIcon, "Path": res.OldPath})
			if res.Status == "REPAIRED" {
				ui.Printf("RepairResultTo", map[string]interface{}{"Path": res.NewPath})
			} else if res.Status == "MISSING" {
				ui.Printf("RepairResultMissing", map[string]interface{}{"Path": res.NewPath})
			}
		}

		if repairedCount > 0 {
			ui.Success("RepairSummarySuccess", map[string]interface{}{"Count": repairedCount})
		} else {
			ui.Info("RepairSummaryNone", nil)
		}
	},
}

func init() {
	configCmd.AddCommand(repairCmd)
}
