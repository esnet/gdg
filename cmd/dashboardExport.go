package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var exportDashboard = &cobra.Command{
	Use:   "export",
	Short: "export all dashboards",
	Long:  `export all dashboards`,
	Run: func(cmd *cobra.Command, args []string) {

		filter := getDashboardGlobalFlags(cmd)
		api.ExportDashboards(client, filter, configProvider)

		tableObj.AppendHeader(table.Row{"Title", "id", "folder", "UID"})
		boards := api.ListDashboards(client, &filter)

		for _, link := range boards {
			tableObj.AppendRow(table.Row{link.Title, link.ID, link.FolderTitle, link.UID})

		}
		if len(boards) > 0 {
			tableObj.Render()
		} else {
			log.Info("No dashboards found")
		}

	},
}

func init() {
	dashboard.AddCommand(exportDashboard)
}
