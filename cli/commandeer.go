package cli

import (
	"context"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/backup"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/cli/tools"
)

// Execute runs the root command with given args and optional RootOptions, returning any error.
// It constructs the root command, executes it via simplecobra, and displays help on failure.
func Execute(args []string, options ...support.RootOption) error {
	var err error
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

// getNewRootCmd creates the root command with name "gdg" and subcommands for version,
// default config, tools, and backup utilities.
func getNewRootCmd() *support.RootCommand {
	return &support.RootCommand{
		NameP: "gdg",
		CommandEntries: []simplecobra.Commander{
			newVersionCmd(),
			newDefaultConfig(),
			tools.NewToolsCommand(),
			backup.NewBackupCommand(),
		},
	}
}
