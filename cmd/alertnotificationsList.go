package cmd

import (
	"fmt"

	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/gdg/api"
	"github.com/netsage-project/gdg/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listAlertNotifications = &cobra.Command{
	Use:   "list",
	Short: "List all alert notification channels",
	Long:  `List all alert notification channels`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "name", "slug", "type", "default", "url"})
		alertnotifications := client.ListAlertNotifications()

		log.Infof("Listing alert notifications channels for context: '%s'", apphelpers.GetContext())
		if len(alertnotifications) == 0 {
			log.Info("No alert notifications found")
		} else {
			for _, link := range alertnotifications {
				url := fmt.Sprintf("%s/alerting/notification/%d/edit", apphelpers.GetCtxDefaultGrafanaConfig().URL, link.ID)
				tableObj.AppendRow(table.Row{link.ID, link.Name, api.GetSlug(link.Name), link.Type, link.IsDefault, url})
			}
			tableObj.Render()
		}
	},
}

func init() {
	alertnotifications.AddCommand(listAlertNotifications)
}
