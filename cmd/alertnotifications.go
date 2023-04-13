package cmd

import (
	"fmt"
	"github.com/esnet/gdg/internal/apphelpers"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var alertnotifications = &cobra.Command{
	Use:     "alertnotifications",
	Aliases: []string{"an", "alertnotification"},
	Short:   "Manage alert notification channels",
	Long:    `Manage alert notification channels`,
}

// clearAlerts
var clearAlertNotifications = &cobra.Command{
	Use:   "clear",
	Short: "delete all alert notification channels",
	Long:  `clear all alert notification channels from grafana`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Alert Notifications will be deprecated as of Grafana 9.0, this API may no longer work soon")
		tableObj.AppendHeader(table.Row{"type", "filename"})

		log.Infof("Clearing all alert notification channels for context: '%s'", apphelpers.GetContext())
		deleted := grafanaSvc.DeleteAllAlertNotifications()
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

var exportAlertNotifications = &cobra.Command{
	Use:   "export",
	Short: "export all alert notification channels",
	Long:  `export all alert notification channels`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Alert Notifications will be deprecated as of Grafana 9.0, this API may no longer work soon")
		tableObj.AppendHeader(table.Row{"name", "id", "UID"})

		log.Infof("Exporting alert notification channels for context: '%s'", apphelpers.GetContext())
		grafanaSvc.ExportAlertNotifications()
		items := grafanaSvc.ListAlertNotifications()
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

var importAlertNotifications = &cobra.Command{
	Use:   "import",
	Short: "import all alert notification channels",
	Long:  `import all alert notification channels from grafana to local filesystem`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Alert Notifications will be deprecated as of Grafana 9.0, this API may no longer work soon")
		tableObj.AppendHeader(table.Row{"type", "filename"})

		log.Infof("Importing alert notification channels for context: '%s'", apphelpers.GetContext())
		savedFiles := grafanaSvc.ImportAlertNotifications()
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"alertnotification", file})
		}
		tableObj.Render()

	},
}

var listAlertNotifications = &cobra.Command{
	Use:   "list",
	Short: "List all alert notification channels",
	Long:  `List all alert notification channels`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Alert Notifications will be deprecated as of Grafana 9.0, this API may no longer work soon")
		tableObj.AppendHeader(table.Row{"id", "name", "slug", "type", "default", "url"})
		alertnotifications := grafanaSvc.ListAlertNotifications()

		log.Infof("Listing alert notifications channels for context: '%s'", apphelpers.GetContext())
		if len(alertnotifications) == 0 {
			log.Info("No alert notifications found")
		} else {
			for _, link := range alertnotifications {
				url := fmt.Sprintf("%s/alerting/notification/%d/edit", apphelpers.GetCtxDefaultGrafanaConfig().URL, link.ID)
				tableObj.AppendRow(table.Row{link.ID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, url})
			}
			tableObj.Render()
		}
	},
}

func init() {
	rootCmd.AddCommand(alertnotifications)
	alertnotifications.AddCommand(clearAlertNotifications)
	alertnotifications.AddCommand(exportAlertNotifications)
	alertnotifications.AddCommand(importAlertNotifications)
	alertnotifications.AddCommand(listAlertNotifications)
}
