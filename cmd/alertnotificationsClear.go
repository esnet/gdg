package cmd

import (
	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var clearAlertNotifications = &cobra.Command{
	Use:   "clear",
	Short: "delete all alert notification channels",
	Long:  `clear all alert notification channels from grafana`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"type", "filename"})

		log.Infof("Clearing all alert notification channels for context: '%s'", apphelpers.GetContext())
		deleted := client.DeleteAllAlertNotifications()
		for _, item := range deleted {
			tableObj.AppendRow(table.Row{"alertnotification", item})
		}
		if len(deleted) == 0 {
			log.Info("No alert notification channels were found. 0 removed")
		} else {
			log.Infof("%d alert notification channels were deleted", len(deleted))
			tableObj.Render()
		}
	},
}

func init() {
	alertnotifications.AddCommand(clearAlertNotifications)
}
