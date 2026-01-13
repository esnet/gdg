package support

import (
	"log"
	"os"

	"github.com/esnet/gdg/internal/config"
	appconfig "github.com/esnet/gdg/internal/log"
	"github.com/spf13/cobra"
)

// InitConfiguration Loads configuration, and setups fail over case
func (c *RootCommand) InitConfiguration(cmd *cobra.Command) {
	configOverride, _ := cmd.Flags().GetString("config")
	contextOverride, _ := cmd.Flags().GetString("context")
	// Registers sub CommandsList
	if c.configObj == nil {
		c.configObj = config.InitGdgConfig(configOverride)
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
}
