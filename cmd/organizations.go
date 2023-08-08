package cmd

import (
	"errors"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
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
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing organizations for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"id", "org"})
		listOrganizations := grafanaSvc.ListOrganizations()
		if len(listOrganizations) == 0 {
			log.Info("No organizations found")
		} else {
			for _, org := range listOrganizations {
				tableObj.AppendRow(table.Row{org.ID, org.Name})
			}
			tableObj.Render()
		}

	},
}

var downloadOrgCmd = &cobra.Command{
	Use:     "download",
	Short:   "download Organizations",
	Long:    `download organizations`,
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Downloading organizations for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"file"})
		listOrganizations := grafanaSvc.DownloadOrganizations()
		if len(listOrganizations) == 0 {
			log.Info("No organizations found")
		} else {
			for _, org := range listOrganizations {
				tableObj.AppendRow(table.Row{org})
			}
			tableObj.Render()
		}
	},
}

var setOrgCmd = &cobra.Command{
	Use:   "set",
	Short: "set <OrgId>, 0 removes filter",
	Long:  `set <OrgId>, 0 removes filter`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires an Org ID")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		ctx := args[0]
		orgId, err := strconv.ParseInt(ctx, 10, 64)
		if err != nil {
			log.Fatal("invalid Org ID, could not parse value to a numeric value")
		}
		err = grafanaSvc.SetOrganization(orgId)
		if err != nil {
			log.WithError(err).Fatal("unable to set Org ID")
		}
		log.Infof("Succesfully set Org ID for context: %s", config.Config().AppConfig.GetContext())
	},
}

var uploadOrgCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload Orgs to grafana",
	Long:    `upload Orgs to grafana`,
	Aliases: []string{"u"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Uploading Folders for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"file"})
		folders := grafanaSvc.UploadOrganizations()
		if len(folders) == 0 {
			log.Info("No Orgs were uploaded")
		} else {
			for _, folder := range folders {
				tableObj.AppendRow(table.Row{folder})
			}
			tableObj.Render()
		}

	},
}

func init() {
	rootCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(listOrgCmd)
	orgCmd.AddCommand(uploadOrgCmd)
	orgCmd.AddCommand(setOrgCmd)
	orgCmd.AddCommand(downloadOrgCmd)
}
