package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/bep/simplecobra"
	appconfig "github.com/esnet/gdg/internal/adapter/logger"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// RootCommand struct wraps the root command and supporting services needed
type RootCommand struct {
	NameP  string
	isInit bool

	configObj *config_domain.GDGAppConfiguration
	app       ports.GrafanaService

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
func (c *RootCommand) SetUpTest(app ports.GrafanaService) {
	c.app = app
	c.configObj = config.NewConfig("testing")
}

// GrafanaSvc returns the configured GrafanaService instance, initializing it if nil.
func (c *RootCommand) GrafanaSvc() ports.GrafanaService {
	return c.app
}

// ConfigSvc returns the root command's configuration object.
func (c *RootCommand) ConfigSvc() *config_domain.GDGAppConfiguration {
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

// ApplyOptions Allows to construct a root command passing any number of arguments to set RootCommand Options
func (c *RootCommand) ApplyOptions(options ...RootOption) error {
	if c == nil {
		return errors.New("unable to apply options on nil RootCommand")
	}
	for _, o := range options {
		o(c)
	}
	return nil
}

func NewRootCommand(name string) *RootCommand {
	return &RootCommand{
		NameP: name,
	}
}

func (c *RootCommand) SetService(svc ports.GrafanaService) {
	c.app = svc
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

// LoadConfig Loads configuration, and setups fail over case
func (c *RootCommand) LoadConfig(configOverride, contextOverride string) *config_domain.GDGAppConfiguration {
	// Registers sub CommandsList
	if c.configObj == nil {
		c.configObj = config.NewConfig(configOverride)
	}
	if contextOverride != "" {
		_, ok := c.configObj.GetContexts()[contextOverride]
		if !ok {
			log.Fatalf("context %s was not found", contextOverride)
		}

		c.configObj.SetContext(contextOverride)
	}

	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, c.configObj.IsDebug())
	c.configObj.GetDefaultGrafanaConfig().Validate()
	if c.configObj.GetAppGlobals().ApiDebug {
		err := os.Setenv("DEBUG", "1")
		if err != nil {
			slog.Debug("unable to set debug env value", slog.Any("err", err))
		}
	} else {
		err := os.Setenv("DEBUG", "0")
		if err != nil {
			slog.Debug("unable to set debug env value", slog.Any("err", err))
		}
	}
	return c.configObj
}
