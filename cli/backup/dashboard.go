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
	"github.com/spf13/cobra"
	"log/slog"
	"net/url"
	"strings"
)

var skipConfirmAction bool

func parseDashboardGlobalFlags(command *cobra.Command) []string {
	folderFilter, _ := command.Flags().GetString("folder")
	dashboardFilter, _ := command.Flags().GetString("dashboard")
	tagsFilter, _ := command.Flags().GetStringSlice("tags")

	return []string{folderFilter, dashboardFilter, strings.Join(tagsFilter, ",")}
}

func newDashboardCommand() simplecobra.Commander {
	description := "Manage Grafana Dashboards"
	return &support.SimpleCommand{
		NameP: "dashboards",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"dash", "dashboard"}
			dashboard := cmd
			dashboard.PersistentFlags().BoolVarP(&skipConfirmAction, "skip-confirmation", "", false, "when set to true, bypass confirmation prompts")
			dashboard.PersistentFlags().StringP("dashboard", "d", "", "filter by dashboard slug")
			dashboard.PersistentFlags().StringP("folder", "f", "", "Filter by Folder Name (Quotes in names not supported)")
			dashboard.PersistentFlags().StringSliceP("tags", "t", []string{}, "Filter by Tags (does not apply on upload)")
		},
		CommandsList: []simplecobra.Commander{
			newListDashboardsCmd(),
			newDownloadDashboardsCmd(),
			newUploadDashboardsCmd(),
			newClearDashboardsCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}

}

func newClearDashboardsCmd() simplecobra.Commander {
	description := "delete all monitored dashboards from grafana"
	return &support.SimpleCommand{
		NameP:        "clear",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			filter := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			deletedDashboards := rootCmd.GrafanaSvc().DeleteAllDashboards(filter)
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			for _, file := range deletedDashboards {
				rootCmd.TableObj.AppendRow(table.Row{"dashboard", file})
			}
			if len(deletedDashboards) == 0 {
				slog.Info("No dashboards were found. 0 dashboards were removed")

			} else {
				slog.Info("dashboards were deleted", "count", len(deletedDashboards))
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}

}

func newUploadDashboardsCmd() simplecobra.Commander {
	description := "upload all dashboards to grafana"
	return &support.SimpleCommand{
		NameP:        "upload",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			filter := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)

			if !skipConfirmAction {
				tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will delete all dashboards from the monitored folders: '%s' "+
					"(or all folders if ignore_dashboard_filters is set to true) and upload your local copy.  Do you wish to "+
					"continue (y/n) ", strings.Join(config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders(), ", "),
				), "", true)
			}
			rootCmd.GrafanaSvc().UploadDashboards(filter)

			rootCmd.TableObj.AppendHeader(table.Row{"Title", "id", "folder", "UID"})
			boards := rootCmd.GrafanaSvc().ListDashboards(filter)

			for _, link := range boards {
				rootCmd.TableObj.AppendRow(table.Row{link.Title, link.ID, link.FolderTitle, link.UID})

			}
			if len(boards) > 0 {
				rootCmd.TableObj.Render()
			} else {
				slog.Info("No dashboards found")
			}
			return nil
		},
	}

}

func newDownloadDashboardsCmd() simplecobra.Commander {
	description := "download all dashboards from grafana"
	return &support.SimpleCommand{
		NameP:        "download",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			filter := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			savedFiles := rootCmd.GrafanaSvc().DownloadDashboards(filter)
			slog.Info("Downloading dashboards for context",
				"context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			for _, file := range savedFiles {
				rootCmd.TableObj.AppendRow(table.Row{"dashboard", file})
			}
			rootCmd.TableObj.Render()
			return nil
		},
	}
}
func newListDashboardsCmd() simplecobra.Commander {
	description := "List all dashboards from grafana"
	return &support.SimpleCommand{
		NameP:        "list",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"id", "Title", "Slug", "Folder", "UID", "Tags", "URL"})

			filters := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			boards := rootCmd.GrafanaSvc().ListDashboards(filters)

			slog.Info("Listing dashboards for context", "context", config.Config().GetGDGConfig().GetContext())
			for _, link := range boards {
				base, err := url.Parse(config.Config().GetDefaultGrafanaConfig().URL)
				var baseHost string
				if err != nil {
					baseHost = "http://unknown/"
				} else {
					base.Path = ""
					baseHost = base.String()
				}
				urlValue := fmt.Sprintf("%s%s", baseHost, link.URL)
				rootCmd.TableObj.AppendRow(table.Row{link.ID, link.Title, link.Slug, link.FolderTitle,
					link.UID, strings.Join(link.Tags, ","), urlValue})

			}
			if len(boards) > 0 {
				rootCmd.TableObj.Render()
			} else {
				slog.Info("No dashboards found")
			}
			return nil
		},
	}

}
