package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/esnet/gdg/internal/tools/ptr"
	"github.com/esnet/gdg/pkg/config/domain"
	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/service/filters/v2"
	"github.com/tidwall/gjson"

	"github.com/esnet/gdg/internal/service/filters"

	"github.com/grafana/grafana-openapi-client-go/client/teams"
	"github.com/grafana/grafana-openapi-client-go/models"
	"golang.org/x/exp/maps"
)

type UserPermission models.PermissionType

const (
	AdminUserPermission = 4
)

func setupTeamReader(filterObj filters.V2Filter) {
	obj := models.TeamDTO{}
	err := filterObj.RegisterReader(reflect.TypeOf(obj), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.(models.TeamDTO)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.Name:
			return ptr.ValueOrDefault(val.Name, ""), nil

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Team Filter, obj entity reader failed, aborting.")
	}
	err = filterObj.RegisterReader(reflect.TypeOf([]byte{}), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.Name:
			{
				r := gjson.GetBytes(val, "name")
				if !r.Exists() || r.IsArray() {
					return nil, fmt.Errorf("no valid connection name found")
				}
				return r.String(), nil

			}

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Team Filter, json reader failed, aborting.")
	}
}

func NewTeamFilter(entries ...string) filters.V2Filter {
	filterObj := v2.NewBaseFilter()
	setupTeamReader(filterObj)
	filterObj.AddValidation(filters.Name, func(value any, expected any) error {
		val, expectedValue, convErr := v2.GetParams[string](value, expected, filters.Name)
		if convErr != nil {
			return convErr
		}
		if expectedValue == "" {
			return nil
		}
		if val != expectedValue {
			return fmt.Errorf("failed Team Name filter, expected %v, got %v", expectedValue, val)
		}
		return nil
	}, entries[0])

	return filterObj
}

// DownloadTeams fetches all teams for a given Org
func (s *DashNGoImpl) DownloadTeams(filter filters.V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO {
	teamListing := maps.Keys(s.ListTeams(filter))
	importedTeams := make(map[*models.TeamDTO][]*models.TeamMemberDTO)
	teamPath := BuildResourceFolder(s.grafanaConf, "", domain.TeamResource, s.isLocal(), s.GetGlobals().ClearOutput)
	for ndx, team := range teamListing {
		// Teams
		teamFileName := filepath.Join(teamPath, GetSlug(ptr.ValueOrDefault(team.Name, "")), "team.json")
		teamData, err := json.MarshalIndent(&teamListing[ndx], "", "\t")
		if err != nil {
			slog.Error("could not serialize team object for team name", "teamName", team.Name)
			continue
		}
		// Members
		memberFileName := filepath.Join(teamPath, GetSlug(ptr.ValueOrDefault(team.Name, "")), "members.json")
		members, err := s.GetClient().Teams.GetTeamMembers(fmt.Sprintf("%d", ptr.ValueOrDefault(team.ID, 0)))
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

// UploadTeams Export Teams
func (s *DashNGoImpl) UploadTeams(filter filters.V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO {
	orgName := s.grafanaConf.GetOrganizationName()
	filesInDir, err := s.storage.FindAllFiles(s.grafanaConf.GetPath(domain.TeamResource, orgName), true)
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
			if !filter.ValidateAll(rawTeam) {
				slog.Debug("Skipping file, failed Team filter", "file", fileLocation)
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
			teamCreated, teamCreatedErr := s.GetClient().Teams.CreateTeam(p)
			if teamCreatedErr != nil {
				slog.Error("failed to create team for file", "filename", fileLocation, "err", teamCreatedErr)
				continue
			}

			newTeam.ID = ptr.Of(teamCreated.GetPayload().TeamID)
			// Export Team Members (if exist)
			var currentMembers []*models.TeamMemberDTO
			var rawMembers []byte

			teamMemberLocation := filepath.Join(s.grafanaConf.GetPath(domain.TeamResource, orgName), GetSlug(ptr.ValueOrDefault(newTeam.Name, "")), "members.json")
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
				_, addTeamErr := s.addTeamMember(newTeam, member)
				if addTeamErr != nil {
					slog.Error("failed to create team member for team", "teamName", newTeam.Name, "MemberID", member.UserID, "err", addTeamErr)
				} else {
					currentMembers = append(currentMembers, member)
				}
			}
			exportedTeams[newTeam] = currentMembers
		}
	}
	return exportedTeams
}

// ListTeams List all Teams in a given org
func (s *DashNGoImpl) ListTeams(filter filters.V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO {
	result := make(map[*models.TeamDTO][]*models.TeamMemberDTO)
	p := teams.NewSearchTeamsParams()
	p.Perpage = ptr.Of[int64](99999)
	data, err := s.GetClient().Teams.SearchTeams(p)
	if err != nil {
		log.Fatal("unable to list teams")
	}

	getTeamMembers := func(team *models.TeamDTO) {
		if ptr.ValueOrDefault(team.MemberCount, 0) > 0 {
			result[team] = s.listTeamMembers(ptr.ValueOrDefault(team.ID, 0))
		} else {
			result[team] = nil
		}
	}

	for _, team := range data.GetPayload().Teams {
		if filter != nil {
			if filter.Validate(filters.Name, *team) {
				getTeamMembers(team)
			}
		} else {
			getTeamMembers(team)
		}
	}

	return result
}

// DeleteTeam removes all Teams
func (s *DashNGoImpl) DeleteTeam(filter filters.V2Filter) ([]*models.TeamDTO, error) {
	teamListing := maps.Keys(s.ListTeams(filter))
	var result []*models.TeamDTO
	for _, team := range teamListing {
		if !filter.ValidateAll(*team) || team.ID == nil {
			slog.Error("Team failed basic validation, could not delete entry")
			continue
		}
		_, err := s.GetClient().Teams.DeleteTeamByID(fmt.Sprintf("%d", *team.ID))
		if err != nil {
			slog.Error("failed to delete team", "teamName", *team.Name)
			continue
		}
		result = append(result, team)
	}

	return result, nil
}

// List Team Members of specific Team
func (s *DashNGoImpl) listTeamMembers(teamID int64) []*models.TeamMemberDTO {
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
		log.Fatal(fmt.Errorf("team:  '%s' could not be found", ptr.ValueOrDefault(team.Name, "")))
	}
	users := s.ListUsers(NewUserFilter(""))
	user, _ := lo.Find(users, func(item *models.UserSearchHitDTO) bool {
		return item.Login == userDTO.Login
	})

	if user == nil {
		log.Fatal(fmt.Errorf("user:  '%s' could not be found", userDTO.Login))
	}
	body := &models.AddTeamMemberCommand{UserID: ptr.Of(user.ID)}
	msg, err := s.GetClient().Teams.AddTeamMember(fmt.Sprintf("%d", ptr.ValueOrDefault(team.ID, 0)), body)
	if err != nil {
		slog.Info(err.Error())
		errorMsg := fmt.Sprintf("failed to add member '%s' to team '%s'", userDTO.Login, ptr.ValueOrDefault(team.Name, ""))
		slog.Error(errorMsg)
		return "", errors.New(errorMsg)
	}
	if userDTO.Permission == AdminUserPermission {
		adminPatch := teams.NewUpdateTeamMemberParams()
		adminPatch.TeamID = fmt.Sprintf("%d", ptr.ValueOrDefault(team.ID, 0))
		adminPatch.UserID = user.ID
		adminPatch.Body = &models.UpdateTeamMemberCommand{Permission: AdminUserPermission}
		response, updateErr := s.GetClient().Teams.UpdateTeamMember(adminPatch)
		if updateErr != nil {
			return "", updateErr
		}
		slog.Debug("Updated permissions for user on team ", "username", userDTO.Name, "teamName", ptr.ValueOrDefault(team.Name, ""), "message", response.GetPayload().Message)
	}

	return msg.GetPayload().Message, nil
}
