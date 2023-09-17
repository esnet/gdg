package cmd

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/backup"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/cmd/tools"
)

// Execute executes a command.
func Execute(defaultCfg string, args []string, options ...support.RootOption) error {
	support.DefaultConfig = defaultCfg
	rootCmd := support.NewRootCmd(getNewRootCmd(), options...)
	x, err := simplecobra.New(rootCmd)
	if err != nil {
		return err
	}

	cd, err := x.Execute(context.Background(), args)

	if err != nil || len(args) == 0 {
		return cd.CobraCommand.Help()
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
