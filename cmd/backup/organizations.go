package backup

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
)

func newOrganizationsCommand() simplecobra.Commander {
	description := "Manage Grafana Organizations."
	return &support.SimpleCommand{
		NameP: "organizations",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"org", "orgs"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		CommandsList: []simplecobra.Commander{
			newOrganizationsListCmd(),
			newOrganizationsDownloadCmd(),
			newOrganizationsUploadCmd(),
		},
	}

}

func newOrganizationsListCmd() simplecobra.Commander {
	description := "List Grafana Organizations."
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Listing organizations for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "org"})
			listOrganizations := rootCmd.GrafanaSvc().ListOrganizations()
			sort.Slice(listOrganizations, func(a, b int) bool {
				return listOrganizations[a].ID < listOrganizations[b].ID
			})
			if len(listOrganizations) == 0 {
				log.Info("No organizations found")
			} else {
				for _, org := range listOrganizations {
					rootCmd.TableObj.AppendRow(table.Row{org.ID, org.Name})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}

}
func newOrganizationsDownloadCmd() simplecobra.Commander {
	description := "download Organizations"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Downloading organizations for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			listOrganizations := rootCmd.GrafanaSvc().DownloadOrganizations()
			if len(listOrganizations) == 0 {
				log.Info("No organizations found")
			} else {
				for _, org := range listOrganizations {
					rootCmd.TableObj.AppendRow(table.Row{org})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}

}
func newOrganizationsUploadCmd() simplecobra.Commander {
	description := "upload Orgs to grafana"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Uploading Folders for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			folders := rootCmd.GrafanaSvc().UploadOrganizations()
			if len(folders) == 0 {
				log.Info("No Orgs were uploaded")
			} else {
				for _, folder := range folders {
					rootCmd.TableObj.AppendRow(table.Row{folder})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}

}
