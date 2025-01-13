package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var skipConfirmAction bool

func parseDashboardGlobalFlags(command *cobra.Command) []string {
	folderFilter, _ := command.Flags().GetString("folder")
	dashboardFilter, _ := command.Flags().GetString("dashboard")
	tagsFilter, _ := command.Flags().GetStringArray("tags")
	rawTags, err := json.Marshal(tagsFilter)
	jsonTags := ""
	if err == nil {
		jsonTags = string(rawTags)
	}

	return []string{folderFilter, dashboardFilter, jsonTags}
}

func newDashboardCommand() simplecobra.Commander {
	description := "Manage Grafana Dashboards"
	return &support.SimpleCommand{
		NameP: "dashboards",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"dash", "dashboard"}
			cmd.PersistentFlags().BoolVarP(&skipConfirmAction, "skip-confirmation", "", false, "when set to true, bypass confirmation prompts")
			cmd.PersistentFlags().StringP("dashboard", "d", "", "filter by dashboard slug")
			cmd.PersistentFlags().StringP("folder", "f", "", "Filter by Folder Name (Quotes in names not supported)")
			cmd.PersistentFlags().StringArrayP("tags", "t", []string{}, "Filter by list of comma delimited tags")
		},
		CommandsList: []simplecobra.Commander{
			newListDashboardsCmd(),
			newDownloadDashboardsCmd(),
			newUploadDashboardsCmd(),
			newClearDashboardsCmd(),
			// Permissions
			newDashboardPermissionCmd(),
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
				rootCmd.Render(cd.CobraCommand, deletedDashboards)
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
			cmd.Aliases = []string{"u", "up"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			filter := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)

			if !skipConfirmAction {
				tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will delete all dashboards from the monitored folders: '%s' "+
					"(or all folders if ignore_dashboard_filters is set to true) and upload your local copy.  Do you wish to "+
					"continue (y/n) ", strings.Join(config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders(), ", "),
				), "", true)
			}
			err := rootCmd.GrafanaSvc().UploadDashboards(filter)
			if err != nil {
				return err
			}

			rootCmd.TableObj.AppendHeader(table.Row{"Title", "id", "folder", "UID"})
			boards := rootCmd.GrafanaSvc().ListDashboards(filter)

			slog.Info("dashboards have been uploaded", slog.Any("count", len(boards)),
				slog.String("context", GetContext()),
				slog.String("Organization", GetOrganizationName()))
			for _, link := range boards {
				rootCmd.TableObj.AppendRow(table.Row{link.Title, link.ID, link.FolderTitle, link.UID})
			}
			if len(boards) > 0 {
				rootCmd.Render(cd.CobraCommand, boards)
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
				slog.String("Organization", GetOrganizationName()),
				"context", GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			for _, file := range savedFiles {
				rootCmd.TableObj.AppendRow(table.Row{"dashboard", file})
			}
			rootCmd.Render(cd.CobraCommand, savedFiles)
			return nil
		},
	}
}

func getDashboardUrl(link *models.Hit) string {
	base, err := url.Parse(config.Config().GetDefaultGrafanaConfig().URL)
	var baseHost string
	if err != nil {
		baseHost = "http://unknown/"
		slog.Warn("unable to determine grafana base host for dashboard", slog.String("dashboard-uid", link.UID))
	} else {
		base.Path = ""
		baseHost = base.String()
	}
	return fmt.Sprintf("%s%s", baseHost, link.URL)
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
			cfg := config.Config().GetDefaultGrafanaConfig()
			rootCmd.TableObj.AppendHeader(table.Row{"id", "Title", "Slug", "Folder", "NestedPath", "UID", "Tags", "URL"})

			filters := service.NewDashboardFilter(parseDashboardGlobalFlags(cd.CobraCommand)...)
			boards := rootCmd.GrafanaSvc().ListDashboards(filters)

			printCount := func(count int) {
				slog.Info("Listing dashboards for context",
					slog.String("context", GetContext()),
					slog.String("orgName", GetOrganizationName()),
					slog.Any("count", count))
			}
			count := 0

			folders := rootCmd.GrafanaSvc().ListFolders(service.NewFolderFilter())
			for _, link := range boards {
				urlValue := getDashboardUrl(link)
				count++
				var tagVal string
				if len(link.Tags) > 0 {
					tagValByte, err := json.Marshal(link.Tags)
					if err == nil {
						tagVal = string(tagValByte)
					}
				}

				baseRow := table.Row{link.ID, link.Title, link.Slug, link.FolderTitle}
				if cfg.GetDashboardSettings().NestedFolders {
					baseRow = append(baseRow, service.GetNestedFolder(link.FolderTitle, link.FolderUID, folders))
				}
				baseRow = append(baseRow, table.Row{link.UID, tagVal, urlValue}...)
				rootCmd.TableObj.AppendRow(baseRow)

			}
			printCount(count)
			if len(boards) > 0 {
				rootCmd.Render(cd.CobraCommand, boards)
			} else {
				slog.Info("No dashboards found")
			}
			return nil
		},
	}
}
