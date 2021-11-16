package cmd

import (
	"errors"

	"github.com/netsage-project/gdg/apphelpers"
	"github.com/spf13/cobra"
)

var contextSet = &cobra.Command{
	Use:   "set",
	Short: "set <context>",
	Long:  `set <contexts>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a context argument")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		context := args[0]
		apphelpers.SetContext(context)

	},
}

func init() {
	context.AddCommand(contextSet)
}
