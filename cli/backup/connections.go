package backup

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

func newConnectionsCommand() simplecobra.Commander {
	description := "Manage connections (formerly Data Sources)"
	return &support.SimpleCommand{
		NameP: "connections",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"connection", "ds", "c", "datasource", "datasources"}
			connections := cmd
			connections.PersistentFlags().StringP("connection", "", "", "filter by connection slug")
		},
		CommandsList: []simplecobra.Commander{
			newClearConnectionsCmd(),
			newUploadConnectionsCmd(),
			newDownloadConnectionsCmd(),
			newListConnectionsCmd(),
			newConnectionsPermissionCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newClearConnectionsCmd() simplecobra.Commander {
	description := "clear all connections for the given Organization"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Delete connections", slog.String("Organization", GetOrganizationName()))
			dashboardFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(dashboardFilter)
			savedFiles := rootCmd.GrafanaSvc().DeleteAllConnections(filters)
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			for _, file := range savedFiles {
				rootCmd.TableObj.AppendRow(table.Row{"datasource", file})
			}
			rootCmd.Render(cd.CobraCommand, savedFiles)
			return nil
		},
	}
}

func newUploadConnectionsCmd() simplecobra.Commander {
	description := "upload all connections to grafana for the given Organization"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Uploading connections", slog.String("Organization", GetOrganizationName()))
			dashboardFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(dashboardFilter)
			exportedList := rootCmd.GrafanaSvc().UploadConnections(filters)
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			for _, file := range exportedList {
				rootCmd.TableObj.AppendRow(table.Row{"datasource", file})
			}
			rootCmd.Render(cd.CobraCommand, exportedList)
			return nil
		},
	}
}

func newDownloadConnectionsCmd() simplecobra.Commander {
	description := "download all connections from grafana for the given Organization"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Importing connections for context",
				slog.String("Organization", GetOrganizationName()),
				"context", config.Config().GetGDGConfig().GetContext())
			dashboardFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(dashboardFilter)
			savedFiles := rootCmd.GrafanaSvc().DownloadConnections(filters)
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			for _, file := range savedFiles {
				rootCmd.TableObj.AppendRow(table.Row{"connection", file})
			}
			rootCmd.Render(cd.CobraCommand, savedFiles)
			return nil
		},
	}
}

func newListConnectionsCmd() simplecobra.Commander {
	description := "List all connections for the given Organization"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"id", "uid", "name", "slug", "type", "default", "url"})
			dashboardFilter, _ := cd.CobraCommand.Flags().GetString("connection")
			filters := service.NewConnectionFilter(dashboardFilter)
			dsListing := rootCmd.GrafanaSvc().ListConnections(filters)
			slog.Info("Listing connections for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))
			if len(dsListing) == 0 {
				slog.Info("No connections found")
			} else {
				for _, link := range dsListing {
					rootCmd.TableObj.AppendRow(table.Row{link.ID, link.UID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, getConnectionURL(link.UID)})
				}
				rootCmd.Render(cd.CobraCommand, dsListing)
			}
			return nil
		},
	}
}

func getConnectionURL(uid string) string {
	url := config.Config().GetDefaultGrafanaConfig().GetURL()
	return fmt.Sprintf("%s/connections/datasources/edit/%s", url, uid)
}
