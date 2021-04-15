package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ImportDataSources = &cobra.Command{
	Use:   "import",
	Short: "import all datasources",
	Long:  `import all datasources from grafana to local filesystem`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Retrieving datasources")
		savedFiles := api.ImportDataSources(client, configProvider)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

func init() {
	datasources.AddCommand(ImportDataSources)
}
