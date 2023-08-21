package backup

import (
	"fmt"
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var connectionsPermissionCmd = &cobra.Command{
	Use:     "permission",
	Aliases: []string{"p", "permissions"},
	Short:   "Connections Permission",
	Long:    `Connections Permission`,
}

var listConnectionsPermissionsCmd = &cobra.Command{
	Use:   "list",
	Short: "list Connections Permissions",
	Long:  `list Connections Permissions`,
	Run: func(command *cobra.Command, args []string) {
		connectionFilter, _ := command.Flags().GetString("connection")
		filters := service.NewConnectionFilter(connectionFilter)
		log.Infof("Listing Connection Permissions for context: '%s'", config.Config().GetAppConfig().GetContext())
		cmd.TableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
		connections := cmd.GetGrafanaSvc().ListConnectionPermissions(filters)
		_ = connections

		if len(connections) == 0 {
			log.Info("No connections found")
		} else {
			for link, perms := range connections {
				url := fmt.Sprintf("%s/datasource/edit/%d", config.Config().GetDefaultGrafanaConfig().URL, link.ID)
				cmd.TableObj.AppendRow(table.Row{link.ID, link.UID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, url})
				if perms != nil && perms.Enabled {
					for _, perm := range perms.Permissions {
						cmd.TableObj.AppendRow(table.Row{link.ID, link.UID, "    PERMISSION-->", perm.PermissionName, perm.Team, perm.UserEmail})
					}
				}
			}
			cmd.TableObj.Render()
		}

	},
}

var downloadConnectionsPermissionsCmd = &cobra.Command{
	Use:     "download",
	Short:   "download Connections Permissions",
	Long:    `downloadConnections Permissions`,
	Aliases: []string{"d"},
	Run: func(command *cobra.Command, args []string) {
		log.Infof("import Connections for context: '%s'", config.Config().GetAppConfig().GetContext())
		cmd.TableObj.AppendHeader(table.Row{"filename"})
		connectionFilter, _ := command.Flags().GetString("connection")
		filters := service.NewConnectionFilter(connectionFilter)
		connections := cmd.GetGrafanaSvc().DownloadConnectionPermissions(filters)
		log.Infof("Downloading connections permissions")

		if len(connections) == 0 {
			log.Info("No connections found")
		} else {
			for _, connections := range connections {
				cmd.TableObj.AppendRow(table.Row{connections})
			}
			cmd.TableObj.Render()
		}

	},
}

var uploadConnectionsPermissionsCmd = &cobra.Command{
	Use:   "upload",
	Short: "upload Connections Permissions",
	Long:  `uploadConnections Permissions`,
	Run: func(command *cobra.Command, args []string) {
		log.Infof("Uploading connections permissions")
		cmd.TableObj.AppendHeader(table.Row{"connection permission"})
		connectionFilter, _ := command.Flags().GetString("connection")
		filters := service.NewConnectionFilter(connectionFilter)
		connections := cmd.GetGrafanaSvc().UploadConnectionPermissions(filters)

		if len(connections) == 0 {
			log.Info("No connections found")
		} else {
			for _, connections := range connections {
				cmd.TableObj.AppendRow(table.Row{connections})
			}
			cmd.TableObj.Render()
		}

	},
}

var clearConnectionsPermissionsCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear Connections Permissions",
	Long:  `clear Connections Permissions`,
	Run: func(command *cobra.Command, args []string) {
		log.Infof("Clear all connections permissions")
		tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will clear all permission from all connections on: '%s' "+
			"(Or all permission matching yoru --connection filter).  Do you wish to continue (y/n) ", config.Config().GetAppConfig().ContextName,
		), "", true)
		cmd.TableObj.AppendHeader(table.Row{"cleared connection permissions"})
		connectionFilter, _ := command.Flags().GetString("connection")
		filters := service.NewConnectionFilter(connectionFilter)
		connections := cmd.GetGrafanaSvc().DeleteAllConnectionPermissions(filters)

		if len(connections) == 0 {
			log.Info("No connections found")
		} else {
			for _, connections := range connections {
				cmd.TableObj.AppendRow(table.Row{connections})
			}
			cmd.TableObj.Render()
		}

	},
}

func init() {
	connections.AddCommand(connectionsPermissionCmd)
	connectionsPermissionCmd.AddCommand(listConnectionsPermissionsCmd)
	connectionsPermissionCmd.AddCommand(downloadConnectionsPermissionsCmd)
	connectionsPermissionCmd.AddCommand(uploadConnectionsPermissionsCmd)
	connectionsPermissionCmd.AddCommand(clearConnectionsPermissionsCmd)

}
