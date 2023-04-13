package cmd

import (
	"fmt"
	"github.com/esnet/gdg/internal/apphelpers"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func parseDashboardGlobalFlags(cmd *cobra.Command) []string {
	folderFilter, _ := cmd.Flags().GetString("folder")
	dashboardFilter, _ := cmd.Flags().GetString("dashboard")
	tagsFilter, _ := cmd.Flags().GetStringSlice("tags")

	return []string{folderFilter, dashboardFilter, strings.Join(tagsFilter, ",")}
}

var dashboard = &cobra.Command{
	Use:     "dashboards",
	Aliases: []string{"dash", "dashboard"},
	Short:   "Manage Dashboards",
	Long:    `Manage Grafana Dashboards.`,
}

var clearDashboards = &cobra.Command{
	Use:   "clear",
	Short: "delete all monitored dashboards",
	Long:  `clear all monitored dashboards from grafana`,
	Run: func(cmd *cobra.Command, args []string) {
		filter := service.NewDashboardFilter(parseDashboardGlobalFlags(cmd)...)
		deletedDashboards := grafanaSvc.DeleteAllDashboards(filter)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range deletedDashboards {
			tableObj.AppendRow(table.Row{"dashboard", file})
		}
		if len(deletedDashboards) == 0 {
			log.Info("No dashboards were found.  0 dashboards removed")

		} else {
			log.Infof("%d dashboards were deleted", len(deletedDashboards))
			tableObj.Render()
		}

	},
}

var exportDashboard = &cobra.Command{
	Use:   "export",
	Short: "export all dashboards",
	Long:  `export all dashboards`,
	Run: func(cmd *cobra.Command, args []string) {

		filter := service.NewDashboardFilter(parseDashboardGlobalFlags(cmd)...)
		grafanaSvc.ExportDashboards(filter)

		tableObj.AppendHeader(table.Row{"Title", "id", "folder", "UID"})
		boards := grafanaSvc.ListDashboards(filter)

		for _, link := range boards {
			tableObj.AppendRow(table.Row{link.Title, link.ID, link.FolderTitle, link.UID})

		}
		if len(boards) > 0 {
			tableObj.Render()
		} else {
			log.Info("No dashboards found")
		}

	},
}

var importDashboard = &cobra.Command{
	Use:   "import",
	Short: "Import all dashboards",
	Long:  `Import all dashboards from grafana to local file system`,
	Run: func(cmd *cobra.Command, args []string) {
		filter := service.NewDashboardFilter(parseDashboardGlobalFlags(cmd)...)
		savedFiles := grafanaSvc.ImportDashboards(filter)
		log.Infof("Importing dashboards for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"dashboard", file})
		}
		tableObj.Render()
	},
}

var listDashboards = &cobra.Command{
	Use:   "list",
	Short: "List all dashboards",
	Long:  `List all dashboards`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "Title", "Slug", "Folder", "UID", "Tags", "URL"})

		filters := service.NewDashboardFilter(parseDashboardGlobalFlags(cmd)...)
		boards := grafanaSvc.ListDashboards(filters)

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		for _, link := range boards {
			url := fmt.Sprintf("%s%s", apphelpers.GetCtxDefaultGrafanaConfig().URL, link.URL)
			tableObj.AppendRow(table.Row{link.ID, link.Title, link.Slug, link.FolderTitle,
				link.UID, strings.Join(link.Tags, ","), url})

		}
		if len(boards) > 0 {
			tableObj.Render()
		} else {
			log.Info("No dashboards found")
		}

	},
}

func init() {
	rootCmd.AddCommand(dashboard)
	dashboard.PersistentFlags().StringP("dashboard", "d", "", "filter by dashboard slug")
	dashboard.PersistentFlags().StringP("folder", "f", "", "Filter by Folder Name (Quotes in names not supported)")
	dashboard.PersistentFlags().StringSliceP("tags", "t", []string{}, "Filter by Tags (does not apply on export)")
	dashboard.AddCommand(clearDashboards)
	dashboard.AddCommand(exportDashboard)
	dashboard.AddCommand(importDashboard)
	dashboard.AddCommand(listDashboards)
}
