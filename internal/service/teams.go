package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"

	"github.com/grafana/grafana-openapi-client-go/client/teams"
	"github.com/grafana/grafana-openapi-client-go/models"
	"golang.org/x/exp/maps"
)

type UserPermission models.PermissionType

const (
	AdminUserPermission = 4
)

func NewTeamFilter(entries ...string) filters.Filter {
	filterObj := filters.NewBaseFilter()

	teamFilter := entries[0]

	filterObj.AddFilter(filters.Name, teamFilter)
	filterObj.AddValidation(filters.Name, func(i interface{}) bool {
		switch val := i.(type) {
		case string:
			if filterObj.GetFilter(filters.Name) == "" {
				return true
			} else if val == filterObj.GetFilter(filters.Name) {
				return true
			}
		default:
			return false
		}

		return false
	})

	return filterObj
}

// DownloadTeams fetches all teams for a given Org
func (s *DashNGoImpl) DownloadTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO {
	teamListing := maps.Keys(s.ListTeams(filter))
	importedTeams := make(map[*models.TeamDTO][]*models.TeamMemberDTO)
	teamPath := BuildResourceFolder("", config.TeamResource)
	for ndx, team := range teamListing {
		// Teams
		teamFileName := filepath.Join(teamPath, GetSlug(team.Name), "team.json")
		teamData, err := json.MarshalIndent(&teamListing[ndx], "", "\t")
		if err != nil {
			slog.Error("could not serialize team object for team name", "teamName", team.Name)
			continue
		}
		// Members
		memberFileName := filepath.Join(teamPath, GetSlug(team.Name), "members.json")
		members, err := s.GetClient().Teams.GetTeamMembers(fmt.Sprintf("%d", team.ID))
		if err != nil {
			slog.Error("could not get team members object for team name", "teamName", team.Name)
			continue
		}
		membersData, err := json.MarshalIndent(members.GetPayload(), "", "\t")
		if err != nil {
			slog.Error("could not serialize team members object for team name", "teamName", team.Name)
			continue
		}
		// Writing Files
		if err = s.storage.WriteFile(teamFileName, teamData); err != nil {
			slog.Error("could not write file", "teamName", team.Name, "err", err)
		} else if err = s.storage.WriteFile(memberFileName, membersData); err != nil {
			slog.Error("could not write team members file", "teamName", team.Name, "err", err)
		} else {
			importedTeams[team] = members.GetPayload()
		}
	}
	return importedTeams
}

// Export Teams
func (s *DashNGoImpl) UploadTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO {
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.TeamResource), true)
	if err != nil {
		slog.Error("failed to list files in directory for teams", "err", err)
	}
	exportedTeams := make(map[*models.TeamDTO][]*models.TeamMemberDTO)
	// Clear previous data.
	_, err = s.DeleteTeam(filter)
	if err != nil {
		log.Fatalf("Failed to clear previous data, aborting")
	}
	for _, fileLocation := range filesInDir {
		if strings.HasSuffix(fileLocation, "team.json") {
			// Export Team
			var rawTeam []byte
			if rawTeam, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
			var newTeam *models.TeamDTO
			if err = json.Unmarshal(rawTeam, &newTeam); err != nil {
				slog.Error("failed to unmarshal file", "filename", fileLocation, "err", err)
				continue
			}
			p := &models.CreateTeamCommand{
				Name:  newTeam.Name,
				Email: newTeam.Email,
			}
			teamCreated, err := s.GetClient().Teams.CreateTeam(p)
			if err != nil {
				slog.Error("failed to create team for file", "filename", fileLocation, "err", err)
			}

			newTeam.ID = teamCreated.GetPayload().TeamID
			// Export Team Members (if exist)
			var currentMembers []*models.TeamMemberDTO
			var rawMembers []byte

			teamMemberLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.TeamResource), GetSlug(newTeam.Name), "members.json")
			if rawMembers, err = s.storage.ReadFile(teamMemberLocation); err != nil {
				slog.Error("failed to find team members", "filename", fileLocation, "err", err)
				continue
			}
			var newMembers []*models.TeamMemberDTO
			if err = json.Unmarshal(rawMembers, &newMembers); err != nil {
				slog.Error("failed to unmarshal file", "filename", fileLocation, "err", err)
				continue
			}
			for _, member := range newMembers {
				if s.isAdminUser(member.UserID, member.Name) {
					slog.Warn("skipping admin user, already added when new team is created")
					continue
				}
				_, err := s.addTeamMember(newTeam, member)
				if err != nil {
					slog.Error("failed to create team member for team", "teamName", newTeam.Name, "MemberID", member.UserID, "err", err)
				} else {
					currentMembers = append(currentMembers, member)
				}
			}
			exportedTeams[newTeam] = currentMembers
		}
	}
	return exportedTeams
}

// List all Teams
func (s *DashNGoImpl) ListTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO {
	result := make(map[*models.TeamDTO][]*models.TeamMemberDTO, 0)
	var pageSize int64 = 99999
	p := teams.NewSearchTeamsParams()
	p.Perpage = &pageSize
	data, err := s.GetClient().Teams.SearchTeams(p)
	if err != nil {
		log.Fatal("unable to list teams")
	}

	getTeamMembers := func(team *models.TeamDTO) {
		if team.MemberCount > 0 {
			result[team] = s.listTeamMembers(filter, team.ID)
		} else {
			result[team] = nil
		}
	}

	for _, team := range data.GetPayload().Teams {
		if filter != nil {
			if filter.InvokeValidation(filters.Name, team.Name) {
				getTeamMembers(team)
			}
		} else {
			getTeamMembers(team)
		}
	}

	return result
}

// Get a specific Team
// Return nil if team cannot be found
//func (s *DashNGoImpl) getTeam(teamName string, filter filters.Filter) *models.TeamDTO {
//	teamListing := maps.Keys(s.ListTeams(filter))
//	var team *models.TeamDTO
//	for ndx, item := range teamListing {
//		if item.Name == teamName {
//			team = teamListing[ndx]
//			break
//		}
//	}
//	return team
//}

// DeleteTeam removes all Teams
func (s *DashNGoImpl) DeleteTeam(filter filters.Filter) ([]*models.TeamDTO, error) {
	teamListing := maps.Keys(s.ListTeams(filter))
	var result []*models.TeamDTO
	for _, team := range teamListing {
		if filter != nil && !filter.ValidateAll(team.Name) {
			continue
		}
		_, err := s.GetClient().Teams.DeleteTeamByID(fmt.Sprintf("%d", team.ID))
		if err != nil {
			slog.Error("failed to delete team", "teamName", team.Name)
			continue
		}
		result = append(result, team)
	}

	return result, nil
}

// List Team Members of specific Team
func (s *DashNGoImpl) listTeamMembers(filter filters.Filter, teamID int64) []*models.TeamMemberDTO {
	teamIDStr := fmt.Sprintf("%d", teamID)
	members, err := s.GetClient().Teams.GetTeamMembers(teamIDStr)
	if err != nil {
		log.Fatal(fmt.Errorf("team:  '%d' could not be found", teamID))
	}

	return members.GetPayload()
}

// Add User to a Team
func (s *DashNGoImpl) addTeamMember(team *models.TeamDTO, userDTO *models.TeamMemberDTO) (string, error) {
	if team == nil {
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", team.Name))
	}
	users := s.ListUsers(NewUserFilter(""))
	var user *models.UserSearchHitDTO
	for ndx, item := range users {
		if item.Login == userDTO.Login {
			user = users[ndx]
			break
		}
	}

	if user == nil {
		log.Fatal(fmt.Errorf("user:  '%s' could not be found", userDTO.Login))
	}
	body := &models.AddTeamMemberCommand{UserID: user.ID}
	msg, err := s.GetClient().Teams.AddTeamMember(fmt.Sprintf("%d", team.ID), body)
	if err != nil {
		slog.Info(err.Error())
		errorMsg := fmt.Sprintf("failed to add member '%s' to team '%s'", userDTO.Login, team.Name)
		slog.Error(errorMsg)
		return "", errors.New(errorMsg)
	}
	if userDTO.Permission == AdminUserPermission {
		adminPatch := teams.NewUpdateTeamMemberParams()
		adminPatch.TeamID = fmt.Sprintf("%d", team.ID)
		adminPatch.UserID = user.ID
		adminPatch.Body = &models.UpdateTeamMemberCommand{Permission: AdminUserPermission}
		response, updateErr := s.GetClient().Teams.UpdateTeamMember(adminPatch)
		if updateErr != nil {
			return "", updateErr
		}
		slog.Debug("Updated permissions for user on team ", "username", userDTO.Name, "teamName", team.Name, "message", response.GetPayload().Message)
	}

	return msg.GetPayload().Message, nil
}
