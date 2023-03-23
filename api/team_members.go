package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
)

// List Team Members of specific Team
func (s *DashNGoImpl) ListTeamMembers(teamName string) []sdk.TeamMember {
	ctx := context.Background()
	team := s.getTeam(teamName)
	if team == nil {
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", teamName))
	}
	teamMembers, err := s.GetAdminClient().GetTeamMembers(ctx, team.ID)
	if err != nil {
		log.Fatal(err)
	}
	return teamMembers
}

// Add User to a Team
func (s *DashNGoImpl) AddTeamMember(teamName string, userLogin string) (*sdk.StatusMessage, error) {
	ctx := context.Background()
	// Get team from name
	team := s.getTeam(teamName)
	if team == nil {
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", teamName))
	}
	// Get user from name
	users := s.ListUsers()
	var user *sdk.User
	for ndx, item := range users {
		if item.Login == userLogin {
			user = &users[ndx]
			break
		}
	}
	if user == nil {
		log.Fatal(fmt.Errorf("user:  '%s' could not be found", userLogin))
	}
	msg, err := s.GetAdminClient().AddTeamMember(ctx, team.ID, user.ID)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to add member '%s' to team '%s'", userLogin, teamName)
		log.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}
	return &msg, nil
}

// Delete a specific Team
func (s *DashNGoImpl) DeleteTeamMember(teamName string, userLogin string) (*sdk.StatusMessage, error) {
	ctx := context.Background()
	// Get team from name
	team := s.getTeam(teamName)
	if team == nil {
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", teamName))
	}
	// Get user from name
	users := s.ListUsers()
	var user *sdk.User
	for ndx, item := range users {
		if item.Login == userLogin {
			user = &users[ndx]
			break
		}
	}
	if user == nil {
		log.Fatal(fmt.Errorf("user:  '%s' could not be found", userLogin))
	}
	msg, err := s.GetAdminClient().DeleteTeamMember(ctx, team.ID, user.ID)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to delete member '%s' from team '%s'", userLogin, teamName)
		log.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}
	return &msg, nil
}
