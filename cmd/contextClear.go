package cmd

import (
	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var contextClear = &cobra.Command{
	Use:   "clear",
	Short: "clear all context",
	Long:  `clear all contexts`,
	Run: func(cmd *cobra.Command, args []string) {
		apphelpers.ClearContexts()
		log.Info("Successfully deleted all configured contexts")
	},
}

func init() {
	context.AddCommand(contextClear)
}
