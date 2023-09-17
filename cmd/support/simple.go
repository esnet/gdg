package support

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/spf13/cobra"
)

type SimpleCommand struct {
	use       string
	NameP     string
	Short     string
	Long      string
	RunFunc   func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *RootCommand, args []string) error
	WithCFunc func(cmd *cobra.Command, r *RootCommand)
	InitCFunc func(cd *simplecobra.Commandeer, r *RootCommand) error

	CommandsList []simplecobra.Commander

	RootCmd *RootCommand
}

func (c *SimpleCommand) Commands() []simplecobra.Commander {
	return c.CommandsList
}

func (c *SimpleCommand) SetName(name string) {
	c.NameP = name
}

func (c *SimpleCommand) Name() string {
	return c.NameP
}

func (c *SimpleCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	if c.RunFunc == nil {
		return nil
	}
	return c.RunFunc(ctx, cd, c.RootCmd, args)
}

func (c *SimpleCommand) Init(cd *simplecobra.Commandeer) error {
	c.RootCmd = cd.Root.Command.(*RootCommand)
	cmd := cd.CobraCommand
	cmd.Short = c.Short
	cmd.Long = c.Long
	if c.use != "" {
		cmd.Use = c.use
	}
	if c.WithCFunc != nil {
		c.WithCFunc(cmd, c.RootCmd)
	}
	return nil
}

func (c *SimpleCommand) PreRun(cd, runner *simplecobra.Commandeer) error {
	if c.InitCFunc != nil {
		return c.InitCFunc(cd, c.RootCmd)
	}
	return nil
}
