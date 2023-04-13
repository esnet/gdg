package cmd

import (
	"github.com/esnet/gdg/internal/apphelpers"
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
	Use:   "list",
	Short: "list orgs",
	Long:  `list organizations`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing organizations for context: '%s'", apphelpers.GetContext())
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

func init() {
	rootCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(listOrgCmd)
}
