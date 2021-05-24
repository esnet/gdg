package cmd

import (
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var datasources = &cobra.Command{
	Use:     "datasources",
	Aliases: []string{"ds", "datasource"},
	Short:   "Manage datasources",
	Long:    `All software has versions.`,
}

func init() {
	rootCmd.AddCommand(datasources)
}
