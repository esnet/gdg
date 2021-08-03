package cmd

import (
	"fmt"
	"os"

	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var contextShow = &cobra.Command{
	Use:   "show",
	Short: "show optional[context]",
	Long:  `show contexts optional[context]`,
	Run: func(cmd *cobra.Command, args []string) {
		context := apphelpers.GetContext()
		if len(args) > 1 && len(args[0]) > 0 {
			context = args[0]
		}

		grafana := apphelpers.GetCtxGrafanaConfig(context)
		d, err := yaml.Marshal(grafana)
		if err != nil {
			log.Info("Failed to serialize context")
			os.Exit(1)
		}
		fmt.Printf("---%s:\n%s\n\n", context, string(d))

	},
}

func init() {
	context.AddCommand(contextShow)
}
