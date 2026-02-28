package cli

import (
	"context"
	"fmt"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/config"
	domain2 "github.com/esnet/gdg/pkg/version"
	"github.com/spf13/cobra"
)

// newVersionCmd creates a command that prints the application version information.
func newVersionCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "version",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			domain2.PrintVersionInfo()
			return nil
		},
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"v"}
		},
		Short: "Print the version number of generated code example",
		Long:  "Print the version number of generated code example",
	}
}

// newDefaultConfig returns a command that prints an example configuration.
func newDefaultConfig() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "default-config",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			fmt.Print(config.DefaultConfig())
			return nil
		},
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"v"}
		},
		Short: "Prints an example configuration",
		Long:  "Prints an example configuration",
	}
}
