package cmd

import (
	"fmt"
	"github.com/esnet/gdg/api"
	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func getDashboardGlobalFlags(cmd *cobra.Command) api.Filter {
	folderFilter, _ := cmd.Flags().GetString("folder")
	dashboardFilter, _ := cmd.Flags().GetString("dashboard")
	tagsFilter, _ := cmd.Flags().GetStringSlice("tags")

	filters := api.NewDashboardFilter()
	filters.AddFilter(api.FolderFilter, folderFilter)
	filters.AddFilter(api.DashFilter, dashboardFilter)
	filters.AddFilter(api.TagsFilter, strings.Join(tagsFilter, ","))

	return filters
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
		filter := getDashboardGlobalFlags(cmd)
		deletedDashboards := client.DeleteAllDashboards(filter)
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

		filter := getDashboardGlobalFlags(cmd)
		client.ExportDashboards(filter)

		tableObj.AppendHeader(table.Row{"Title", "id", "folder", "UID"})
		boards := client.ListDashboards(filter)

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
		filter := getDashboardGlobalFlags(cmd)
		savedFiles := client.ImportDashboards(filter)
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
		tableObj.AppendHeader(table.Row{"id", "Title", "Slug", "Folder", "UID", "URL"})

		filters := getDashboardGlobalFlags(cmd)
		boards := client.ListDashboards(filters)

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		for _, link := range boards {
			url := fmt.Sprintf("%s%s", apphelpers.GetCtxDefaultGrafanaConfig().URL, link.URL)
			tableObj.AppendRow(table.Row{link.ID, link.Title, link.Slug, link.FolderTitle,
				link.UID, url})

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
