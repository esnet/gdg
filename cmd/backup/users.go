package backup

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func newUsersCommand() simplecobra.Commander {
	description := "Manage users"
	return &support.SimpleCommand{
		NameP: "users",
		Short: description,
		Long:  `Provides some utility to manage grafana users from the CLI.  Please note, as the credentials cannot be imported, the export with generate a default password for any user not already present`,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"user", "u"}
			cmd.PersistentFlags().StringP("authlabel", "", "", "filter by a given auth label")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		CommandsList: []simplecobra.Commander{
			newUsersListCmd(),
			newUsersDownloadCmd(),
			newUsersUploadCmd(),
			newUsersClearCmd(),
		},
	}

}

func newUsersListCmd() simplecobra.Commander {
	description := "list users from grafana"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			authLabel, _ := cd.CobraCommand.Flags().GetString("authlabel")
			log.Infof("Listing users for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "login", "name", "email", "admin", "disabled", "default Password", "authLabels"})
			users := rootCmd.GrafanaSvc().ListUsers(service.NewUserFilter(authLabel))
			if len(users) == 0 {
				log.Info("No users found")
			} else {
				for _, user := range users {
					var labels string
					if len(user.AuthLabels) > 0 {
						labels = strings.Join(user.AuthLabels, ", ")

					}
					rootCmd.TableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsAdmin, user.IsDisabled, service.DefaultUserPassword(user.Login), labels})
				}
				rootCmd.TableObj.Render()
			}

			return nil
		},
	}
}
func newUsersDownloadCmd() simplecobra.Commander {
	description := "download users from grafana"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			authLabel, _ := cd.CobraCommand.Flags().GetString("authlabel")
			savedFiles := rootCmd.GrafanaSvc().DownloadUsers(service.NewUserFilter(authLabel))
			log.Infof("Importing Users for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			if len(savedFiles) == 0 {
				log.Info("No users found")
			} else {
				for _, file := range savedFiles {
					rootCmd.TableObj.AppendRow(table.Row{"user", file})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}
func newUsersUploadCmd() simplecobra.Commander {
	description := "upload users to grafana"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			authLabel, _ := cd.CobraCommand.Flags().GetString("authlabel")
			log.Infof("Uploading Users to context: '%s'", config.Config().AppConfig.GetContext())
			savedFiles := rootCmd.GrafanaSvc().UploadUsers(service.NewUserFilter(authLabel))
			rootCmd.TableObj.AppendHeader(table.Row{"id", "login", "name", "email", "grafanaAdmin", "disabled", "default Password", "authLabels"})
			if len(savedFiles) == 0 {
				log.Info("No users found")
			} else {
				for _, user := range savedFiles {
					var labels string
					if len(user.AuthLabels) > 0 {
						labels = strings.Join(user.AuthLabels, ", ")

					}
					rootCmd.TableObj.AppendRow(table.Row{user.ID, user.Login, user.Name, user.Email, user.IsGrafanaAdmin, user.IsDisabled, service.DefaultUserPassword(user.Login), labels})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}
func newUsersClearCmd() simplecobra.Commander {
	description := "delete all users"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			authLabel, _ := cd.CobraCommand.Flags().GetString("authlabel")
			savedFiles := rootCmd.GrafanaSvc().DeleteAllUsers(service.NewUserFilter(authLabel))
			log.Infof("Delete Users for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			if len(savedFiles) == 0 {
				log.Info("No users found")
			} else {
				for _, file := range savedFiles {
					rootCmd.TableObj.AppendRow(table.Row{"user", file})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}
