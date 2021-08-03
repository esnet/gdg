package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ImportDataSources = &cobra.Command{
	Use:   "import",
	Short: "import all datasources",
	Long:  `import all datasources from grafana to local filesystem`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Importing datasources for context: '%s'", apphelpers.GetContext())
		filters := getDatasourcesGlobalFlags(cmd)
		savedFiles := client.ImportDataSources(filters)
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
