package api

import (
	"context"
	"errors"
	"fmt"
	"math"
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
func (s *DashNGoImpl) ImportTeams() []string {
	var (
		teamData []byte
	)
	ctx := context.Background()
	searchParams := sdk.WithPagesize(math.MaxUint64)
	pageTeams, err := s.GetAdminClient().SearchTeams(ctx, searchParams)
	if err != nil {
		log.Fatal(err)
	}
	teams := pageTeams.Teams
	importedTeams := []string{}
	teamPath := buildResourceFolder("", config.TeamResource)
	for ndx, team := range teams {
		fileName := filepath.Join(teamPath, fmt.Sprintf("%s.json", GetSlug(team.Name)))
		teamData, err = json.Marshal(&teams[ndx])
		if err != nil {
			log.Errorf("could not serialize team object for team name: %d", team.Name)
			continue
		}
		if err = s.storage.WriteFile(fileName, pretty.Pretty(teamData), os.FileMode(int(0666))); err != nil {
			log.WithError(err).Errorf("for %s\n", team.Name)
		} else {
			importedTeams = append(importedTeams, fileName)
		}
	}
	return importedTeams
}

// Export Teams
func (s *DashNGoImpl) ExportTeams() []sdk.Team {
	ctx := context.Background()
	filesInDir, err := s.storage.FindAllFiles(getResourcePath(config.TeamResource), false)
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for teams")
	}
	var teams []sdk.Team
	var rawTeam []byte
	for _, file := range filesInDir {
		fileLocation := filepath.Join(getResourcePath(config.TeamResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawTeam, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file: %s", fileLocation)
				continue
			}
			var newTeam sdk.Team
			if err = json.Unmarshal(rawTeam, &newTeam); err != nil {
				log.WithError(err).Errorf("failed to unmarshall file: %s", fileLocation)
				continue
			}
			_, err = s.GetAdminClient().CreateTeam(ctx, newTeam)
			if err != nil {
				log.WithError(err).Errorf("failed to create team for file: %s", fileLocation)
				continue
			}
			teams = append(teams, newTeam)
		}
	}
	return teams
}

// List all Teams
func (s *DashNGoImpl) ListTeams() []sdk.Team {
	ctx := context.Background()
	searchParams := sdk.WithPagesize(math.MaxUint64)
	pageTeams, err := s.GetAdminClient().SearchTeams(ctx, searchParams)
	if err != nil {
		log.Fatal(err)
	}
	return pageTeams.Teams
}

// Get a specific Team
// Return nil if team cannot be found
func (s *DashNGoImpl) GetTeam(teamName string) *sdk.Team {
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
	team := s.GetTeam(teamName)
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
