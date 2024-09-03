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
	if DefaultConfig == "" {
		raw, err := os.ReadFile("config/importer-example.yml")
		if err == nil {
			DefaultConfig = string(raw)
		} else {
			DefaultConfig = ""
		}
	}

	// Registers sub CommandsList
	config.InitGdgConfig(configOverride, DefaultConfig)
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, config.Config().IsDebug())

	// Validate current configuration
	config.Config().GetDefaultGrafanaConfig().Validate()
}
