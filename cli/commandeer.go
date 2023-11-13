package cli

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/backup"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/cli/tools"
	assets "github.com/esnet/gdg/config"
	"log/slog"
)

// Execute executes a command.
func Execute(defaultCfg string, args []string, options ...support.RootOption) error {
	data, err := assets.Assets.ReadFile(defaultCfg)
	if err != nil {
		slog.Info("unable to find load default configuration", "err", err)
	}
	support.DefaultConfig = string(data)
	rootCmd := support.NewRootCmd(getNewRootCmd(), options...)
	x, err := simplecobra.New(rootCmd)
	if err != nil {
		return err
	}

	cd, err := x.Execute(context.Background(), args)

	if err != nil || len(args) == 0 {
		_ = cd.CobraCommand.Help()
		return err
	}

	return nil
}

func getNewRootCmd() *support.RootCommand {
	return &support.RootCommand{
		NameP: "gdg",
		CommandEntries: []simplecobra.Commander{
			newVersionCmd(),
			newContextCmd(),
			tools.NewToolsCommand(),
			backup.NewBackupCommand(),
		},
	}
}
