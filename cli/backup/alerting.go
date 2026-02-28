package backup

import (
	"context"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/spf13/cobra"
)

func newAlertingCommand() simplecobra.Commander {
	description := "Manage Alerting resources"
	return &domain.SimpleCommand{
		NameP: "alerting",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"alert"}
		},
		CommandsList: []simplecobra.Commander{
			newAlertingContactCommand(),
			newAlertingRulesCommand(),
			newAlertingTemplatesCommand(),
			newAlertingNotificationCommand(),
			newAlertingTimingsCommand(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}
