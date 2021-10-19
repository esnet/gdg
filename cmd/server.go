package cmd

import (
	"github.com/spf13/cobra"
)

var server = &cobra.Command{
	Use:     "server",
	Aliases: []string{"srv", "servers"},
	Short:   "Get Server Info",
	Long:    `Retrieve Server info`,
}

func init() {
	rootCmd.AddCommand(server)
}
