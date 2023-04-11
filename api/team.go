package api

import (
	"errors"
	"fmt"
	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/teams"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"os"
	"strings"

	"encoding/json"
	"path/filepath"

	"github.com/esnet/gdg/config"
	log "github.com/sirupsen/logrus"
)

type TeamsApi interface {
	//Team
	ImportTeams() map[*models.TeamDTO][]*models.TeamMemberDTO
	ExportTeams() map[*models.TeamDTO][]*models.TeamMemberDTO
	ListTeams() []*models.TeamDTO
	DeleteTeam(teamName string) (string, error)
	TeamMembersApi
}

// Import Teams
func (s *DashNGoImpl) ImportTeams() map[*models.TeamDTO][]*models.TeamMemberDTO {
	teamListing := s.ListTeams()
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
func (s *DashNGoImpl) ExportTeams() map[*models.TeamDTO][]*models.TeamMemberDTO {
	filesInDir, err := s.storage.FindAllFiles(apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.TeamResource), true)
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for teams")
	}
	exportedTeams := make(map[*models.TeamDTO][]*models.TeamMemberDTO)
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
			_, err = s.client.Teams.CreateTeam(p, s.getAuth())
			if err != nil {
				log.WithError(err).Errorf("failed to create team for file: %s", fileLocation)
				continue
			}
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
					log.Info("skipping admin user, already added when new team is created")
					continue
				}
				p := teams.NewAddTeamMemberParams()
				p.TeamID = fmt.Sprintf("%d", member.TeamID)
				p.Body = &models.AddTeamMemberCommand{
					UserID: member.UserID,
				}
				_, err = s.client.Teams.AddTeamMember(p, s.getAdminAuth())
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
func (s *DashNGoImpl) ListTeams() []*models.TeamDTO {
	var pageSize int64 = 99999
	p := teams.NewSearchTeamsParams()
	p.Perpage = &pageSize
	data, err := s.client.Teams.SearchTeams(p, s.getAuth())
	if err != nil {
		log.Fatal("unable to list teams")
	}

	return data.GetPayload().Teams
}

// Get a specific Team
// Return nil if team cannot be found
func (s *DashNGoImpl) getTeam(teamName string) *models.TeamDTO {
	teamListing := s.ListTeams()
	var team *models.TeamDTO
	for ndx, item := range teamListing {
		if item.Name == teamName {
			team = teamListing[ndx]
			break
		}
	}
	return team
}

// Delete a specific Team
func (s *DashNGoImpl) DeleteTeam(teamName string) (string, error) {
	team := s.getTeam(teamName)
	if team == nil {
		return "", fmt.Errorf("team:  '%s' could not be found", teamName)
	}
	p := teams.NewDeleteTeamByIDParams()
	p.TeamID = fmt.Sprintf("%d", team.ID)
	msg, err := s.client.Teams.DeleteTeamByID(p, s.getAuth())
	if err != nil {
		errorMsg := fmt.Sprintf("failed to delete team: '%s'", teamName)
		log.Error(errorMsg)
		return "", errors.New(errorMsg)
	}
	return msg.GetPayload().Message, nil
}
