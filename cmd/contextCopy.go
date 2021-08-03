package cmd

import (
	"errors"

	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	"github.com/spf13/cobra"
)

var contextCopy = &cobra.Command{
	Use:     "copy",
	Short:   "copy context <src> <dest>",
	Long:    `copy contexts  <src> <dest>`,
	Aliases: []string{"cp"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires a src and destination argument")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		src := args[0]
		dest := args[1]
		apphelpers.CopyContext(src, dest)

	},
}

func init() {
	context.AddCommand(contextCopy)
}
