package api

import (
	"errors"
	"fmt"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/teams"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"

	log "github.com/sirupsen/logrus"
)

type TeamMembersApi interface {
	//TeamMembers
	ListTeamMembers(teamName string) []*models.TeamMemberDTO
	AddTeamMember(teamName string, userLogin string) (string, error)
	DeleteTeamMember(teamName string, userLogin string) (string, error)
}

// List Team Members of specific Team
func (s *DashNGoImpl) ListTeamMembers(teamName string) []*models.TeamMemberDTO {
	team := s.getTeam(teamName)
	if team == nil {
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", teamName))
	}
	p := teams.NewGetTeamMembersParams()
	p.TeamID = fmt.Sprintf("%d", team.ID)
	teamMembers, err := s.client.Teams.GetTeamMembers(p, s.getAuth())
	if err != nil {
		log.Fatal(err)
	}
	return teamMembers.GetPayload()
}

// Add User to a Team
func (s *DashNGoImpl) AddTeamMember(teamName string, userLogin string) (string, error) {
	team := s.getTeam(teamName)
	if team == nil {
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", teamName))
	}
	// Get user from name
	users := s.ListUsers()
	var user *models.UserSearchHitDTO
	for ndx, item := range users {
		if item.Login == userLogin {
			user = users[ndx]
			break
		}
	}
	if user == nil {
		log.Fatal(fmt.Errorf("user:  '%s' could not be found", userLogin))
	}
	p := teams.NewAddTeamMemberParams()
	p.TeamID = fmt.Sprintf("%d", team.ID)
	p.Body = &models.AddTeamMemberCommand{UserID: user.ID}
	msg, err := s.client.Teams.AddTeamMember(p, s.getAuth())
	if err != nil {
		errorMsg := fmt.Sprintf("failed to add member '%s' to team '%s'", userLogin, teamName)
		log.Error(errorMsg)
		return "", errors.New(errorMsg)
	}
	return msg.GetPayload().Message, nil
}

// Delete a specific Team
func (s *DashNGoImpl) DeleteTeamMember(teamName string, userLogin string) (string, error) {
	team := s.getTeam(teamName)
	if team == nil {
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", teamName))
	}
	// Get user from name
	users := s.ListUsers()
	var user *models.UserSearchHitDTO
	for ndx, item := range users {
		if item.Login == userLogin {
			user = users[ndx]
			break
		}
	}
	if user == nil {
		log.Fatal(fmt.Errorf("user:  '%s' could not be found", userLogin))
	}
	p := teams.NewRemoveTeamMemberParams()
	p.TeamID = fmt.Sprintf("%d", team.ID)
	p.UserID = user.ID
	msg, err := s.client.Teams.RemoveTeamMember(p, s.getAuth())
	if err != nil {
		errorMsg := fmt.Sprintf("failed to delete member '%s' from team '%s'", userLogin, teamName)
		log.Error(errorMsg)
		return "", errors.New(errorMsg)
	}
	return msg.GetPayload().Message, nil
}
