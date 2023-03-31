package cmd

import (
	"fmt"
	"github.com/esnet/gdg/api"
	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var datasources = &cobra.Command{
	Use:     "datasources",
	Aliases: []string{"ds", "datasource"},
	Short:   "Manage datasources",
	Long:    `All software has versions.`,
}

var clearDataSources = &cobra.Command{
	Use:   "clear",
	Short: "clear all datasources",
	Long:  `clear all datasources from grafana`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Delete datasources")
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := api.NewDataSourceFilter(dashboardFilter)
		savedFiles := client.DeleteAllDataSources(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var exportDataSources = &cobra.Command{
	Use:   "export ",
	Short: "export all datasources",
	Long:  `export all datasources`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Exporting datasources")
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := api.NewDataSourceFilter(dashboardFilter)
		exportedList := client.ExportDataSources(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range exportedList {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var ImportDataSources = &cobra.Command{
	Use:   "import",
	Short: "import all datasources",
	Long:  `import all datasources from grafana to local filesystem`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Importing datasources for context: '%s'", apphelpers.GetContext())
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := api.NewDataSourceFilter(dashboardFilter)
		savedFiles := client.ImportDataSources(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var listDataSources = &cobra.Command{
	Use:   "list",
	Short: "List all dashboards",
	Long:  `List all dashboards`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := api.NewDataSourceFilter(dashboardFilter)
		dsListing := client.ListDataSources(filters)
		log.Infof("Listing datasources for context: '%s'", apphelpers.GetContext())
		if len(dsListing) == 0 {
			log.Info("No datasources found")
		} else {
			for _, link := range dsListing {
				url := fmt.Sprintf("%s/datasource/edit/%d", apphelpers.GetCtxDefaultGrafanaConfig().URL, link.ID)
				tableObj.AppendRow(table.Row{link.ID, link.UID, link.Name, api.GetSlug(link.Name), link.Type, link.IsDefault, url})
			}
			tableObj.Render()
		}
	},
}

func init() {
	rootCmd.AddCommand(datasources)
	datasources.PersistentFlags().StringP("datasource", "d", "", "filter by datasource slug")
	datasources.AddCommand(clearDataSources)
	datasources.AddCommand(exportDataSources)
	datasources.AddCommand(ImportDataSources)
	datasources.AddCommand(listDataSources)

}
