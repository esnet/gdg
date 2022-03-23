package cmd

import (
	"fmt"

	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
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
		boards := client.ListDashboards(filters)

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		for _, link := range boards {
			url := fmt.Sprintf("%s%s", apphelpers.GetCtxDefaultGrafanaConfig().URL, link.URL)
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
