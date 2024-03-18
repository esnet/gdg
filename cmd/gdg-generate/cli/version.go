package cli

import (
	"github.com/esnet/gdg/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gdg-generate",
	Long:  `Print the version number of gdg-generate`,
	Run: func(cmd *cobra.Command, args []string) {
		version.PrintVersionInfo()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
