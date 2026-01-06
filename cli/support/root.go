package support

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/domain"
	appconfig "github.com/esnet/gdg/internal/log"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// RootCommand struct wraps the root command and supporting services needed
type RootCommand struct {
	NameP  string
	isInit bool

	configObj *domain.GDGAppConfiguration
	app       service.GrafanaService

	ctx                  context.Context
	initThis             *simplecobra.Commandeer
	initRunner           *simplecobra.Commandeer
	failWithCobraCommand bool
	failRun              bool

	TableObj table.Writer

	CommandEntries []simplecobra.Commander
}

// SetUpTest initializes the RootCommand for testing by setting a mock GrafanaService and loading test configuration.
// It only runs when the TESTING environment variable is set to "1".
func (c *RootCommand) SetUpTest(app service.GrafanaService) {
	if os.Getenv("TESTING") != "1" {
		return
	}

	c.app = app
	c.configObj = config.InitGdgConfig("testing")
}

// GrafanaSvc returns the configured GrafanaService instance, initializing it if nil.
func (c *RootCommand) GrafanaSvc() service.GrafanaService {
	if c.app == nil {
		c.app = service.NewDashNGo(c.configObj)
	}
	return c.app
}

// ConfigSvc returns the root command's configuration object.
func (c *RootCommand) ConfigSvc() *domain.GDGAppConfiguration {
	return c.configObj
}

// Render outputs data as JSON if --output=json, otherwise renders a table.
func (c *RootCommand) Render(command *cobra.Command, data any) {
	output, _ := command.Flags().GetString("output")
	if output == "json" {
		data, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Fatal("unable to render result to JSON", err)
		}
		fmt.Print(string(data))

	} else {
		c.TableObj.Render()
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
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, false)
	return nil
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
