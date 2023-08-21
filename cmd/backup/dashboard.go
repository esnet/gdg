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
	"strings"
)

var (
	skipConfirmAction bool
)

func parseDashboardGlobalFlags(command *cobra.Command) []string {
	folderFilter, _ := command.Flags().GetString("folder")
	dashboardFilter, _ := command.Flags().GetString("dashboard")
	tagsFilter, _ := command.Flags().GetStringSlice("tags")

	return []string{folderFilter, dashboardFilter, strings.Join(tagsFilter, ",")}
}

var dashboard = &cobra.Command{
	Use:     "dashboards",
	Aliases: []string{"dash", "dashboard"},
	Short:   "Manage Dashboards",
	Long:    `Manage Grafana Dashboards.`,
}

var clearDashboards = &cobra.Command{
	Use:     "clear",
	Short:   "delete all monitored dashboards from grafana",
	Long:    `clear all monitored dashboards from grafana`,
	Aliases: []string{"c"},
	Run: func(command *cobra.Command, args []string) {
		filter := service.NewDashboardFilter(parseDashboardGlobalFlags(command)...)
		deletedDashboards := cmd.GetGrafanaSvc().DeleteAllDashboards(filter)
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range deletedDashboards {
			cmd.TableObj.AppendRow(table.Row{"dashboard", file})
		}
		if len(deletedDashboards) == 0 {
			log.Info("No dashboards were found.  0 dashboards removed")

		} else {
			log.Infof("%d dashboards were deleted", len(deletedDashboards))
			cmd.TableObj.Render()
		}

	},
}

var uploadDashboard = &cobra.Command{
	Use:     "upload",
	Short:   "upload all dashboards to grafana",
	Long:    `upload all dashboards to grafana`,
	Aliases: []string{"u"},
	Run: func(command *cobra.Command, args []string) {

		filter := service.NewDashboardFilter(parseDashboardGlobalFlags(command)...)

		if !skipConfirmAction {
			tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will delete all dashboards from the monitored folders: '%s' "+
				"(or all folders if ignore_dashboard_filters is set to true) and upload your local copy.  Do you wish to "+
				"continue (y/n) ", strings.Join(config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders(), ", "),
			), "", true)
		}
		cmd.GetGrafanaSvc().UploadDashboards(filter)

		cmd.TableObj.AppendHeader(table.Row{"Title", "id", "folder", "UID"})
		boards := cmd.GetGrafanaSvc().ListDashboards(filter)

		for _, link := range boards {
			cmd.TableObj.AppendRow(table.Row{link.Title, link.ID, link.FolderTitle, link.UID})

		}
		if len(boards) > 0 {
			cmd.TableObj.Render()
		} else {
			log.Info("No dashboards found")
		}

	},
}

var downloadDashboard = &cobra.Command{
	Use:     "download",
	Short:   "download all dashboards from grafana",
	Aliases: []string{"d"},
	Long:    `Download all dashboards from grafana to local file system`,
	Run: func(command *cobra.Command, args []string) {
		filter := service.NewDashboardFilter(parseDashboardGlobalFlags(command)...)
		savedFiles := cmd.GetGrafanaSvc().DownloadDashboards(filter)
		log.Infof("Importing dashboards for context: '%s'", config.Config().GetAppConfig().GetContext())
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			cmd.TableObj.AppendRow(table.Row{"dashboard", file})
		}
		cmd.TableObj.Render()
	},
}

var listDashboards = &cobra.Command{
	Use:     "list",
	Short:   "List all dashboards from grafana",
	Long:    `List all dashboards from grafana`,
	Aliases: []string{"l"},
	Run: func(command *cobra.Command, args []string) {
		cmd.TableObj.AppendHeader(table.Row{"id", "Title", "Slug", "Folder", "UID", "Tags", "URL"})

		filters := service.NewDashboardFilter(parseDashboardGlobalFlags(command)...)
		boards := cmd.GetGrafanaSvc().ListDashboards(filters)

		log.Infof("Listing dashboards for context: '%s'", config.Config().GetAppConfig().GetContext())
		for _, link := range boards {
			url := fmt.Sprintf("%s%s", config.Config().GetDefaultGrafanaConfig().URL, link.URL)
			cmd.TableObj.AppendRow(table.Row{link.ID, link.Title, link.Slug, link.FolderTitle,
				link.UID, strings.Join(link.Tags, ","), url})

		}
		if len(boards) > 0 {
			cmd.TableObj.Render()
		} else {
			log.Info("No dashboards found")
		}

	},
}

func init() {
	backupCmd.AddCommand(dashboard)
	dashboard.PersistentFlags().BoolVarP(&skipConfirmAction, "skip-confirmation", "", false, "when set to true, bypass confirmation prompts")
	dashboard.PersistentFlags().StringP("dashboard", "d", "", "filter by dashboard slug")
	dashboard.PersistentFlags().StringP("folder", "f", "", "Filter by Folder Name (Quotes in names not supported)")
	dashboard.PersistentFlags().StringSliceP("tags", "t", []string{}, "Filter by Tags (does not apply on upload)")
	dashboard.AddCommand(clearDashboards)
	dashboard.AddCommand(uploadDashboard)
	dashboard.AddCommand(downloadDashboard)
	dashboard.AddCommand(listDashboards)
}
