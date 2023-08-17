package backup

import (
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:     "organizations",
	Aliases: []string{"org", "orgs"},
	Short:   "Manage Organizations",
	Long:    `Manage Grafana Organizations.`,
}

var listOrgCmd = &cobra.Command{
	Use:     "list",
	Short:   "list orgs",
	Long:    `list organizations`,
	Aliases: []string{"l"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Listing organizations for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"id", "org"})
		listOrganizations := cmd.GetGrafanaSvc().ListOrganizations()
		if len(listOrganizations) == 0 {
			log.Info("No organizations found")
		} else {
			for _, org := range listOrganizations {
				cmd.TableObj.AppendRow(table.Row{org.ID, org.Name})
			}
			cmd.TableObj.Render()
		}

	},
}

var downloadOrgCmd = &cobra.Command{
	Use:     "download",
	Short:   "download Organizations",
	Long:    `download organizations`,
	Aliases: []string{"d"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Downloading organizations for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"file"})
		listOrganizations := cmd.GetGrafanaSvc().DownloadOrganizations()
		if len(listOrganizations) == 0 {
			log.Info("No organizations found")
		} else {
			for _, org := range listOrganizations {
				cmd.TableObj.AppendRow(table.Row{org})
			}
			cmd.TableObj.Render()
		}
	},
}

var uploadOrgCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload Orgs to grafana",
	Long:    `upload Orgs to grafana`,
	Aliases: []string{"u"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Uploading Folders for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"file"})
		folders := cmd.GetGrafanaSvc().UploadOrganizations()
		if len(folders) == 0 {
			log.Info("No Orgs were uploaded")
		} else {
			for _, folder := range folders {
				cmd.TableObj.AppendRow(table.Row{folder})
			}
			cmd.TableObj.Render()
		}

	},
}

func init() {
	backupCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(listOrgCmd)
	orgCmd.AddCommand(uploadOrgCmd)
	orgCmd.AddCommand(downloadOrgCmd)
}
