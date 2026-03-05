package tools

import (
	"context"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
)

func newDevelCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP:        "devel",
		Short:        "Developer Tooling",
		Long:         "Developer Tooling",
		CommandsList: []simplecobra.Commander{newServerInfoCmd()},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newServerInfoCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "srvinfo",
		Short: "server health info",
		Long:  "server health info",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			result := rootCmd.GrafanaSvc().GetServerInfo()
			for key, value := range result {
				slog.Info("", key, value)
			}
			return nil
		},
	}
}
