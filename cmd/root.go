package cmd

import (
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

var (
	TableObj      table.Writer
	grafanaSvc    service.GrafanaService
	DefaultConfig string
	once          sync.Once
)

// GetGrafanaSvc returns the GrafanaService
func GetGrafanaSvc() service.GrafanaService {
	if grafanaSvc == nil {
		initConfig()
	}
	return grafanaSvc
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gdg Grafana Dash-N-Grab",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//      Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	RootCmd.PersistentFlags().StringP("config", "c", "", "Configuration Override")

}

func initConfig() {
	once.Do(func() {
		configOverride, _ := RootCmd.Flags().GetString("config")
		if DefaultConfig == "" {
			raw, err := os.ReadFile("config/importer-example.yml")
			if err == nil {
				DefaultConfig = string(raw)
			} else {
				DefaultConfig = ""
			}
		}
		config.InitConfig(configOverride, DefaultConfig)

		grafanaSvc = service.NewApiService()
		//Output Renderer
		TableObj = table.NewWriter()
		TableObj.SetOutputMirror(os.Stdout)
		TableObj.SetStyle(table.StyleLight)

		if config.Config().IsDebug() {
			log.SetLevel(log.DebugLevel)
		}
		//Validate current configuration
		config.Config().GetDefaultGrafanaConfig().Validate()
	})

}
