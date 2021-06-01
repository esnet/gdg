package cmd

import (
	"github.com/spf13/cobra"
)

// userCmd represents the version command
var userCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user"},
	Short:   "Manage users",
	Long:    `Manage users.`,
}

func init() {
	rootCmd.AddCommand(userCmd)
}
