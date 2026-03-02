package tools

import (
	"context"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/spf13/cobra"
)

func newUserCommand() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "users",
		Short: "Manage users",
		Long:  "Provides some utility to manage grafana users from the CLI.  Please note, as the credentials cannot be imported, the export with generate a default password for any user not already present",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"u", "user"}
		},
		InitCFunc:    nil,
		CommandsList: []simplecobra.Commander{newPromoteUserCmd()},
	}
}

func newPromoteUserCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "makeGrafanaAdmin",
		Short: "Promote User to Grafana Admin",
		Long:  "Promote User to Grafana Admin",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			slog.Info("Promoting User to Grafana Admin for context: '%s'", "context", rootCmd.ConfigSvc().GetContext())
			userLogin, _ := cd.CobraCommand.Flags().GetString("user")

			msg, err := rootCmd.GrafanaSvc().PromoteUser(userLogin)
			if err != nil {
				slog.Error(err.Error())
			} else {
				slog.Info(msg)
				slog.Info("Please note user is a grafana admin, not necessarily an Org admin.  You may need to promote yourself manually per org")
			}
			return nil
		},
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"godmode", "promote"}
			cmd.Flags().StringP("user", "u", "", "user email")
			err := cmd.MarkFlagRequired("user")
			if err != nil {
				slog.Debug("Failed to mark user flag as required")
			}
		},
		InitCFunc:    nil,
		CommandsList: nil,
	}
}
