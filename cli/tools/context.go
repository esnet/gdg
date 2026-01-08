package tools

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

func newContextCmd() simplecobra.Commander {
	v := &support.SimpleCommand{
		NameP: "contexts",
		CommandsList: []simplecobra.Commander{
			newContextClearCmd(),
			newListContextCmd(),
			newContextCopy(),
			newShowContext(),
			newDeleteContext(),
			newContext(),
			newSetContext(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		Short: "Manage Context configuration",
		Long:  "Manage Context configuration which allows multiple grafana configs to be used.",
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"ctx", "context"}
		},
	}
	return v
}

func newContextClearCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "clear",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *support.RootCommand, args []string) error {
			config.ClearContexts(r.ConfigSvc())
			slog.Info("Successfully deleted all configured contexts")
			return nil
		},
		Short: "Manage Context configuration",
		Long:  "Manage Context configuration which allows multiple grafana configs to be used.",
	}
}

func newListContextCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "list",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"context", "active"})
			contexts := rootCmd.ConfigSvc().GetContexts()
			activeContext := rootCmd.ConfigSvc().GetContext()
			for key := range contexts {
				active := false
				if key == strings.ToLower(activeContext) {
					key = fmt.Sprintf("*%s", activeContext)
					active = true
				}
				rootCmd.TableObj.AppendRow(table.Row{key, active})
			}

			rootCmd.Render(cd.CobraCommand, contexts)

			return nil
		},
		Short: "List context",
		Long:  "List contexts.",
	}
}

func newContextCopy() simplecobra.Commander {
	v := &support.SimpleCommand{
		NameP: "copy",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			src := args[0]
			dest := args[1]
			config.CopyContext(rootCmd.ConfigSvc(), src, dest)
			return nil
		},
		InitCFunc: func(cd *simplecobra.Commandeer, r *support.RootCommand) error {
			cd.CobraCommand.Aliases = []string{"cp"}
			cd.CobraCommand.Args = func(cmd *cobra.Command, args []string) error {
				if len(args) < 2 {
					return errors.New("requires a src and destination argument")
				}
				return nil
			}
			return nil
		},
		Short: "copy context <src> <dest>",
		Long:  "copy contexts  <src> <dest>",
	}

	return v
}

func newShowContext() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "show",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			contextEntry := rootCmd.ConfigSvc().GetContext()
			if len(args) > 0 && len(args[0]) > 0 {
				contextEntry = args[0]
			}
			rootCmd.ConfigSvc().PrintContext(contextEntry)
			return nil
		},
		Short: "show optional[context]",
		Long:  "show optional[context]",
	}
}

func newDeleteContext() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "delete",
		Short: "delete context <context>",
		Long:  "delete context <context>",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a context argument")
			}
			contextEntry := args[0]
			config.DeleteContext(rootCmd.ConfigSvc(), contextEntry)
			slog.Info("Successfully deleted context", "context", contextEntry)
			return nil
		},
		InitCFunc: func(cd *simplecobra.Commandeer, r *support.RootCommand) error {
			cd.CobraCommand.Aliases = []string{"del"}
			return nil
		},
	}
}

func newContext() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "new",
		Short: "new <context>",
		Long:  "new <context>",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a context NameP")
			}
			contextEntry := args[0]
			config.CreateNewContext(rootCmd.ConfigSvc(), contextEntry)
			return nil
		},
	}
}

func newSetContext() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "set",
		Short: "set <context>",
		Long:  "set <context>",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a context argument")
			}
			contextEntry := args[0]
			rootCmd.ConfigSvc().ChangeContext(contextEntry)
			return nil
		},
	}
}
