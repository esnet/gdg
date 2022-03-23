package cmd

import (
	"errors"

	"github.com/esnet/gdg/apphelpers"
	"github.com/spf13/cobra"
)

var contextNew = &cobra.Command{
	Use:   "new",
	Short: "new <context>",
	Long:  `new <context>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a context name")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		ctx := args[0]
		apphelpers.NewContext(ctx)

	},
}

func init() {
	context.AddCommand(contextNew)
}
