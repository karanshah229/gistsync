package cmd

import (
	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/spf13/cobra"
)

var configSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync your configuration folder to a gist provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := sync.NewSyncManager(Version)
		if err != nil {
			return err
		}

		providerName, _ := cmd.Flags().GetString("provider")
		return manager.ConfigSync(providerName)
	},
}

func init() {
	configSyncCmd.Flags().String("provider", "", "Provider to use for initial sync (github, gitlab)")
}
