package cmd

import (
	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:     "organizations",
	Aliases: []string{"org"},
	Short:   "Manage Organizations",
	Long:    `Manage Grafana Organizations.`,
}

func init() {
	rootCmd.AddCommand(orgCmd)
}
