package cmd

import (
	"github.com/netsage-project/gdg/apphelpers"
	"github.com/spf13/cobra"
)

var contextShow = &cobra.Command{
	Use:   "show",
	Short: "show optional[context]",
	Long:  `show contexts optional[context]`,
	Run: func(cmd *cobra.Command, args []string) {
		context := apphelpers.GetContext()
		if len(args) > 1 && len(args[0]) > 0 {
			context = args[0]
		}
		apphelpers.ShowContext(context)

	},
}

func init() {
	context.AddCommand(contextShow)
}
