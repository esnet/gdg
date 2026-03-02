package cli

import (
	"log"
	"log/slog"
	"os"

	appconfig "github.com/esnet/gdg/internal/adapter/logger"
	"github.com/esnet/gdg/internal/adapter/templating"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile        string
	tplCfgFile     string
	templateConfig *config_domain.TemplatingConfig
	template       ports.Templating
	rootCmd        = &cobra.Command{
		Use:   "gdg-generate",
		Short: "Generates dashboard templates for use with GDG given a valid configuration",
		Long:  `Generates dashboard templates for use with GDG given a valid configuration`,
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default: config/gdg.yml)")
	rootCmd.PersistentFlags().StringP("template-config", "", "", "GDG Template configuration file override. (default: config/templates.yml)")
}

func initConfig() {
	var err error
	slog.Info("Running gdg-generate", slog.Any("version", version.Version))
	cfgFile, err = rootCmd.Flags().GetString("config")
	if err != nil {
		log.Fatal("unable to get config file")
	}
	tplCfgFile, err = rootCmd.Flags().GetString("template-config")
	if err != nil {
		log.Fatal("unable to get template config file")
	}

	appCfg := config.NewConfig(cfgFile)
	templateConfig = config.InitTemplateConfig(tplCfgFile)
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, appCfg.IsDebug())
	template = templating.NewTemplate(templateConfig, appCfg.GetDefaultGrafanaConfig())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
