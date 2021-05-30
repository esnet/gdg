package cmd

import (
	"github.com/netsage-project/grafana-dashboard-manager/api"
	"github.com/spf13/cobra"
)

func getDashboardGlobalFlags(cmd *cobra.Command) api.Filter {
	folderFilter, _ := cmd.Flags().GetString("folder")
	dashboardFilter, _ := cmd.Flags().GetString("dashboard")

	filters := api.NewDashboardFilter()
	filters.AddFilter(api.FolderFilter, folderFilter)
	filters.AddFilter(api.DashFilter, dashboardFilter)

	return filters

}

var dashboard = &cobra.Command{
	Use:     "dashboards",
	Aliases: []string{"dash", "dashboard"},
	Short:   "Manage Dashboards",
	Long:    `Manage Grafana Dashboards.`,
}

func init() {
	rootCmd.AddCommand(dashboard)
	dashboard.PersistentFlags().StringP("dashboard", "d", "", "filter by dashboard slug")
	dashboard.PersistentFlags().StringP("folder", "f", "", "Filter by Folder Name (Quotes in names not supported)")
}
