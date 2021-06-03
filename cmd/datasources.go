package cmd

import (
	"github.com/netsage-project/grafana-dashboard-manager/api"
	"github.com/spf13/cobra"
)

func getDatasourcesGlobalFlags(cmd *cobra.Command) api.DatasourceFilter {
	dashboardFilter, _ := cmd.Flags().GetString("datasource")

	filters := api.DatasourceFilter{
		Name: dashboardFilter,
	}

	return filters

}

// versionCmd represents the version command
var datasources = &cobra.Command{
	Use:     "datasources",
	Aliases: []string{"ds", "datasource"},
	Short:   "Manage datasources",
	Long:    `All software has versions.`,
}

func init() {
	rootCmd.AddCommand(datasources)
	datasources.PersistentFlags().StringP("datasource", "d", "", "filter by datasource slug")

}
