package support

import (
	"context"
	"errors"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/internal/config"
	appconfig "github.com/esnet/gdg/internal/log"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

var (
	DefaultConfig string
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
	cmd := c.initRunner.CobraCommand
	configOverride, _ := cmd.Flags().GetString("config")
	if DefaultConfig == "" {
		raw, err := os.ReadFile("config/importer-example.yml")
		if err == nil {
			DefaultConfig = string(raw)
		} else {
			DefaultConfig = ""
		}
	}
	//Registers sub CommandsList
	config.InitConfig(configOverride, DefaultConfig)
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr)

	//Validate current configuration
	config.Config().GetDefaultGrafanaConfig().Validate()

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
	if c.TableObj == nil {
		c.TableObj = table.NewWriter()
		c.TableObj.SetOutputMirror(os.Stdout)
		c.TableObj.SetStyle(table.StyleLight)
	}

	return nil
}
