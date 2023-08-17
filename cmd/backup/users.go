package backup

import (
	"github.com/esnet/gdg/cmd"
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

var deleteUsersCmd = &cobra.Command{
	Use:     "clear",
	Short:   "delete all users",
	Long:    `delete all users`,
	Aliases: []string{"c"},
	Run: func(command *cobra.Command, args []string) {
		authLabel, _ := command.Flags().GetString("authlabel")
		savedFiles := cmd.GetGrafanaSvc().DeleteAllUsers(service.NewUserFilter(authLabel))
		log.Infof("Delete Users for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		if len(savedFiles) == 0 {
			log.Info("No users found")
		} else {
			for _, file := range savedFiles {
				cmd.TableObj.AppendRow(table.Row{"user", file})
			}
			cmd.TableObj.Render()
		}
	},
}

var uploadUsersCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload users to grafana",
	Long:    `upload users to grafana`,
	Aliases: []string{"export", "u"},
	Run: func(command *cobra.Command, args []string) {
		log.Infof("Uploading Users to context: '%s'", config.Config().AppConfig.GetContext())
		authLabel, _ := command.Flags().GetString("authlabel")
		savedFiles := cmd.GetGrafanaSvc().UploadUsers(service.NewUserFilter(authLabel))
		cmd.TableObj.AppendHeader(table.Row{"id", "login", "name", "email", "grafanaAdmin", "disabled", "default Password", "authLabels"})
		if len(savedFiles) == 0 {
			log.Info("No users found")
		} else {
			for _, user := range savedFiles {
				var labels string
				if len(user.AuthLabels) > 0 {
					labels = strings.Join(user.AuthLabels, ", ")

				}
				cmd.TableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsGrafanaAdmin, user.IsDisabled, service.DefaultUserPassword(user.Login), labels})
			}
			cmd.TableObj.Render()
		}
	},
}

var downloadUsersCmd = &cobra.Command{
	Use:     "download",
	Short:   "download users from grafana",
	Long:    `download users from grafana`,
	Aliases: []string{"d"},
	Run: func(command *cobra.Command, args []string) {
		authLabel, _ := command.Flags().GetString("authlabel")
		savedFiles := cmd.GetGrafanaSvc().DownloadUsers(service.NewUserFilter(authLabel))
		log.Infof("Importing Users for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		if len(savedFiles) == 0 {
			log.Info("No users found")
		} else {
			for _, file := range savedFiles {
				cmd.TableObj.AppendRow(table.Row{"user", file})
			}
			cmd.TableObj.Render()
		}
	},
}

var listUserCmd = &cobra.Command{
	Use:   "list",
	Short: "list users",
	Long:  `list users`,
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Listing users for context: '%s'", config.Config().AppConfig.GetContext())
		authLabel, _ := command.Flags().GetString("authlabel")
		cmd.TableObj.AppendHeader(table.Row{"id", "login", "name", "email", "admin", "disabled", "default Password", "authLabels"})
		users := cmd.GetGrafanaSvc().ListUsers(service.NewUserFilter(authLabel))
		if len(users) == 0 {
			log.Info("No users found")
		} else {
			for _, user := range users {
				var labels string
				if len(user.AuthLabels) > 0 {
					labels = strings.Join(user.AuthLabels, ", ")

				}
				cmd.TableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsAdmin, user.IsDisabled, service.DefaultUserPassword(user.Login), labels})
			}
			cmd.TableObj.Render()
		}

	},
}

func init() {
	backupCmd.AddCommand(userCmd)
	userCmd.AddCommand(deleteUsersCmd)
	userCmd.AddCommand(uploadUsersCmd)
	userCmd.AddCommand(downloadUsersCmd)
	userCmd.AddCommand(listUserCmd)
	userCmd.PersistentFlags().StringP("authlabel", "", "", "filter by a given auth label")

}
