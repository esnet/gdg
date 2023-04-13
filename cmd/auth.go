package cmd

import (
	"github.com/spf13/cobra"
)

// userCmd represents the version command
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage auth via API",
	Long:  `Provides some utility to help the user manage their auth keys`,
}

func init() {
	rootCmd.AddCommand(AuthCmd)
	AuthCmd.AddCommand(tokensCmd)

}
