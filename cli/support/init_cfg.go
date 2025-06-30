package support

import (
	"log"
	"os"

	"github.com/esnet/gdg/internal/config"
	appconfig "github.com/esnet/gdg/internal/log"
	"github.com/spf13/cobra"
)

// InitConfiguration Loads configuration, and setups fail over case
func InitConfiguration(cmd *cobra.Command) {
	configOverride, _ := cmd.Flags().GetString("config")
	contextOverride, _ := cmd.Flags().GetString("context")
	// Registers sub CommandsList
	if config.Config() == nil {
		config.InitGdgConfig(configOverride)
	}
	if contextOverride != "" {
		cfg := config.Config()
		_, ok := cfg.GetGDGConfig().GetContexts()[contextOverride]
		if !ok {
			log.Fatalf("context %s was not found", contextOverride)
		}
		cfg.GetContexts()
		cfg.ChangeContext(contextOverride)
	}
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, config.Config().IsDebug())

	// Validate current configuration
	config.Config().GetDefaultGrafanaConfig().Validate()
}
