package backup

import (
	"context"
	"errors"
	"fmt"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"log"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func newConnectionsPermissionCmd() simplecobra.Commander {
	description := "Connections Permission"
	return &support.SimpleCommand{
		NameP: "permission",
		Short: description,
		Long:  description,
		InitCFunc: func(cd *simplecobra.Commandeer, r *support.RootCommand) error {
			valid := tools.ValidateMinimumVersion("v10.4.0", r.GrafanaSvc()) && config.Config().GetDefaultGrafanaConfig().EnterpriseSupport
			if !valid {
				return errors.New("connection Permissions requires grafana version v10.4.0 and enterprise_support to be configured for the gdg context")
			}
			return nil

		},
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
func getConnectionTblWriter() table.Writer {
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
			connections := rootCmd.GrafanaSvc().ListConnectionPermissions(filters)
			if len(connections) == 0 {
				slog.Info("No connections found")
			} else {
				output, _ := cd.CobraCommand.Flags().GetString("output")
				if output == "json" {
					log.Fatal("json output is not supported for connection permission listing")
				}
				for link, perms := range connections {
					wr := getConnectionTblWriter()
					url := fmt.Sprintf("%s/datasource/edit/%d", config.Config().GetDefaultGrafanaConfig().URL, link.ID)
					wr.AppendRow(table.Row{link.ID, link.UID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, url})
					wr.Render()
					if perms != nil {
						twConfigs := table.NewWriter()
						twConfigs.SetOutputMirror(os.Stdout)
						twConfigs.SetStyle(table.StyleColoredCyanWhiteOnBlack)
						twConfigs.AppendHeader(table.Row{"Connection UID", "Permission", "RoleName", "Team", "UserLogin"})
						for _, perm := range perms {
							permLabel := ""
							_ = permLabel
							if perm.BuiltInRole == "" {
								permLabel = perm.Permission
							} else {
								permLabel = perm.BuiltInRole
							}
							twConfigs.AppendRow(table.Row{link.UID, permLabel, perm.RoleName, perm.Team, perm.UserLogin})
						}
						twConfigs.Render()
					}
				}
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
				"(Or all permission matching your --connection filter).  Do you wish to continue (y/n) ", config.Config().GetGDGConfig().ContextName,
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
