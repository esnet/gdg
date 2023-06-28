package cmd

import (
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
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

		log.Infof("Listing dashboards for context: '%s'", config.Config().AppConfig.GetContext())
		userLogin, _ := cmd.Flags().GetString("user")

		msg, err := grafanaSvc.PromoteUser(userLogin)
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(msg)
			log.Info("Please note user is a grafana admin, not necessarily an Org admin.  You may need to promote yourself manually per org")
		}

	},
}
var deleteUsersCmd = &cobra.Command{
	Use:     "clear",
	Short:   "delete all users",
	Long:    `delete all users`,
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		authLabel, _ := cmd.Flags().GetString("authlabel")
		savedFiles := grafanaSvc.DeleteAllUsers(service.NewUserFilter(authLabel))
		log.Infof("Delete Users for context: '%s'", config.Config().AppConfig.GetContext())
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

var uploadUsersCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload users to grafana",
	Long:    `upload users to grafana`,
	Aliases: []string{"export", "u"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Uploading Users to context: '%s'", config.Config().AppConfig.GetContext())
		authLabel, _ := cmd.Flags().GetString("authlabel")
		savedFiles := grafanaSvc.ExportUsers(service.NewUserFilter(authLabel))
		tableObj.AppendHeader(table.Row{"id", "login", "name", "email", "grafanaAdmin", "disabled", "default Password", "authLabels"})
		if len(savedFiles) == 0 {
			log.Info("No users found")
		} else {
			for _, user := range savedFiles {
				var labels string
				if len(user.AuthLabels) > 0 {
					labels = strings.Join(user.AuthLabels, ", ")

				}
				tableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsGrafanaAdmin, user.IsDisabled, service.DefaultUserPassword(user.Login), labels})
			}
			tableObj.Render()
		}
	},
}

var downloadUsersCmd = &cobra.Command{
	Use:     "download",
	Short:   "download users from grafana",
	Long:    `download users from grafana`,
	Aliases: []string{"import", "d"},
	Run: func(cmd *cobra.Command, args []string) {
		authLabel, _ := cmd.Flags().GetString("authlabel")
		savedFiles := grafanaSvc.ImportUsers(service.NewUserFilter(authLabel))
		log.Infof("Importing Users for context: '%s'", config.Config().AppConfig.GetContext())
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

		log.Infof("Listing users for context: '%s'", config.Config().AppConfig.GetContext())
		authLabel, _ := cmd.Flags().GetString("authlabel")
		tableObj.AppendHeader(table.Row{"id", "login", "name", "email", "admin", "disabled", "default Password", "authLabels"})
		users := grafanaSvc.ListUsers(service.NewUserFilter(authLabel))
		if len(users) == 0 {
			log.Info("No users found")
		} else {
			for _, user := range users {
				var labels string
				if len(user.AuthLabels) > 0 {
					labels = strings.Join(user.AuthLabels, ", ")

				}
				tableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsAdmin, user.IsDisabled, service.DefaultUserPassword(user.Login), labels})
			}
			tableObj.Render()
		}

	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(promoteUser)
	userCmd.AddCommand(deleteUsersCmd)
	userCmd.AddCommand(uploadUsersCmd)
	userCmd.AddCommand(downloadUsersCmd)
	userCmd.AddCommand(listUserCmd)
	promoteUser.Flags().StringP("user", "u", "", "user email")
	userCmd.PersistentFlags().StringP("authlabel", "", "", "filter by a given auth label")
	err := promoteUser.MarkFlagRequired("user")
	if err != nil {
		log.Debug("Failed to mark user flag as required")
	}
}
