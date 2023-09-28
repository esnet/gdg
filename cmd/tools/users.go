package tools

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newUserCommand() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "users",
		Short: "Manage users",
		Long:  "Provides some utility to manage grafana users from the CLI.  Please note, as the credentials cannot be imported, the export with generate a default password for any user not already present",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()

		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u", "user"}
		},
		InitCFunc:    nil,
		CommandsList: []simplecobra.Commander{newPromoteUserCmd()},
	}

}

func newPromoteUserCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "makeGrafanaAdmin",
		Short: "Promote User to Grafana Admin",
		Long:  "Promote User to Grafana Admin",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Promoting User to Grafana Admin for context: '%s'", config.Config().AppConfig.GetContext())
			userLogin, _ := cd.CobraCommand.Flags().GetString("user")

			msg, err := rootCmd.GrafanaSvc().PromoteUser(userLogin)
			if err != nil {
				log.Error(err.Error())
			} else {
				log.Info(msg)
				log.Info("Please note user is a grafana admin, not necessarily an Org admin.  You may need to promote yourself manually per org")
			}
			return nil

		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"godmode", "promote"}
			cmd.Flags().StringP("user", "u", "", "user email")
			err := cmd.MarkFlagRequired("user")
			if err != nil {
				log.Debug("Failed to mark user flag as required")
			}
		},
		InitCFunc:    nil,
		CommandsList: nil,
	}
}
