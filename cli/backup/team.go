package backup

import (
	"context"
	"log/slog"
	"os"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	api "github.com/esnet/gdg/internal/service"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func parseTeamGlobalFlags(command *cobra.Command) []string {
	teamName, _ := command.Flags().GetString("team")
	return []string{teamName}
}

func getTeamPermission(permissionType models.PermissionType) string {
	permission := "Member"
	if permissionType == models.PermissionType(api.AdminUserPermission) {
		permission = "Admin"
	}
	return permission
}

func newTeamsCommand() simplecobra.Commander {
	description := "Manage teams"
	return &support.SimpleCommand{
		NameP: "teams",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"team", "t"}
			cmd.PersistentFlags().StringP("team", "t", "", "team ID")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		CommandsList: []simplecobra.Commander{
			newTeamsListCmd(),
			newTeamsDownloadCmd(),
			newTeamsUploadCmd(),
			newTeamsClearCmd(),
		},
	}
}

func getTeamTableWriter() table.Writer {
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)
	writer.SetStyle(table.StyleLight)
	writer.AppendHeader(table.Row{"ID", "Name", "Org ID", "Member Count"}, table.RowConfig{AutoMerge: true})
	return writer
}

func printTeamListing(teams map[*models.TeamDTO][]*models.TeamMemberDTO) {
	for team, members := range teams {
		writer := getTeamTableWriter()
		writer.AppendRow(table.Row{team.ID, team.Name, team.OrgID, team.MemberCount})
		writer.Render()
		var success bool
		twConfigs := table.NewWriter()
		twConfigs.SetOutputMirror(os.Stdout)
		twConfigs.SetStyle(table.StyleDouble)
		twConfigs.AppendHeader(table.Row{"Team Name", "User ID", "LoginID", "Member Permission"})
		for _, member := range members {
			if member == nil {
				continue
			}
			success = true
			twConfigs.AppendRow(table.Row{team.Name, member.UserID, member.Login, getTeamPermission(member.Permission)})
		}
		if success {
			twConfigs.Render()
		}
	}
}

func newTeamsListCmd() simplecobra.Commander {
	description := "list teams from grafana"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Listing teams for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "name", "email", "orgID", "memberCount", "memberID", "member Permission"})
			filter := api.NewTeamFilter(parseTeamGlobalFlags(cd.CobraCommand)...)
			teams := rootCmd.GrafanaSvc().ListTeams(filter)
			if len(teams) == 0 {
				slog.Info("No teams found")
			} else {
				printTeamListing(teams)
			}
			return nil
		},
	}
}

func newTeamsDownloadCmd() simplecobra.Commander {
	description := "download teams from grafana"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Downloading Teams and Member data for context", "context", config.Config().GetGDGConfig().GetContext())
			filter := api.NewTeamFilter(parseTeamGlobalFlags(cd.CobraCommand)...)
			savedFiles := rootCmd.GrafanaSvc().DownloadTeams(filter)
			if len(savedFiles) == 0 {
				slog.Info("No teams found")
			} else {
				printTeamListing(savedFiles)
			}
			return nil
		},
	}
}

func newTeamsUploadCmd() simplecobra.Commander {
	description := "upload teams to grafana"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Exporting Teams for context", "context", config.Config().GetGDGConfig().GetContext())
			filter := api.NewTeamFilter(parseTeamGlobalFlags(cd.CobraCommand)...)
			savedFiles := rootCmd.GrafanaSvc().UploadTeams(filter)
			if len(savedFiles) == 0 {
				slog.Info("No teams found")
			} else {
				printTeamListing(savedFiles)
			}
			return nil
		},
	}
}

func newTeamsClearCmd() simplecobra.Commander {
	description := "Delete All Team from grafana"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Deleting teams for context", "context", config.Config().GetGDGConfig().GetContext())
			filter := api.NewTeamFilter(parseTeamGlobalFlags(cd.CobraCommand)...)
			rootCmd.TableObj.AppendHeader(table.Row{"type", "team ID", "team Name"})
			teams, err := rootCmd.GrafanaSvc().DeleteTeam(filter)
			if err != nil {
				slog.Error(err.Error())
			} else {
				for _, team := range teams {
					rootCmd.TableObj.AppendRow(table.Row{"team", team.ID, team.Name})
				}
				rootCmd.Render(cd.CobraCommand, teams)
			}
			return nil
		},
	}
}
