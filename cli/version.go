package cli

import (
	"context"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/version"
	"github.com/spf13/cobra"
)

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
		Long:  "All software has versions. This is generated code example",
	}
}
