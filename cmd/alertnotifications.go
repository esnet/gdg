package cmd

import (
	"github.com/netsage-project/gdg/api"
	"github.com/spf13/cobra"
)

func getAlertNotificationsGlobalFlags(cmd *cobra.Command) api.Filter {
	alertNotificationFilter, _ := cmd.Flags().GetString("alertnotification")

	filters := api.AlertNotificationFilter{}
	filters.Init()
	filters.AddFilter("Name", alertNotificationFilter)

	return filters
}

// versionCmd represents the version command
var alertnotifications = &cobra.Command{
	Use:     "alertnotifications",
	Aliases: []string{"an", "alertnotification"},
	Short:   "Manage alert notification channels",
	Long:    `Manage alert notification channels`,
}

func init() {
	rootCmd.AddCommand(alertnotifications)
	alertnotifications.PersistentFlags().StringP("alertnotification", "a", "", "filter by alert notification slug")

}
