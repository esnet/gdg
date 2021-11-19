package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/gdg/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var importAlertNotifications = &cobra.Command{
	Use:   "import",
	Short: "import all alert notification channels",
	Long:  `import all alert notification channels from grafana to local filesystem`,
	Run: func(cmd *cobra.Command, args []string) {
		filters := getAlertNotificationsGlobalFlags(cmd)
		tableObj.AppendHeader(table.Row{"type", "filename"})

		log.Infof("Importing alert notification channels for context: '%s'", apphelpers.GetContext())
		savedFiles := client.ImportAlertNotifications(filters)
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"alertnotification", file})
		}
		tableObj.Render()

	},
}

func init() {
	alertnotifications.AddCommand(importAlertNotifications)
}
