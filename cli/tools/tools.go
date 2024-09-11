package tools

import (
	"context"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/spf13/cobra"
)

func NewToolsCommand() simplecobra.Commander {
	description := "A collection of tools to manage a grafana instance"
	return &support.SimpleCommand{
		NameP:        "tools",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{newContextCmd(), newDevelCmd(), newUserCommand(), newAuthCmd(), newOrgCommand(), newDashboardCmd()},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"t"}
		},
		InitCFunc: func(cd *simplecobra.Commandeer, r *support.RootCommand) error {
			support.InitConfiguration(cd.CobraCommand)
			return nil
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}
