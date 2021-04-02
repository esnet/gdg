package cmd

import (
	"github.com/esnet/grafana-dashboard-manager/api"
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var exportDashboard = &cobra.Command{
	Use:   "export",
	Short: "export all dashboards",
	Long:  `export all dashboards`,
	Run: func(cmd *cobra.Command, args []string) {
		api.ExportDashboards(client, nil, "", configProvider)

		tableObj.AppendHeader(table.Row{"Title", "id", "folder", "UID"})
		boards := api.ListDashboards(client, nil, "")

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
