package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/gdg/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var exportAlertNotifications = &cobra.Command{
	Use:   "export",
	Short: "export all alert notification channels",
	Long:  `export all alert notification channels`,
	Run: func(cmd *cobra.Command, args []string) {
		filter := getAlertNotificationsGlobalFlags(cmd)
		tableObj.AppendHeader(table.Row{"name", "id", "UID"})

		log.Infof("Exporting alert notification channels for context: '%s'", apphelpers.GetContext())
		client.ExportAlertNotifications(filter)
		items := client.ListAlertNotifications(filter)
		for _, item := range items {
			tableObj.AppendRow(table.Row{item.Name, item.ID, item.UID})
		}
		if len(items) > 0 {
			tableObj.Render()
		} else {
			log.Info("No alert notification channels found")
		}
	},
}

func init() {
	alertnotifications.AddCommand(exportAlertNotifications)
}
