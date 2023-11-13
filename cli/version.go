package cli

import (
	"context"
	"fmt"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/version"
	"github.com/spf13/cobra"
	"os"
)

func newVersionCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "version",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *support.RootCommand, args []string) error {
			stdout := os.Stdout
			fmt.Fprintf(stdout, "Build Date: %s\n", version.BuildDate)
			fmt.Fprintf(stdout, "Git Commit: %s\n", version.GitCommit)
			fmt.Fprintf(stdout, "Version: %s\n", version.Version)
			fmt.Fprintf(stdout, "Go Version: %s\n", version.GoVersion)
			fmt.Fprintf(stdout, "OS / Arch: %s\n", version.OsArch)
			return nil
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"v"}
		},
		Short: "Print the version number of generated code example",
		Long:  "All software has versions. This is generated code example",
	}
}
