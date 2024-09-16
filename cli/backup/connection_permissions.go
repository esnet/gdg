package backup

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/tools"
	"github.com/jedib0t/go-pretty/v6/table"

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

// getConnectionTbWriter returns a table object for use with newConnectionsPermissionListCmd
func getConnPermTblWriter() table.Writer {
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)
	writer.SetStyle(table.StyleLight)
	writer.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"}, table.RowConfig{AutoMerge: true})
	return writer
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
			slog.Info("Listing Connection Permissions for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
			connections := rootCmd.GrafanaSvc().ListConnectionPermissions(filters)

			if len(connections) == 0 {
				slog.Info("No connections found")
			} else {
				for _, item := range connections {
					writer := getConnPermTblWriter()
					writer.AppendRow(table.Row{item.Connection.ID, item.Connection.UID, item.Connection.Name, service.GetSlug(item.Connection.Name), item.Connection.Type, item.Connection.IsDefault, getConnectionURL(item.Connection.UID)})
					writer.Render()
					if item.Permissions != nil {
						twConfigs := table.NewWriter()
						twConfigs.SetOutputMirror(os.Stdout)
						twConfigs.SetStyle(table.StyleDouble)
						twConfigs.AppendHeader(table.Row{"Connection UID", "Permission Granted", "Permission Type", "Permission Grantee"})
						for _, perm := range item.Permissions {
							permissionType := "BuiltinRole"
							value := ""
							if perm.BuiltInRole != "" {
								value = "builtInRole:" + perm.BuiltInRole
							} else if perm.UserLogin != "" {
								permissionType = "User"
								value = "user:" + perm.UserLogin
							} else if perm.Team != "" {
								permissionType = "Team"
								value = "team:" + perm.Team
							} else {
								permissionType = "unsupported"
							}
							twConfigs.AppendRow(table.Row{item.Connection.UID, perm.Permission, permissionType, value})
						}
						twConfigs.Render()
					}
				}
				// rootCmd.Render(cd.CobraCommand, connections)
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
			slog.Info("Clear all connections permissions")
			tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will clear all permission from all connections on: '%s' "+
				"(Or all permission matching yoru --connection filter).  Do you wish to continue (y/n) ", config.Config().GetGDGConfig().ContextName,
			), "", true)
			rootCmd.TableObj.AppendHeader(table.Row{"cleared connection permissions"})
			connectionFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(connectionFilter)
			connections := rootCmd.GrafanaSvc().DeleteAllConnectionPermissions(filters)

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
			slog.Info("Uploading connections permissions")
			rootCmd.TableObj.AppendHeader(table.Row{"connection permission applied"})
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
