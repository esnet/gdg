package cmd

import "github.com/spf13/cobra"

var develCmd = &cobra.Command{
	Use:   "devel",
	Short: "Developer Tooling",
	Long:  `Developer Tooling`,
}

func init() {
	rootCmd.AddCommand(develCmd)
}
