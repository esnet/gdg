package cmd

import (
	"fmt"

	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/api"
	"github.com/netsage-project/grafana-dashboard-manager/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listDataSources = &cobra.Command{
	Use:   "list",
	Short: "List all dashboards",
	Long:  `List all dashboards`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "name", "slug", "type", "default", "url"})
		filters := getDatasourcesGlobalFlags(cmd)
		datasources := client.ListDataSources(filters)
		log.Infof("Listing datasources for context: '%s'", config.GetContext())
		if len(datasources) == 0 {
			log.Info("No datasources found")
		} else {
			for _, link := range datasources {
				url := fmt.Sprintf("%s/datasource/edit/%d", config.GetDefaultGrafanaConfig().URL, link.ID)
				tableObj.AppendRow(table.Row{link.ID, link.Name, api.GetSlug(link.Name), link.Type, link.IsDefault, url})
			}
			tableObj.Render()
		}
	},
}

func init() {
	datasources.AddCommand(listDataSources)
}
