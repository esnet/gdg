package service

import (
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/apphelpers"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"

	"github.com/esnet/grafana-swagger-api-golang/goclient/client/teams"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"golang.org/x/exp/maps"
	"os"
	"strings"

	"encoding/json"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type TeamsApi interface {
	//Team
	ImportTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	ExportTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	ListTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	DeleteTeam(filter filters.Filter) ([]*models.TeamDTO, error)
}

type UserPermission models.PermissionType

const (
	AdminUserPermission = 4
)

func NewTeamFilter(entries ...string) filters.Filter {
	filterObj := filters.NewBaseFilter()

	teamFilter := entries[0]

	filterObj.AddFilter(filters.Name, teamFilter)
	filterObj.AddValidation(filters.Name, func(i interface{}) bool {
		switch i.(type) {
		case string:
			{
				val := i.(string)
				if filterObj.GetFilter(filters.Name) == "" {
					return true
				} else if val == filterObj.GetFilter(filters.Name) {
					return true
				}
			}

		default:
			return false
		}

		return false
	})

	return filterObj
}

// Import Teams
func (s *DashNGoImpl) ImportTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO {
	teamListing := maps.Keys(s.ListTeams(filter))
	importedTeams := make(map[*models.TeamDTO][]*models.TeamMemberDTO)
	teamPath := buildResourceFolder("", config.TeamResource)
	for ndx, team := range teamListing {
		//Teams
		teamFileName := filepath.Join(teamPath, GetSlug(team.Name), "team.json")
		teamData, err := json.MarshalIndent(&teamListing[ndx], "", "\t")
		if err != nil {
			log.Errorf("could not serialize team object for team name: %s", team.Name)
			continue
		}
		//Members
		memberFileName := filepath.Join(teamPath, GetSlug(team.Name), "members.json")
		p := teams.NewGetTeamMembersParams()
		p.TeamID = fmt.Sprintf("%d", team.ID)
		members, err := s.client.Teams.GetTeamMembers(p, s.getAuth())
		if err != nil {
			log.Errorf("could not get team members object for team name: %s", team.Name)
			continue
		}
		membersData, err := json.MarshalIndent(members.GetPayload(), "", "\t")
		if err != nil {
			log.Errorf("could not serialize team members object for team name: %s", team.Name)
			continue
		}
		//Writing Files
		if err = s.storage.WriteFile(teamFileName, teamData, os.FileMode(int(0666))); err != nil {
			log.WithError(err).Errorf("for %s\n", team.Name)
		} else if err = s.storage.WriteFile(memberFileName, membersData, os.FileMode(int(0666))); err != nil {
			log.WithError(err).Errorf("for %s\n", team.Name)
		} else {
			importedTeams[team] = members.GetPayload()
		}
	}
	return importedTeams
}

// Export Teams
func (s *DashNGoImpl) ExportTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO {
	filesInDir, err := s.storage.FindAllFiles(apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.TeamResource), true)
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for teams")
	}
	exportedTeams := make(map[*models.TeamDTO][]*models.TeamMemberDTO)
	//Clear previous data.
	_, err = s.DeleteTeam(filter)
	if err != nil {
		log.Fatalf("Failed to clear previous data, aborting")
	}
	for _, fileLocation := range filesInDir {
		if strings.HasSuffix(fileLocation, "team.json") {
			//Export Team
			var rawTeam []byte
			if rawTeam, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file: %s", fileLocation)
				continue
			}
			var newTeam *models.TeamDTO
			if err = json.Unmarshal(rawTeam, &newTeam); err != nil {
				log.WithError(err).Errorf("failed to unmarshal file: %s", fileLocation)
				continue
			}
			p := teams.NewCreateTeamParams()
			p.Body = &models.CreateTeamCommand{
				Name:  newTeam.Name,
				Email: newTeam.Email,
			}
			teamCreated, err := s.client.Teams.CreateTeam(p, s.getAuth())
			if err != nil {
				log.WithError(err).Errorf("failed to create team for file: %s", fileLocation)
				continue
			}

			newTeam.ID = teamCreated.GetPayload().TeamID
			//Export Team Members (if exist)
			var currentMembers []*models.TeamMemberDTO
			var rawMembers []byte

			teamMemberLocation := filepath.Join(apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.TeamResource), GetSlug(newTeam.Name), "members.json")
			if rawMembers, err = s.storage.ReadFile(teamMemberLocation); err != nil {
				log.WithError(err).Errorf("failed to find team members: %s", fileLocation)
				continue
			}
			var newMembers []*models.TeamMemberDTO
			if err = json.Unmarshal(rawMembers, &newMembers); err != nil {
				log.WithError(err).Errorf("failed to unmarshal file: %s", fileLocation)
				continue
			}
			for _, member := range newMembers {
				if s.isAdmin(member.UserID, member.Name) {
					log.Warnf("skipping admin user, already added when new team is created")
					continue
				}
				_, err := s.addTeamMember(newTeam, member)
				if err != nil {
					log.WithError(err).Errorf("failed to create team member for team %s with ID %d", newTeam.Name, member.UserID)
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
	data, err := s.client.Teams.SearchTeams(p, s.getAuth())
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
func (s *DashNGoImpl) getTeam(teamName string, filter filters.Filter) *models.TeamDTO {
	teamListing := maps.Keys(s.ListTeams(filter))
	var team *models.TeamDTO
	for ndx, item := range teamListing {
		if item.Name == teamName {
			team = teamListing[ndx]
			break
		}
	}
	return team
}

// DeleteTeam removes all Teams
func (s *DashNGoImpl) DeleteTeam(filter filters.Filter) ([]*models.TeamDTO, error) {
	teamListing := maps.Keys(s.ListTeams(filter))
	var result []*models.TeamDTO
	for _, team := range teamListing {
		if filter != nil && !filter.ValidateAll(team.Name) {
			continue
		}
		p := teams.NewDeleteTeamByIDParams()
		p.TeamID = fmt.Sprintf("%d", team.ID)
		_, err := s.client.Teams.DeleteTeamByID(p, s.getAuth())
		if err != nil {
			log.Errorf("failed to delete team: '%s'", team.Name)
			continue
		}
		result = append(result, team)
	}

	return result, nil
}

// List Team Members of specific Team
func (s *DashNGoImpl) listTeamMembers(filter filters.Filter, teamID int64) []*models.TeamMemberDTO {
	teamIDStr := fmt.Sprintf("%d", teamID)
	fetchTeamParam := teams.NewGetTeamMembersParams()
	fetchTeamParam.TeamID = teamIDStr
	members, err := s.client.Teams.GetTeamMembers(fetchTeamParam, s.getAuth())
	if err != nil {
		log.Fatal(fmt.Errorf("team:  '%d' could not be found", teamID))
	}

	return members.GetPayload()
}

// Add User to a Team
// TODO: add support to import member with correct permission granted.
func (s *DashNGoImpl) addTeamMember(team *models.TeamDTO, userDTO *models.TeamMemberDTO) (string, error) {
	if team == nil {
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", team.Name))
	}
	users := s.ListUsers()
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
	p := teams.NewAddTeamMemberParams()
	p.TeamID = fmt.Sprintf("%d", team.ID)
	p.Body = &models.AddTeamMemberCommand{UserID: user.ID}
	msg, err := s.client.Teams.AddTeamMember(p, s.getAuth())
	if err != nil {
		log.Info(err.Error())
		errorMsg := fmt.Sprintf("failed to add member '%s' to team '%s'", userDTO.Login, team.Name)
		log.Error(errorMsg)
		return "", errors.New(errorMsg)
	}
	return msg.GetPayload().Message, nil
}
