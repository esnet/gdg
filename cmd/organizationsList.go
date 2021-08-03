package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listOrgCmd = &cobra.Command{
	Use:   "list",
	Short: "list orgs",
	Long:  `list organizations`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing organizations for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"id", "org", "address1", "address2", "city"})
		orgs := client.ListOrganizations()
		if len(orgs) == 0 {
			log.Info("No organizations found")
		} else {
			for _, org := range orgs {
				tableObj.AppendRow(table.Row{org.ID, org.Name, org.Address.Address1, org.Address.Address2, org.Address.City})
			}
			tableObj.Render()
		}

	},
}

func init() {
	orgCmd.AddCommand(listOrgCmd)
}
