package cli

import (
	"context"
	"fmt"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/version"
	"github.com/spf13/cobra"
)

// newVersionCmd creates a command that prints the application version information.
func newVersionCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "version",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *support.RootCommand, args []string) error {
			version.PrintVersionInfo()
			return nil
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"v"}
		},
		Short: "Print the version number of generated code example",
		Long:  "Print the version number of generated code example",
	}
}

// newDefaultConfig returns a command that prints an example configuration.
func newDefaultConfig() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "default-config",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *support.RootCommand, args []string) error {
			o := config.Configuration{}
			fmt.Print(o.DefaultConfig())
			return nil
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"v"}
		},
		Short: "Prints an example configuration",
		Long:  "Prints an example configuration",
	}
}
