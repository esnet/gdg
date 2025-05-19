package backup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/tools"
	"github.com/jedib0t/go-pretty/v6/table"

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

// getConnectionTbWriter returns a table object for use with newConnectionsPermissionListCmd
func getDashboardPermTblWriter() table.Writer {
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)
	writer.SetStyle(table.StyleLight)
	writer.AppendHeader(table.Row{"id", "name", "slug", "folder", "uid", "url"}, table.RowConfig{AutoMerge: true})
	return writer
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
			slog.Info("Listing Dashboard Permissions for context", "context", config.Config().GetGDGConfig().GetContext())
			filters := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			permissions, err := rootCmd.GrafanaSvc().ListDashboardPermissions(filters)
			if err != nil {
				slog.Error("Failed to retrieve Dashboard Permissions", "error", err)
				os.Exit(1)
			}

			if len(permissions) == 0 {
				slog.Info("No Dashboards found")
			} else {
				for _, perms := range permissions {
					writer := getDashboardPermTblWriter()
					urlValue := getDashboardUrl(perms.Dashboard.Hit)
					link := perms.Dashboard
					writer.AppendRow(table.Row{
						link.ID, link.Title, link.Slug, link.NestedPath,
						link.UID, urlValue,
					})
					writer.Render()
					if perms.Permissions != nil {
						twConfigs := table.NewWriter()
						twConfigs.SetOutputMirror(os.Stdout)
						twConfigs.SetStyle(table.StyleDouble)
						twConfigs.AppendHeader(table.Row{"Dashboard UID", "Dashboard Title", "UserLogin", "Team", "RoleName", "Permission"})
						for _, dashPerm := range perms.Permissions {
							var userLogin string
							if len(dashPerm.UserLogin) > 0 {
								if strings.HasPrefix(dashPerm.UserEmail, "sa-") && !strings.Contains(dashPerm.UserEmail, "@") {
									userLogin = fmt.Sprintf("service:%s", dashPerm.UserLogin)
								} else {
									userLogin = fmt.Sprintf("user:%s", dashPerm.UserLogin)
								}
							}
							twConfigs.AppendRow(table.Row{link.UID, link.Title, userLogin, dashPerm.Team, dashPerm.Role, dashPerm.PermissionName})
						}
						if len(perms.Permissions) > 0 {
							twConfigs.Render()
						}
					}
				}
			}
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
			slog.Info("Clear all Dashboard permissions")
			tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will clear all permission from all Dashboards on: '%s' "+
				"(Or all permission matching your filters).  Do you wish to continue (y/n) ", config.Config().GetGDGConfig().ContextName,
			), "", true)
			rootCmd.TableObj.AppendHeader(table.Row{"cleared Dashboard permissions"})
			filters := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			err := rootCmd.GrafanaSvc().ClearDashboardPermissions(filters)
			if err != nil {
				slog.Error("Failed to retrieve Dashboard Permissions", "error", err)
			} else {
				slog.Info("All dashboard permissions have been cleared")
			}
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
			filters := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			permissions, err := rootCmd.GrafanaSvc().DownloadDashboardPermissions(filters)
			if err != nil {
				slog.Error("Failed to retrieve Dashboard Permissions", "error", err)
				os.Exit(1)
			}
			slog.Info("Downloading Dashboard permissions")

			if len(permissions) == 0 {
				slog.Info("No Dashboard permissions")
			} else {
				for _, perm := range permissions {
					rootCmd.TableObj.AppendRow(table.Row{perm})
				}
				rootCmd.Render(cd.CobraCommand, permissions)
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
			slog.Info("Uploading dashboard permissions")
			rootCmd.TableObj.AppendHeader(table.Row{"dashboard permission"})
			filters := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			permissions, err := rootCmd.GrafanaSvc().UploadDashboardPermissions(filters)
			if err != nil {
				slog.Error("Failed to retrieve Dashboard Permissions", "error", err)
				os.Exit(1)
			}

			if len(permissions) == 0 {
				slog.Info("No permissions found")
			} else {
				for _, perm := range permissions {
					rootCmd.TableObj.AppendRow(table.Row{perm})
				}
				rootCmd.Render(cd.CobraCommand, permissions)
			}
			return nil
		},
	}
}
