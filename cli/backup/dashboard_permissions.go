package backup

import (
	"context"
	"fmt"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"log/slog"

	"github.com/spf13/cobra"
)

func newDashboardPermissionCmd() simplecobra.Commander {
	description := "Dashboard Permission"
	return &support.SimpleCommand{
		NameP: "permission",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"p", "permissions"}
		},
		CommandsList: []simplecobra.Commander{
			newDashboardPermissionListCmd(),
			newDashboardPermissionDownloadCmd(),
			newDashboardPermissionUploadCmd(),
			newDashboardPermissionClearCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newDashboardPermissionListCmd() simplecobra.Commander {
	description := "List Dashboard Permissions"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			filters := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			slog.Info("Listing Dashboard Permissions for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
			permissions := rootCmd.GrafanaSvc().ListDashboardPermissions(filters)

			if len(permissions) == 0 {
				slog.Info("No Dashboards found")
			} else {
				rootCmd.TableObj.AppendHeader(table.Row{"id", "Title", "Slug", "Folder", "UID", "URL"})
				for _, perms := range permissions {
					urlValue := getDashboardUrl(perms.Dashboard)
					link := perms.Dashboard
					rootCmd.TableObj.AppendRow(table.Row{
						link.ID, link.Title, link.Slug, link.FolderTitle,
						link.UID, urlValue,
					})
					if perms.Permissions != nil {
						for _, dashPerm := range perms.Permissions {

						}
					}
					//		if perms != nil && perms.Enabled {
					//			for _, perm := range perms.Permissions {
					//				rootCmd.TableObj.AppendRow(table.Row{link.ID, link.UID, "    PERMISSION-->", perm.PermissionName, perm.Team, perm.UserEmail})
					//			}
					//		}
				}
				//	rootCmd.Render(cd.CobraCommand, connections)
			}
			//else {

			//}
			return nil
		},
	}
}
func newDashboardPermissionClearCmd() simplecobra.Commander {
	description := "Clear Connection Permissions"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Clear all connections permissions")
			tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will clear all permission from all connections on: '%s' "+
				"(Or all permission matching yoru --connection filter).  Do you wish to continue (y/n) ", config.Config().GetGDGConfig().ContextName,
			), "", true)
			rootCmd.TableObj.AppendHeader(table.Row{"cleared Dashboard permissions"})
			//connectionFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			//filters := service.NewConnectionFilter(connectionFilter)
			//connections := rootCmd.GrafanaSvc().DeleteAllConnectionPermissions(filters)

			//if len(connections) == 0 {
			//	slog.Info("No connections found")
			//} else {
			//	for _, connections := range connections {
			//		rootCmd.TableObj.AppendRow(table.Row{connections})
			//	}
			//	rootCmd.Render(cd.CobraCommand, connections)
			//}

			return nil
		},
	}
}

func newDashboardPermissionDownloadCmd() simplecobra.Commander {
	description := "Download Connection Permissions"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Download Connections for context",
				"context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"filename"})
			connectionFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(connectionFilter)
			connections := rootCmd.GrafanaSvc().DownloadConnectionPermissions(filters)
			slog.Info("Downloading connections permissions")

			if len(connections) == 0 {
				slog.Info("No connections found")
			} else {
				for _, connections := range connections {
					rootCmd.TableObj.AppendRow(table.Row{connections})
				}
				rootCmd.Render(cd.CobraCommand, connections)
			}
			return nil
		},
	}
}
func newDashboardPermissionUploadCmd() simplecobra.Commander {
	description := "Upload Connection Permissions"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Uploading connections permissions")
			rootCmd.TableObj.AppendHeader(table.Row{"connection permission"})
			connectionFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(connectionFilter)
			connections := rootCmd.GrafanaSvc().UploadConnectionPermissions(filters)

			if len(connections) == 0 {
				slog.Info("No connections found")
			} else {
				for _, connections := range connections {
					rootCmd.TableObj.AppendRow(table.Row{connections})
				}
				rootCmd.Render(cd.CobraCommand, connections)
			}
			return nil
		},
	}
}
