package cmd

import (
	"github.com/spf13/cobra"
)

var dashboard = &cobra.Command{
	Use:     "dashboards",
	Aliases: []string{"dash"},
	Short:   "Manage Dashboards",
	Long:    `Manage Grafana Dashboards.`,
}

func init() {
	rootCmd.AddCommand(dashboard)
}
