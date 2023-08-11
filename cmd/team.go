package cmd

import (
	"github.com/esnet/gdg/internal/config"
	api "github.com/esnet/gdg/internal/service"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// teamCmd represents the version command
var teamCmd = &cobra.Command{
	Use:     "teams",
	Aliases: []string{"team"},
	Short:   "Manage teams",
	Long:    `Manage teams.`,
}

func parseTeamGlobalFlags(cmd *cobra.Command) []string {
	teamName, _ := cmd.Flags().GetString("team")
	return []string{teamName}
}

var downloadTeamCmd = &cobra.Command{
	Use:     "download",
	Short:   "download teams from grafana",
	Long:    `download teams from grafana`,
	Aliases: []string{"import", "d"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Importing Teams for context: '%s'", config.Config().AppConfig.GetContext())
		filter := api.NewTeamFilter(parseTeamGlobalFlags(cmd)...)
		savedFiles := grafanaSvc.DownloadTeams(filter)
		if len(savedFiles) == 0 {
			log.Info("No teams found")
		} else {
			tableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "memberCount", "member user ID", "Member Permission"})
			for team, members := range savedFiles {
				tableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.MemberCount})
				for _, member := range members {
					tableObj.AppendRow(table.Row{"", "", "", "", "", member.Login, getTeamPermission(member.Permission)})
				}
			}
			tableObj.Render()
		}
	},
}

var uploadTeamCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload teams to grafana",
	Long:    `upload teams to grafana`,
	Aliases: []string{"export", "u"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Exporting Teams for context: '%s'", config.Config().AppConfig.GetContext())
		log.Warn("Currently support for import Admin members is not support, there will be 1 admin, which is the default admin user")
		filter := api.NewTeamFilter(parseTeamGlobalFlags(cmd)...)
		savedFiles := grafanaSvc.UploadTeams(filter)
		if len(savedFiles) == 0 {
			log.Info("No teams found")
		} else {
			tableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "created", "memberCount", "member Login", "member Permission"})
			for team, members := range savedFiles {
				tableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.MemberCount})
				if team.MemberCount > 0 {
					for _, member := range members {
						tableObj.AppendRow(table.Row{"", "", "", "", "", member.Login, getTeamPermission(member.Permission)})
					}
				}
			}
			tableObj.Render()
		}
	},
}

func getTeamPermission(permissionType models.PermissionType) string {
	permission := "Member"
	if permissionType == models.PermissionType(api.AdminUserPermission) {
		permission = "Admin"
	}
	return permission
}

var listTeamCmd = &cobra.Command{
	Use:     "list",
	Short:   "list teams",
	Long:    `list teams`,
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing teams for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "memberCount", "memberID", "member Permission"})
		filter := api.NewTeamFilter(parseTeamGlobalFlags(cmd)...)
		teams := grafanaSvc.ListTeams(filter)
		if len(teams) == 0 {
			log.Info("No teams found")
		} else {
			for team, members := range teams {
				tableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.MemberCount})
				if team.MemberCount > 0 {
					for _, member := range members {
						tableObj.AppendRow(table.Row{"", "", "", "", "", member.Login, getTeamPermission(member.Permission)})
					}
				}
			}
			tableObj.Render()
		}

	},
}

var deleteTeamCmd = &cobra.Command{
	Use:     "clear",
	Short:   "Delete All Team from grafana",
	Long:    `Delete All Team from grafana`,
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Deleting teams for context: '%s'", config.Config().AppConfig.GetContext())
		filter := api.NewTeamFilter(parseTeamGlobalFlags(cmd)...)
		tableObj.AppendHeader(table.Row{"type", "team ID", "team Name"})
		teams, err := grafanaSvc.DeleteTeam(filter)
		if err != nil {
			log.Error(err.Error())
		} else {
			for _, team := range teams {
				tableObj.AppendRow(table.Row{"team", team.ID, team.Name})
			}
			tableObj.Render()
		}
	},
}

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(downloadTeamCmd)
	teamCmd.AddCommand(uploadTeamCmd)
	teamCmd.AddCommand(listTeamCmd)
	teamCmd.AddCommand(deleteTeamCmd)
	teamCmd.PersistentFlags().StringP("team", "t", "", "team ID")
}
