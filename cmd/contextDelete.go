package cmd

import (
	"errors"

	"github.com/netsage-project/gdg/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var contextDelete = &cobra.Command{
	Use:     "delete",
	Short:   "delete context <context>",
	Long:    `delete context <context>.`,
	Aliases: []string{"del"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a context argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := args[0]
		apphelpers.DeleteContext(ctx)
		log.Infof("Successfully deleted context %s", ctx)
	},
}

func init() {
	context.AddCommand(contextDelete)
}
