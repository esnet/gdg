package cmd

import (
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var connections = &cobra.Command{
	Use:     "connections",
	Aliases: []string{"connection", "ds", "c", "datasource", "datasources"},
	Short:   "Manage connections (formerly Data Sources)",
	Long:    `All software has versions.`,
}

var clearConnections = &cobra.Command{
	Use:     "clear",
	Short:   "clear all connections",
	Long:    `clear all connections from grafana`,
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Delete connections")
		dashboardFilter, _ := cmd.Flags().GetString("datasource")
		filters := service.NewConnectionFilter(dashboardFilter)
		savedFiles := grafanaSvc.DeleteAllConnections(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var uploadConnections = &cobra.Command{
	Use:     "upload ",
	Short:   "upload all connections to grafana",
	Long:    `upload all connections to grafana`,
	Aliases: []string{"u", "export"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Uploading connections")
		dashboardFilter, _ := cmd.Flags().GetString("connection")
		filters := service.NewConnectionFilter(dashboardFilter)
		exportedList := grafanaSvc.UploadConnections(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range exportedList {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var downloadConnections = &cobra.Command{
	Use:     "download",
	Short:   "download all connections from grafana",
	Long:    `download all connections from grafana to local filesystem`,
	Aliases: []string{"d", "import"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Importing connections for context: '%s'", config.Config().GetAppConfig().GetContext())
		dashboardFilter, _ := cmd.Flags().GetString("connection")
		filters := service.NewConnectionFilter(dashboardFilter)
		savedFiles := grafanaSvc.DownloadConnections(filters)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"datasource", file})
		}
		tableObj.Render()

	},
}

var listConnections = &cobra.Command{
	Use:     "list",
	Short:   "List all connections",
	Long:    `List all connections`,
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
		dashboardFilter, _ := cmd.Flags().GetString("connection")
		filters := service.NewConnectionFilter(dashboardFilter)
		dsListing := grafanaSvc.ListConnections(filters)
		log.Infof("Listing connections for context: '%s'", config.Config().GetAppConfig().GetContext())
		if len(dsListing) == 0 {
			log.Info("No connections found")
		} else {
			for _, link := range dsListing {
				url := fmt.Sprintf("%s/datasource/edit/%d", config.Config().GetDefaultGrafanaConfig().URL, link.ID)
				tableObj.AppendRow(table.Row{link.ID, link.UID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, url})
			}
			tableObj.Render()
		}
	},
}

func init() {
	rootCmd.AddCommand(connections)
	connections.PersistentFlags().StringP("connection", "", "", "filter by connection slug")
	connections.AddCommand(clearConnections)
	connections.AddCommand(uploadConnections)
	connections.AddCommand(downloadConnections)
	connections.AddCommand(listConnections)

}
