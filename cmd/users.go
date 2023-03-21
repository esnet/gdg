package cmd

import (
	"github.com/esnet/gdg/api"
	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
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
			log.Info(msg)
			log.Info("Please note user is a grafana admin, not necessarily an Org admin.  You may need to promote yourself manually per org")
		}

	},
}
var deleteUsersCmd = &cobra.Command{
	Use:   "clear",
	Short: "delete all users",
	Long:  `delete all users`,
	Run: func(cmd *cobra.Command, args []string) {

		savedFiles := client.DeleteAllUsers()
		log.Infof("Delete Users for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"type", "filename"})
		if len(savedFiles) == 0 {
			log.Info("No users found")
		} else {
			for _, file := range savedFiles {
				tableObj.AppendRow(table.Row{"user", file})
			}
			tableObj.Render()
		}
	},
}

var exportUserCmd = &cobra.Command{
	Use:   "export",
	Short: "export users",
	Long:  `export users`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		savedFiles := client.ExportUsers()
		log.Infof("Exporting Users for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"id", "login", "name", "email", "grafanaAdmin", "disabled", "default Password", "authLabels"})
		if len(savedFiles) == 0 {
			log.Info("No users found")
		} else {
			for _, user := range savedFiles {
				var labels string
				if len(user.AuthLabels) > 0 {
					labels = strings.Join(user.AuthLabels, ", ")

				}
				tableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsGrafanaAdmin, user.IsDisabled, api.DefaultUserPassword(user.Login), labels})
			}
			tableObj.Render()
		}
	},
}

var importUserCmd = &cobra.Command{
	Use:   "import",
	Short: "import users",
	Long:  `import users`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		savedFiles := client.ImportUsers()
		log.Infof("Importing Users for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"type", "filename"})
		if len(savedFiles) == 0 {
			log.Info("No users found")
		} else {
			for _, file := range savedFiles {
				tableObj.AppendRow(table.Row{"user", file})
			}
			tableObj.Render()
		}
	},
}

var listUserCmd = &cobra.Command{
	Use:   "list",
	Short: "list users",
	Long:  `list users`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("all credentials for imported users defaults to the sha256 of login.")
		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"id", "login", "name", "email", "admin", "disabled", "default Password", "authLabels"})
		users := client.ListUsers()
		if len(users) == 0 {
			log.Info("No users found")
		} else {
			for _, user := range users {
				var labels string
				if len(user.AuthLabels) > 0 {
					labels = strings.Join(user.AuthLabels, ", ")

				}
				tableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsAdmin, user.IsDisabled, api.DefaultUserPassword(user.Login), labels})
			}
			tableObj.Render()
		}

	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(promoteUser)
	userCmd.AddCommand(deleteUsersCmd)
	userCmd.AddCommand(exportUserCmd)
	userCmd.AddCommand(importUserCmd)
	userCmd.AddCommand(listUserCmd)
	promoteUser.Flags().StringP("user", "u", "", "user email")
	err := promoteUser.MarkFlagRequired("user")
	if err != nil {
		log.Debug("Failed to mark user flag as required")
	}
}
