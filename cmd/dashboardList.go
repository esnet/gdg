package cmd

import (
	"fmt"

	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/api"
	"github.com/netsage-project/grafana-dashboard-manager/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listDashboards = &cobra.Command{
	Use:   "list",
	Short: "List all dashboards",
	Long:  `List all dashboards`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "Title", "Slug", "Folder", "UID", "URL"})

		filters := getDashboardGlobalFlags(cmd)
		boards := api.ListDashboards(client, &filters)

		log.Infof("Listing dashboards for context: '%s'", config.GetContext())
		for _, link := range boards {
			url := fmt.Sprintf("%s%s", config.GetDefaultGrafanaConfig().URL, link.URL)
			tableObj.AppendRow(table.Row{link.ID, link.Title, link.Slug, link.FolderTitle,
				link.UID, url})

		}
		if len(boards) > 0 {
			tableObj.Render()
		} else {
			log.Info("No dashboards found")
		}

	},
}

func init() {
	dashboard.AddCommand(listDashboards)
}
