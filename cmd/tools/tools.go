package tools

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	"github.com/spf13/cobra"
)

func NewToolsCommand() simplecobra.Commander {
	description := "A collection of tools to manage a grafana instance"
	return &support.SimpleCommand{
		NameP:        "tools",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{newDevelCmd(), newUserCommand(), newAuthCmd(), newOrgCommand()},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"t"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}

}
