package support

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/bep/simplecobra"
	appconfig "github.com/esnet/gdg/internal/log"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// RootCommand struct wraps the root command and supporting services needed
type RootCommand struct {
	NameP  string
	isInit bool

	GrafanaSvc func() service.GrafanaService

	ctx                  context.Context
	initThis             *simplecobra.Commandeer
	initRunner           *simplecobra.Commandeer
	failWithCobraCommand bool
	failRun              bool

	TableObj table.Writer

	CommandEntries []simplecobra.Commander
}

// Render outputs data as JSON if --output=json, otherwise renders a table.
func (cmd *RootCommand) Render(command *cobra.Command, data any) {
	output, _ := command.Flags().GetString("output")
	if output == "json" {
		data, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Fatal("unable to render result to JSON", err)
		}
		fmt.Print(string(data))

	} else {
		cmd.TableObj.Render()
	}
}

// RootOption used to configure the Root Command struct
type RootOption func(command *RootCommand)

// NewRootCmd Allows to construct a root command passing any number of arguments to set RootCommand Options
func NewRootCmd(root *RootCommand, options ...RootOption) *RootCommand {
	if root == nil {
		root = &RootCommand{}
	}
	for _, o := range options {
		o(root)
	}
	return root
}

// Commands returns a list of Cobra commands
func (c *RootCommand) Commands() []simplecobra.Commander {
	return c.CommandEntries
}

// PreRun executed prior to command invocation
func (c *RootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	c.isInit = true
	c.initThis = this
	c.initRunner = runner
	c.initConfiguration()
	return nil
}

// initConfiguration Loads configuration, and setups fail over case
func (c *RootCommand) initConfiguration() {
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, false)
}

// Name returns the cli command name
func (c *RootCommand) Name() string {
	return c.NameP
}

// Run invokes the CLI command
func (c *RootCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	if c.failRun {
		return errors.New("failRun")
	}
	c.ctx = ctx
	return nil
}

// Init invoked to Initialize the RootCommand object
func (c *RootCommand) Init(cd *simplecobra.Commandeer) error {
	if c.failWithCobraCommand {
		return errors.New("failWithCobraCommand")
	}
	cmd := cd.CobraCommand

	persistentFlags := cmd.PersistentFlags()
	persistentFlags.StringP("config", "c", "", "Configuration Override")
	persistentFlags.StringP("context", "", "", "Context Override")
	persistentFlags.StringP("output", "", "table", "output format: (table, json)")
	if c.TableObj == nil {
		c.TableObj = table.NewWriter()
		c.TableObj.SetOutputMirror(os.Stdout)
		c.TableObj.SetStyle(table.StyleLight)
	}

	return nil
}
