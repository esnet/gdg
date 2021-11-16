package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/gdg/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var importDashboard = &cobra.Command{
	Use:   "import",
	Short: "Import all dashboards",
	Long:  `Import all dashboards from grafana to local file system`,
	Run: func(cmd *cobra.Command, args []string) {
		filter := getDashboardGlobalFlags(cmd)
		savedFiles := client.ImportDashboards(filter)
		log.Infof("Importing dashboards for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"dashboard", file})
		}
		tableObj.Render()
	},
}

func init() {
	dashboard.AddCommand(importDashboard)

}
