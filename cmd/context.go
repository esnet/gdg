package cmd

import (
	"github.com/spf13/cobra"
)

var context = &cobra.Command{
	Use:     "contexts",
	Aliases: []string{"ctx", "context"},
	Short:   "Manage Context configuration",
	Long:    `Manage Context configuration which allows multiple grafana configs to be used.`,
}

func init() {
	rootCmd.AddCommand(context)
}
