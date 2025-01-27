package tools

import (
	"context"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
)

func newAuthCmd() simplecobra.Commander {
	description := "Manage auth via API"
	return &support.SimpleCommand{
		NameP:        "auth",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{newServiceAccountCmd()},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}
