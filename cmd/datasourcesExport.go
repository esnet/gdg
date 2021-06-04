package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var exportDataSources = &cobra.Command{
	Use:   "export ",
	Short: "export all datasources",
	Long:  `export all datasources`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Exporting datasources")
		filters := getDatasourcesGlobalFlags(cmd)
		exportedList := client.ExportDataSources(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range exportedList {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

func init() {
	datasources.AddCommand(exportDataSources)
}
