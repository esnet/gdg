package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ClearDataSources = &cobra.Command{
	Use:   "clear",
	Short: "clear all datasources",
	Long:  `clear all datasources from grafana`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Delete datasources")
		filters := getDatasourcesGlobalFlags(cmd)
		savedFiles := api.DeleteAllDataSources(client, filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

func init() {
	datasources.AddCommand(ClearDataSources)
}
