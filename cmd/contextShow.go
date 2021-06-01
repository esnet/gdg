package cmd

import (
	"fmt"
	"os"

	"github.com/netsage-project/grafana-dashboard-manager/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var contextShow = &cobra.Command{
	Use:   "show",
	Short: "show context",
	Long:  `show contexts.`,
	Run: func(cmd *cobra.Command, args []string) {
		context, _ := cmd.Flags().GetString("context")
		grafana := config.GetGrafanaConfig(context)
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
	contextShow.Flags().StringP("context", "c", "", "context")
	contextShow.MarkFlagRequired("context")
}
