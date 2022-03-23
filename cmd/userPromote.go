package cmd

import (
	"github.com/esnet/gdg/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var promoteUser = &cobra.Command{
	Use:     "promote",
	Short:   "Promote User to Grafana Admin",
	Long:    `Promote User to Grafana Admin`,
	Aliases: []string{"godmode"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		userLogin, _ := cmd.Flags().GetString("user")

		msg, err := client.PromoteUser(userLogin)
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(*msg.Message)
			log.Info("Please note user is a grafana admin, not necessarily an Org admin.  You may need to promote yourself manually per org")
		}

	},
}

func init() {
	userCmd.AddCommand(promoteUser)
	promoteUser.Flags().StringP("user", "u", "", "user email")
	err := promoteUser.MarkFlagRequired("user")
	if err != nil {
		log.Debug("Failed to mark user flag as required")
	}
}
