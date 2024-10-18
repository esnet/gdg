package cli

import (
	"context"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/backup"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/cli/tools"
	assets "github.com/esnet/gdg/config"
)

// Execute executes a command.
func Execute(defaultCfg string, args []string, options ...support.RootOption) error {
	var err error
	support.DefaultConfig, err = assets.GetFile(defaultCfg)
	if err != nil {
		slog.Warn("unable to find load default configuration", "err", err)
	}
	rootCmd := support.NewRootCmd(getNewRootCmd(), options...)
	x, err := simplecobra.New(rootCmd)
	if err != nil {
		return err
	}

	cd, err := x.Execute(context.Background(), args)

	if err != nil || len(args) == 0 {
		if cd != nil {
			_ = cd.CobraCommand.Help()
		}
		return err
	}

	return nil
}

func getNewRootCmd() *support.RootCommand {
	return &support.RootCommand{
		NameP: "gdg",
		CommandEntries: []simplecobra.Commander{
			newVersionCmd(),
			tools.NewToolsCommand(),
			backup.NewBackupCommand(),
		},
	}
}
