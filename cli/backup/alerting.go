package backup

import (
	"context"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/spf13/cobra"
)

func newAlertingCommand() simplecobra.Commander {
	description := "Manage Alerting resources"
	return &support.SimpleCommand{
		NameP: "alerting",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"alert"}
		},
		CommandsList: []simplecobra.Commander{
			newAlertingContactCommand(),
			newAlertingRulesCommand(),
			newAlertingTemplatesCommand(),
			newAlertingNotificationCommand(),
			newAlertingTimingsCommand(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}
