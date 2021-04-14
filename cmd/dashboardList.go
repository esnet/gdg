package cmd

import (
	"fmt"
	"strings"

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
		tableObj.AppendHeader(table.Row{"id", "Title", "Slug", "folder", "UID", "URL"})

		filter, _ := cmd.Flags().GetString("filter")
		var folders []string

		if filter != "" {
			folders = strings.Split(filter, ",")
		}
		boards := api.ListDashboards(client, folders, "")

		for _, link := range boards {
			url := fmt.Sprintf("%s%s", config.GetGrafanaConfig().URL, link.URL)
			elements := strings.Split(link.URI, "/")
			var slug string = ""
			if len(elements) > 1 {
				slug = elements[len(elements)-1]
			}
			tableObj.AppendRow(table.Row{link.ID, link.Title, slug, link.FolderTitle, link.UID, url})

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
	listDashboards.Flags().StringP("filter", "f", "", "folders UID filter")
}
