package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"

	"github.com/esnet/gdg/api"
	"github.com/esnet/gdg/config"
	"github.com/spf13/cobra"
)

var (
	tableObj      table.Writer
	client        api.ApiService
	DefaultConfig string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
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
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration Override")

}

func initConfig() {
	configOverride, _ := rootCmd.Flags().GetString("config")
	if DefaultConfig == "" {
		raw, err := ioutil.ReadFile("conf/importer-example.yml")
		if err == nil {
			DefaultConfig = string(raw)
		} else {
			DefaultConfig = ""
		}
	}
	config.InitConfig(configOverride, DefaultConfig)

	setupGrafanaClient()
	log.Debug("Creating output locations")
	//Output Renderer
	tableObj = table.NewWriter()
	tableObj.SetOutputMirror(os.Stdout)
	tableObj.SetStyle(table.StyleLight)

	if config.Config().IsDebug() {
		log.SetLevel(log.DebugLevel)
	}
}

func setupGrafanaClient() {
	client = api.NewApiService()

}
