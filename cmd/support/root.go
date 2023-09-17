package support

import (
	"context"
	"errors"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	DefaultConfig string
)

type RootCommand struct {
	NameP  string
	isInit bool

	GrafanaSvc func() service.GrafanaService

	localFlagName string

	persistentFlagNameC string
	localFlagNameC      string

	ctx                  context.Context
	initThis             *simplecobra.Commandeer
	initRunner           *simplecobra.Commandeer
	failWithCobraCommand bool
	failRun              bool

	TableObj table.Writer

	CommandEntries []simplecobra.Commander
}

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

func (c *RootCommand) Commands() []simplecobra.Commander {
	return c.CommandEntries
}

func (c *RootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	c.isInit = true
	c.initThis = this
	c.initRunner = runner
	c.initConfiguration()
	return nil
}

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

	if config.Config().IsDebug() {
		log.SetLevel(log.DebugLevel)
	}
	//Validate current configuration
	config.Config().GetDefaultGrafanaConfig().Validate()

}

func (c *RootCommand) Name() string {
	return c.NameP
}

func (c *RootCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	if c.failRun {
		return errors.New("failRun")
	}
	c.ctx = ctx
	return nil
}

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
