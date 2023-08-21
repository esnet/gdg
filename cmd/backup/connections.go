package backup

import (
	"fmt"
	"github.com/esnet/gdg/cmd"
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
	Run: func(command *cobra.Command, args []string) {

		log.Info("Delete connections")
		dashboardFilter, _ := command.Flags().GetString("datasource")
		filters := service.NewConnectionFilter(dashboardFilter)
		savedFiles := cmd.GetGrafanaSvc().DeleteAllConnections(filters)
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			cmd.TableObj.AppendRow(table.Row{"datasource", file})
		}
		cmd.TableObj.Render()

	},
}

var uploadConnections = &cobra.Command{
	Use:     "upload ",
	Short:   "upload all connections to grafana",
	Long:    `upload all connections to grafana`,
	Aliases: []string{"u"},
	Run: func(command *cobra.Command, args []string) {
		log.Info("Uploading connections")
		dashboardFilter, _ := command.Flags().GetString("connection")
		filters := service.NewConnectionFilter(dashboardFilter)
		exportedList := cmd.GetGrafanaSvc().UploadConnections(filters)
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range exportedList {
			cmd.TableObj.AppendRow(table.Row{"datasource", file})
		}
		cmd.TableObj.Render()

	},
}

var downloadConnections = &cobra.Command{
	Use:     "download",
	Short:   "download all connections from grafana",
	Long:    `download all connections from grafana to local filesystem`,
	Aliases: []string{"d"},
	Run: func(command *cobra.Command, args []string) {
		log.Infof("Importing connections for context: '%s'", config.Config().GetAppConfig().GetContext())
		dashboardFilter, _ := command.Flags().GetString("connection")
		filters := service.NewConnectionFilter(dashboardFilter)
		savedFiles := cmd.GetGrafanaSvc().DownloadConnections(filters)
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			cmd.TableObj.AppendRow(table.Row{"datasource", file})
		}
		cmd.TableObj.Render()

	},
}

var listConnections = &cobra.Command{
	Use:     "list",
	Short:   "List all connections",
	Long:    `List all connections`,
	Aliases: []string{"l"},
	Run: func(command *cobra.Command, args []string) {
		cmd.TableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
		dashboardFilter, _ := command.Flags().GetString("connection")
		filters := service.NewConnectionFilter(dashboardFilter)
		dsListing := cmd.GetGrafanaSvc().ListConnections(filters)
		log.Infof("Listing connections for context: '%s'", config.Config().GetAppConfig().GetContext())
		if len(dsListing) == 0 {
			log.Info("No connections found")
		} else {
			for _, link := range dsListing {
				url := fmt.Sprintf("%s/datasource/edit/%d", config.Config().GetDefaultGrafanaConfig().URL, link.ID)
				cmd.TableObj.AppendRow(table.Row{link.ID, link.UID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, url})
			}
			cmd.TableObj.Render()
		}
	},
}

func init() {
	backupCmd.AddCommand(connections)
	connections.PersistentFlags().StringP("connection", "", "", "filter by connection slug")
	connections.AddCommand(clearConnections)
	connections.AddCommand(uploadConnections)
	connections.AddCommand(downloadConnections)
	connections.AddCommand(listConnections)

}
