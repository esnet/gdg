package support

import (
	"os"

	"github.com/esnet/gdg/internal/config"
	appconfig "github.com/esnet/gdg/internal/log"
	"github.com/spf13/cobra"
)

// InitConfiguration Loads configuration, and setups fail over case
func InitConfiguration(cmd *cobra.Command) {
	configOverride, _ := cmd.Flags().GetString("config")
	// Registers sub CommandsList
	if config.Config() == nil {
		config.InitGdgConfig(configOverride)
	}
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, config.Config().IsDebug())

	// Validate current configuration
	config.Config().GetDefaultGrafanaConfig().Validate()
}
