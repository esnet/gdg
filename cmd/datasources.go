package cmd

import (
	"fmt"
	"github.com/esnet/gdg/internal/apphelpers"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
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
	Use:     "clear",
	Short:   "clear all datasources",
	Long:    `clear all datasources from grafana`,
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Delete datasources")
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := service.NewDataSourceFilter(dashboardFilter)
		savedFiles := grafanaSvc.DeleteAllDataSources(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var uploadDataSources = &cobra.Command{
	Use:     "upload ",
	Short:   "upload all datasources to grafana",
	Long:    `upload all datasources to grafana`,
	Aliases: []string{"u", "export"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Exporting datasources")
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := service.NewDataSourceFilter(dashboardFilter)
		exportedList := grafanaSvc.ExportDataSources(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range exportedList {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var downloadDataSources = &cobra.Command{
	Use:     "download",
	Short:   "download all datasources from grafana",
	Long:    `download all datasources from grafana to local filesystem`,
	Aliases: []string{"d", "import"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Importing datasources for context: '%s'", apphelpers.GetContext())
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := service.NewDataSourceFilter(dashboardFilter)
		savedFiles := grafanaSvc.ImportDataSources(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var listDataSources = &cobra.Command{
	Use:     "list",
	Short:   "List all dashboards",
	Long:    `List all dashboards`,
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := service.NewDataSourceFilter(dashboardFilter)
		dsListing := grafanaSvc.ListDataSources(filters)
		log.Infof("Listing datasources for context: '%s'", apphelpers.GetContext())
		if len(dsListing) == 0 {
			log.Info("No datasources found")
		} else {
			for _, link := range dsListing {
				url := fmt.Sprintf("%s/datasource/edit/%d", apphelpers.GetCtxDefaultGrafanaConfig().URL, link.ID)
				tableObj.AppendRow(table.Row{link.ID, link.UID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, url})
			}
			tableObj.Render()
		}
	},
}

func init() {
	rootCmd.AddCommand(datasources)
	datasources.PersistentFlags().StringP("datasource", "d", "", "filter by datasource slug")
	datasources.AddCommand(clearDataSources)
	datasources.AddCommand(uploadDataSources)
	datasources.AddCommand(downloadDataSources)
	datasources.AddCommand(listDataSources)

}
