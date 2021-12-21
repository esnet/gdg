package cmd

import (
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var alertnotifications = &cobra.Command{
	Use:     "alertnotifications",
	Aliases: []string{"an", "alertnotification"},
	Short:   "Manage alert notification channels",
	Long:    `Manage alert notification channels`,
}

func init() {
	rootCmd.AddCommand(alertnotifications)
}
