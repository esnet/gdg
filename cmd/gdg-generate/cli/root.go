package cli

import (
	assets "github.com/esnet/gdg/config"
	"github.com/esnet/gdg/internal/config"
	appconfig "github.com/esnet/gdg/internal/log"
	"github.com/esnet/gdg/internal/templating"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"os"
)

var (
	cfgFile    string
	tplCfgFile string
	template   templating.Templating
	rootCmd    = &cobra.Command{
		Use:   "gdg-generate",
		Short: "Generates dashboard templates for use with GDG given a valid configuration",
		Long:  `Generates dashboard templates for use with GDG given a valid configuration`,
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default: config/importer.yml)")
	rootCmd.PersistentFlags().StringP("template-config", "", "", "GDG Template configuration file override. (default: config/templates.yml)")
}

func initConfig() {
	var err error
	cfgFile, err = rootCmd.Flags().GetString("config")
	if err != nil {
		log.Fatal("unable to get config file")
	}
	tplCfgFile, err = rootCmd.Flags().GetString("template-config")
	if err != nil {
		log.Fatal("unable to get template config file")
	}

	defaultConfiguration, err := assets.GetFile("importer-example.yml")
	if err != nil {
		slog.Warn("unable to load default configuration, no fallback")
	}

	config.InitGdgConfig(cfgFile, defaultConfiguration)
	config.InitTemplateConfig(tplCfgFile)
	cfg := config.Config()
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, cfg.IsDebug())
	template = templating.NewTemplate()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
