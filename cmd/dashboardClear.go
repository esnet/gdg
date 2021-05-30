package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ClearDashboards = &cobra.Command{
	Use:   "clear",
	Short: "delete all monitored dashboards",
	Long:  `clear all monitored dashboards from grafana`,
	Run: func(cmd *cobra.Command, args []string) {
		filter := getDashboardGlobalFlags(cmd)
		deletedDashboards := client.DeleteAllDashboards(filter)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range deletedDashboards {
			tableObj.AppendRow(table.Row{"dashboard", file})
		}
		if len(deletedDashboards) == 0 {
			log.Info("No dashboards were found.  0 dashboards removed")

		} else {
			log.Infof("%d dashboards were deleted", len(deletedDashboards))
			tableObj.Render()
		}

	},
}

func init() {
	dashboard.AddCommand(ClearDashboards)
}
