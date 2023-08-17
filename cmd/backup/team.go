package backup

import (
	"github.com/esnet/gdg/cmd"
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

func parseTeamGlobalFlags(command *cobra.Command) []string {
	teamName, _ := command.Flags().GetString("team")
	return []string{teamName}
}

var downloadTeamCmd = &cobra.Command{
	Use:     "download",
	Short:   "download teams from grafana",
	Long:    `download teams from grafana`,
	Aliases: []string{"d"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Importing Teams for context: '%s'", config.Config().AppConfig.GetContext())
		filter := api.NewTeamFilter(parseTeamGlobalFlags(command)...)
		savedFiles := cmd.GetGrafanaSvc().DownloadTeams(filter)
		if len(savedFiles) == 0 {
			log.Info("No teams found")
		} else {
			cmd.TableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "memberCount", "member user ID", "Member Permission"})
			for team, members := range savedFiles {
				cmd.TableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.MemberCount})
				for _, member := range members {
					cmd.TableObj.AppendRow(table.Row{"", "", "", "", "", member.Login, getTeamPermission(member.Permission)})
				}
			}
			cmd.TableObj.Render()
		}
	},
}

var uploadTeamCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload teams to grafana",
	Long:    `upload teams to grafana`,
	Aliases: []string{"u"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Exporting Teams for context: '%s'", config.Config().AppConfig.GetContext())
		log.Warn("Currently support for import Admin members is not support, there will be 1 admin, which is the default admin user")
		filter := api.NewTeamFilter(parseTeamGlobalFlags(command)...)
		savedFiles := cmd.GetGrafanaSvc().UploadTeams(filter)
		if len(savedFiles) == 0 {
			log.Info("No teams found")
		} else {
			cmd.TableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "created", "memberCount", "member Login", "member Permission"})
			for team, members := range savedFiles {
				cmd.TableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.MemberCount})
				if team.MemberCount > 0 {
					for _, member := range members {
						cmd.TableObj.AppendRow(table.Row{"", "", "", "", "", member.Login, getTeamPermission(member.Permission)})
					}
				}
			}
			cmd.TableObj.Render()
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
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Listing teams for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "memberCount", "memberID", "member Permission"})
		filter := api.NewTeamFilter(parseTeamGlobalFlags(command)...)
		teams := cmd.GetGrafanaSvc().ListTeams(filter)
		if len(teams) == 0 {
			log.Info("No teams found")
		} else {
			for team, members := range teams {
				cmd.TableObj.AppendRow(table.Row{team.ID, team.Name, team.Email, team.OrgID, team.MemberCount})
				if team.MemberCount > 0 {
					for _, member := range members {
						cmd.TableObj.AppendRow(table.Row{"", "", "", "", "", member.Login, getTeamPermission(member.Permission)})
					}
				}
			}
			cmd.TableObj.Render()
		}

	},
}

var deleteTeamCmd = &cobra.Command{
	Use:     "clear",
	Short:   "Delete All Team from grafana",
	Long:    `Delete All Team from grafana`,
	Aliases: []string{"c"},
	Run: func(command *cobra.Command, args []string) {
		log.Infof("Deleting teams for context: '%s'", config.Config().AppConfig.GetContext())
		filter := api.NewTeamFilter(parseTeamGlobalFlags(command)...)
		cmd.TableObj.AppendHeader(table.Row{"type", "team ID", "team Name"})
		teams, err := cmd.GetGrafanaSvc().DeleteTeam(filter)
		if err != nil {
			log.Error(err.Error())
		} else {
			for _, team := range teams {
				cmd.TableObj.AppendRow(table.Row{"team", team.ID, team.Name})
			}
			cmd.TableObj.Render()
		}
	},
}

func init() {
	backupCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(downloadTeamCmd)
	teamCmd.AddCommand(uploadTeamCmd)
	teamCmd.AddCommand(listTeamCmd)
	teamCmd.AddCommand(deleteTeamCmd)
	teamCmd.PersistentFlags().StringP("team", "t", "", "team ID")
}
