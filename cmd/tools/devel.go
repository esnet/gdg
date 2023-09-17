package tools

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func newDevelCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP:        "devel",
		Short:        "Developer Tooling",
		Long:         "Developer Tooling",
		CommandsList: []simplecobra.Commander{newServerInfoCmd(), newCompletion()},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newServerInfoCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "srvinfo",
		Short: "server health info",
		Long:  "server health info",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			result := rootCmd.GrafanaSvc().GetServerInfo()
			for key, value := range result {
				log.Infof("%s:  %s", key, value)
			}
			return nil
		},
	}
}

func newCompletion() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long:  "Generate completion script",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			var err error
			switch args[0] {
			case "bash":
				err = cd.CobraCommand.GenBashCompletion(os.Stdout)
			case "zsh":
				err = cd.CobraCommand.GenZshCompletion(os.Stdout)
			case "fish":
				err = cd.CobraCommand.GenFishCompletion(os.Stdout, true)
			case "powershell":
				err = cd.CobraCommand.GenPowerShellCompletion(os.Stdout)
			}
			return err
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.ValidArgs = []string{"bash", "zsh", "fish", "powershell"}
			cmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

		},
	}
}
