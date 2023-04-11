package cmd

import (
	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// teamMemberCmd represents the version command
var teamMemberCmd = &cobra.Command{
	Use:     "teammembers",
	Aliases: []string{"teammembers"},
	Short:   "Manage team members",
	Long:    `Manage team members.`,
}

var listTeamMemberCmd = &cobra.Command{
	Use:   "list",
	Short: "list team members for team",
	Long:  `list team members for team`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		teamName, _ := cmd.Flags().GetString("team")
		tableObj.AppendHeader(table.Row{"orgID", "teamID", "userID", "email", "login", "avatarURL"})
		teamMembers := client.ListTeamMembers(teamName)
		if len(teamMembers) == 0 {
			log.Info("No team members found")
		} else {
			for _, member := range teamMembers {
				tableObj.AppendRow(table.Row{member.OrgId, member.TeamId, member.UserId, member.Email, member.Login, member.AvatarUrl})
			}
			tableObj.Render()
		}
	},
}

var deleteTeamMemberCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete team member",
	Long:  `delete team member`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		teamName, _ := cmd.Flags().GetString("team")
		userLogin, _ := cmd.Flags().GetString("user")
		msg, err := client.DeleteTeamMember(teamName, userLogin)
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(*msg.Message)
		}
	},
}

var addTeamMemberCmd = &cobra.Command{
	Use:   "add",
	Short: "add team member",
	Long:  `add team member`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing dashboards for context: '%s'", apphelpers.GetContext())
		teamName, _ := cmd.Flags().GetString("team")
		userLogin, _ := cmd.Flags().GetString("user")
		msg, err := client.AddTeamMember(teamName, userLogin)
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(*msg.Message)
		}
	},
}

func init() {
	rootCmd.AddCommand(teamMemberCmd)
	teamMemberCmd.AddCommand(listTeamMemberCmd)
	teamMemberCmd.AddCommand(deleteTeamMemberCmd)
	teamMemberCmd.AddCommand(addTeamMemberCmd)
	listTeamMemberCmd.Flags().StringP("team", "t", "", "team ID")
	deleteTeamMemberCmd.Flags().StringP("team", "t", "", "team ID")
	deleteTeamMemberCmd.Flags().StringP("user", "u", "", "user login")
	addTeamMemberCmd.Flags().StringP("team", "t", "", "team ID")
	addTeamMemberCmd.Flags().StringP("user", "u", "", "user login")
	err := listTeamMemberCmd.MarkFlagRequired("team")
	if err != nil {
		log.Debug("Failed to mark team flag as required")
	}
	err = deleteTeamMemberCmd.MarkFlagRequired("team")
	if err != nil {
		log.Debug("Failed to mark team flag as required")
	}
	err = addTeamMemberCmd.MarkFlagRequired("team")
	if err != nil {
		log.Debug("Failed to mark team flag as required")
	}
	err = deleteTeamMemberCmd.MarkFlagRequired("user")
	if err != nil {
		log.Debug("Failed to mark user flag as required")
	}
	err = addTeamMemberCmd.MarkFlagRequired("user")
	if err != nil {
		log.Debug("Failed to mark user flag as required")
	}
}
