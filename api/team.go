package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"encoding/json"
	"path/filepath"

	"github.com/esnet/gdg/config"
	"github.com/tidwall/pretty"

	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
)

// Import Teams
func (s *DashNGoImpl) ImportTeams() map[sdk.Team][]sdk.TeamMember {
	var (
		teamData []byte
	)
	ctx := context.Background()
	searchParams := sdk.WithPagesize(99999)
	pageTeams, err := s.GetAdminClient().SearchTeams(ctx, searchParams)
	if err != nil {
		log.Fatal(err)
	}
	teams := pageTeams.Teams
	importedTeams := make(map[sdk.Team][]sdk.TeamMember)
	teamPath := buildResourceFolder("", config.TeamResource)
	for ndx, team := range teams {
		//Teams
		teamFileName := filepath.Join(teamPath, GetSlug(team.Name), "team.json")
		teamData, err = json.Marshal(&teams[ndx])
		if err != nil {
			log.Errorf("could not serialize team object for team name: %s", team.Name)
			continue
		}
		//Members
		memberFileName := filepath.Join(teamPath, GetSlug(team.Name), "members.json")
		members, err := s.GetAdminClient().GetTeamMembers(ctx, team.ID)
		if err != nil {
			log.Errorf("could not get team members object for team name: %s", team.Name)
			continue
		}
		membersData, err := json.Marshal(members)
		if err != nil {
			log.Errorf("could not serialize team members object for team name: %s", team.Name)
			continue
		}
		//Writing Files
		if err = s.storage.WriteFile(teamFileName, pretty.Pretty(teamData), os.FileMode(int(0666))); err != nil {
			log.WithError(err).Errorf("for %s\n", team.Name)
		} else if err = s.storage.WriteFile(memberFileName, pretty.Pretty(membersData), os.FileMode(int(0666))); err != nil {
			log.WithError(err).Errorf("for %s\n", team.Name)
		} else {
			importedTeams[team] = members
		}
	}
	return importedTeams
}

// Export Teams
func (s *DashNGoImpl) ExportTeams() map[sdk.Team][]sdk.TeamMember {
	ctx := context.Background()
	filesInDir, err := s.storage.FindAllFiles(getResourcePath(config.TeamResource), true)
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for teams")
	}
	exportedTeams := make(map[sdk.Team][]sdk.TeamMember)
	for _, fileLocation := range filesInDir {
		if strings.HasSuffix(fileLocation, "team.json") {
			//Export Team
			var rawTeam []byte
			if rawTeam, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file: %s", fileLocation)
				continue
			}
			var newTeam sdk.Team
			if err = json.Unmarshal(rawTeam, &newTeam); err != nil {
				log.WithError(err).Errorf("failed to unmarshal file: %s", fileLocation)
				continue
			}
			_, err = s.GetAdminClient().CreateTeam(ctx, newTeam)
			if err != nil {
				log.WithError(err).Errorf("failed to create team for file: %s", fileLocation)
				continue
			}
			//Export Team Members (if exist)
			var currentMembers []sdk.TeamMember
			var rawMembers []byte
			teamMemberLocation := filepath.Join(getResourcePath(config.TeamResource), GetSlug(newTeam.Name), "members.json")
			if rawMembers, err = s.storage.ReadFile(teamMemberLocation); err != nil {
				log.WithError(err).Errorf("failed to find team members: %s", fileLocation)
				continue
			}
			var newMembers []sdk.TeamMember
			if err = json.Unmarshal(rawMembers, &newMembers); err != nil {
				log.WithError(err).Errorf("failed to unmarshal file: %s", fileLocation)
				continue
			}
			for _, member := range newMembers {
				_, err = s.GetAdminClient().AddTeamMember(ctx, member.TeamId, member.UserId)
				if err != nil {
					log.WithError(err).Errorf("failed to create team member for team %s with ID %d", newTeam.Name, member.UserId)
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
func (s *DashNGoImpl) ListTeams() []sdk.Team {
	ctx := context.Background()
	searchParams := sdk.WithPagesize(99999)
	pageTeams, err := s.GetAdminClient().SearchTeams(ctx, searchParams)
	if err != nil {
		log.Fatal(err)
	}
	return pageTeams.Teams
}

// Get a specific Team
// Return nil if team cannot be found
func (s *DashNGoImpl) getTeam(teamName string) *sdk.Team {
	teams := s.ListTeams()
	var team *sdk.Team
	for ndx, item := range teams {
		if item.Name == teamName {
			team = &teams[ndx]
			break
		}
	}
	return team
}

// Delete a specific Team
func (s *DashNGoImpl) DeleteTeam(teamName string) (*sdk.StatusMessage, error) {
	ctx := context.Background()
	team := s.getTeam(teamName)
	if team == nil {
		return nil, fmt.Errorf("team:  '%s' could not be found", teamName)
	}
	msg, err := s.GetAdminClient().DeleteTeam(ctx, team.ID)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to delete team: '%s'", teamName)
		log.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}
	return &msg, nil
}
