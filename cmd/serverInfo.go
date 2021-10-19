package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverInfo = &cobra.Command{
	Use:   "info",
	Short: "server health info",
	Long:  `server health info`,
	Run: func(cmd *cobra.Command, args []string) {
		result := client.GetServerInfo()
		for key, value := range result {
			log.Infof("%s:  %s", key, value)
		}
	},
}

func init() {
	server.AddCommand(serverInfo)
}
