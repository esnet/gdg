package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ClearDashboards = &cobra.Command{
	Use:   "clear",
	Short: "delete all monitored dashboards",
	Long:  `clear all monitored dashboards from grafana`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Delete all dashboards")
		savedFiles := api.DeleteAllDashboards(client, nil)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

func init() {
	dashboard.AddCommand(ClearDashboards)
}
