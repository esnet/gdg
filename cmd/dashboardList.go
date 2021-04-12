package cmd

import (
	"strings"

	"github.com/esnet/grafana-dashboard-manager/api"
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listDashboards = &cobra.Command{
	Use:   "list",
	Short: "List all dashboards",
	Long:  `List all dashboards`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"Title", "id", "folder", "UID"})

		filter, _ := cmd.Flags().GetString("filter")
		var folders []string

		if filter != "" {
			folders = strings.Split(filter, ",")
		}
		boards := api.ListDashboards(client, folders, "")

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
	dashboard.AddCommand(listDashboards)
	listDashboards.Flags().StringP("filter", "f", "", "folders UID filter")
}
