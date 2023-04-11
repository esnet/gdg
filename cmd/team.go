package cmd

import (
	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// teamCmd represents the version command
var teamCmd = &cobra.Command{
	Use:     "teams",
	Aliases: []string{"user"},
	Short:   "Manage teams",
	Long:    `Manage teams.`,
}

var importTeamCmd = &cobra.Command{
	Use:   "import",
	Short: "import teams",
	Long:  `import teams`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Importing Teams for context: '%s'", apphelpers.GetContext())
		savedFiles := client.ImportTeams()
		if len(savedFiles) == 0 {
			log.Info("No teams found")
		} else {
			tableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "created", "updated"})
			for team := range savedFiles {
				tableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.Created, team.Updated})
			}
			tableObj.Render()
		}
	},
}

var exportTeamCmd = &cobra.Command{
	Use:   "export",
	Short: "export teams",
	Long:  `export teams`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Exporting Teams for context: '%s'", apphelpers.GetContext())
		savedFiles := client.ExportTeams()
		if len(savedFiles) == 0 {
			log.Info("No teams found")
		} else {
			tableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "created", "updated"})
			for team := range savedFiles {
				tableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.Created, team.Updated})
			}
			tableObj.Render()
		}
	},
}

var listTeamCmd = &cobra.Command{
	Use:   "list",
	Short: "list teams",
	Long:  `list teams`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing teams for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "created", "updated"})
		teams := client.ListTeams()
		if len(teams) == 0 {
			log.Info("No teams found")
		} else {
			for _, team := range teams {
				tableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.Created, team.Updated})
			}
			tableObj.Render()
		}

	},
}

var deleteTeamCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Specific Team",
	Long:  `Delete Specific Team`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Deleting team for context: '%s'", apphelpers.GetContext())
		teamName, _ := cmd.Flags().GetString("team")

		msg, err := client.DeleteTeam(teamName)
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(*msg.Message)
		}

	},
}

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(importTeamCmd)
	teamCmd.AddCommand(exportTeamCmd)
	teamCmd.AddCommand(listTeamCmd)
	teamCmd.AddCommand(deleteTeamCmd)
	deleteTeamCmd.Flags().StringP("team", "t", "", "team ID")
	err := deleteTeamCmd.MarkFlagRequired("team")
	if err != nil {
		log.Debug("Failed to mark team flag as required")
	}
}
