package tools

import (
	cmd "github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// userCmd represents the version command
var userCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user"},
	Short:   "Manage users",
	Long: `Provides some utility to manage grafana users from the CLI.  Please note, as the credentials cannot be imported, 
              the export with generate a default password for any user not already present`,
}

var promoteUser = &cobra.Command{
	Use:     "makeGrafanaAdmin",
	Short:   "Promote User to Grafana Admin",
	Long:    `Promote User to Grafana Admin`,
	Aliases: []string{"godmode", "promote"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Promoting User to Grafana Admin for context: '%s'", config.Config().AppConfig.GetContext())
		userLogin, _ := command.Flags().GetString("user")

		msg, err := cmd.GetGrafanaSvc().PromoteUser(userLogin)
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(msg)
			log.Info("Please note user is a grafana admin, not necessarily an Org admin.  You may need to promote yourself manually per org")
		}

	},
}

func init() {
	toolsCmd.AddCommand(userCmd)
	userCmd.AddCommand(promoteUser)
	promoteUser.Flags().StringP("user", "u", "", "user email")
	err := promoteUser.MarkFlagRequired("user")
	if err != nil {
		log.Debug("Failed to mark user flag as required")
	}
}
