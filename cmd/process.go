package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Manage running gistsync processes",
}

var processListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all running gistsync processes",
	Run: func(cmd *cobra.Command, args []string) {
		procs, err := internal.ListProcesses()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to list processes: %v\n", err)
			os.Exit(1)
		}

		if len(procs) == 0 {
			ui.Print("NoProcessesFound", nil)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PID\tPPID\tSTART TIME\tCPU%\tMEM (RSS)\tCOMMAND")
		for _, p := range procs {
			memMB := fmt.Sprintf("%.2f MB", float64(p.Memory)/1024/1024)
			fmt.Fprintf(w, "%d\t%d\t%s\t%.1f%%\t%s\t%s\n",
				p.PID, p.PPID, p.StartTime.Format("2006-01-02 15:04:05"), p.CPU, memMB, p.Cmdline)
		}
		w.Flush()
	},
}

var processKillOthersCmd = &cobra.Command{
	Use:   "kill-others",
	Short: "Kill all running gistsync processes except the current one",
	Run: func(cmd *cobra.Command, args []string) {
		killed, err := internal.KillOtherProcesses()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to kill other processes: %v\n", err)
			os.Exit(1)
		}

		if killed == 0 {
			ui.Print("NoOtherProcessesRunning", nil)
		} else {
			ui.Success("KilledOtherProcesses", map[string]interface{}{"Count": killed})
		}
	},
}

func init() {
	processCmd.AddCommand(processListCmd)
	processCmd.AddCommand(processKillOthersCmd)
	rootCmd.AddCommand(processCmd)
}
