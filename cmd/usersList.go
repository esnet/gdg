package cmd

import (
	"strings"

	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/api"
	"github.com/netsage-project/grafana-dashboard-manager/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listUserCmd = &cobra.Command{
	Use:   "list",
	Short: "list users",
	Long:  `list users`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing dashboards for context: '%s'", config.GetContext())
		tableObj.AppendHeader(table.Row{"id", "login", "name", "email", "admin", "grafanaAdmin", "disabled", "authLabels"})
		users := api.ListUsers(adminClient)
		if len(users) == 0 {
			log.Info("No users found")
		} else {
			for _, user := range users {
				var labels string
				if len(user.AuthLabels) > 0 {
					labels = strings.Join(user.AuthLabels, ", ")

				}
				tableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsAdmin, user.IsAdmin, user.IsDisabled, labels})
			}
			tableObj.Render()
		}

	},
}

func init() {
	userCmd.AddCommand(listUserCmd)
}
