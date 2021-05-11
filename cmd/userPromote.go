package cmd

import (
	"github.com/netsage-project/grafana-dashboard-manager/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var promoteUser = &cobra.Command{
	Use:     "promote",
	Short:   "Promote User to Grafana Admin",
	Long:    `Promote User to Grafana Admin`,
	Aliases: []string{"godmode"},
	Run: func(cmd *cobra.Command, args []string) {

		userLogin, _ := cmd.Flags().GetString("user")

		msg, err := api.PromoteUser(client, userLogin)
		if err != nil {
			logrus.Error(err.Error())
		} else {
			logrus.Info(*msg.Message)
			logrus.Info("Please not user is a grafana admin, not necessarily an Org admin.  You may need to promote yourself manually per org")
		}

	},
}

func init() {
	userCmd.AddCommand(promoteUser)
	promoteUser.Flags().StringP("user", "u", "", "user email")
	promoteUser.MarkFlagRequired("user")
}
