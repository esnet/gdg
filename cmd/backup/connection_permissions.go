package backup

import (
	"context"
	"fmt"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newConnectionsPermissionCmd() simplecobra.Commander {
	description := "Connections Permission"
	return &support.SimpleCommand{
		NameP: "permission",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l", "permissions"}
		},
		CommandsList: []simplecobra.Commander{
			newConnectionsPermissionListCmd(),
			newConnectionsPermissionDownloadCmd(),
			newConnectionsPermissionUploadCmd(),
			newConnectionsPermissionClearCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newConnectionsPermissionListCmd() simplecobra.Commander {
	description := "List Connection Permissions"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			connectionFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(connectionFilter)
			log.Infof("Listing Connection Permissions for context: '%s'", config.Config().GetAppConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
			connections := rootCmd.GrafanaSvc().ListConnectionPermissions(filters)
			_ = connections

			if len(connections) == 0 {
				log.Info("No connections found")
			} else {
				for link, perms := range connections {
					url := fmt.Sprintf("%s/datasource/edit/%d", config.Config().GetDefaultGrafanaConfig().URL, link.ID)
					rootCmd.TableObj.AppendRow(table.Row{link.ID, link.UID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, url})
					if perms != nil && perms.Enabled {
						for _, perm := range perms.Permissions {
							rootCmd.TableObj.AppendRow(table.Row{link.ID, link.UID, "    PERMISSION-->", perm.PermissionName, perm.Team, perm.UserEmail})
						}
					}
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}
func newConnectionsPermissionClearCmd() simplecobra.Commander {
	description := "Clear Connection Permissions"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Clear all connections permissions")
			tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will clear all permission from all connections on: '%s' "+
				"(Or all permission matching yoru --connection filter).  Do you wish to continue (y/n) ", config.Config().GetAppConfig().ContextName,
			), "", true)
			rootCmd.TableObj.AppendHeader(table.Row{"cleared connection permissions"})
			connectionFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(connectionFilter)
			connections := rootCmd.GrafanaSvc().DeleteAllConnectionPermissions(filters)

			if len(connections) == 0 {
				log.Info("No connections found")
			} else {
				for _, connections := range connections {
					rootCmd.TableObj.AppendRow(table.Row{connections})
				}
				rootCmd.TableObj.Render()
			}

			return nil
		},
	}
}

func newConnectionsPermissionDownloadCmd() simplecobra.Commander {
	description := "Download Connection Permissions"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("import Connections for context: '%s'", config.Config().GetAppConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"filename"})
			connectionFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(connectionFilter)
			connections := rootCmd.GrafanaSvc().DownloadConnectionPermissions(filters)
			log.Infof("Downloading connections permissions")

			if len(connections) == 0 {
				log.Info("No connections found")
			} else {
				for _, connections := range connections {
					rootCmd.TableObj.AppendRow(table.Row{connections})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}
func newConnectionsPermissionUploadCmd() simplecobra.Commander {
	description := "Upload Connection Permissions"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Uploading connections permissions")
			rootCmd.TableObj.AppendHeader(table.Row{"connection permission"})
			connectionFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(connectionFilter)
			connections := rootCmd.GrafanaSvc().UploadConnectionPermissions(filters)

			if len(connections) == 0 {
				log.Info("No connections found")
			} else {
				for _, connections := range connections {
					rootCmd.TableObj.AppendRow(table.Row{connections})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}
