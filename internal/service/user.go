package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/types"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/admin_users"
	"github.com/grafana/grafana-openapi-client-go/client/users"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/pretty"
)

func NewUserFilter(label string) filters.Filter {
	filterEntity := filters.NewBaseFilter()
	if label == "" {
		return filterEntity
	}
	filterEntity.AddFilter(filters.AuthLabel, label)
	filterEntity.AddValidation(filters.DefaultFilter, func(i interface{}) bool {
		val, ok := i.(map[filters.FilterType]string)
		if !ok {
			return ok
		}
		if filterEntity.GetFilter(filters.AuthLabel) == "" {
			return true
		}
		return val[filters.AuthLabel] == filterEntity.GetFilter(filters.AuthLabel)
	})

	return filterEntity
}

// GetUserInfo get signed-in user info, requires Basic authentication
func (s *DashNGoImpl) GetUserInfo() (*models.UserProfileDTO, error) {
	userInfo, err := s.GetBasicAuthClient().SignedInUser.GetSignedInUser()
	if err == nil {
		return userInfo.GetPayload(), err
	}
	return nil, err
}

func (s *DashNGoImpl) DownloadUsers(filter filters.Filter) []string {
	var (
		userData []byte
		err      error
	)

	userListing := s.ListUsers(filter)
	var importedUsers []string

	userPath := BuildResourceFolder("", config.UserResource, s.isLocal(), s.globalConf.ClearOutput)
	for ndx, user := range userListing {
		if s.isAdminUser(user.ID, user.Name) {
			slog.Info("Skipping admin super user")
			continue
		}
		fileName := filepath.Join(userPath, fmt.Sprintf("%s.json", GetSlug(user.Login)))
		userData, err = json.Marshal(&userListing[ndx])
		if err != nil {
			slog.Error("could not serialize user object for userId", "userID", user.ID)
			continue
		}
		if err = s.storage.WriteFile(fileName, pretty.Pretty(userData)); err != nil {
			slog.Error("Failed to write file", "filename", user.Login, "err", err)
		} else {
			importedUsers = append(importedUsers, fileName)
		}

	}
	return importedUsers
}

func (s *DashNGoImpl) isAdminUser(id int64, name string) bool {
	return id == 1 || name == "admin"
}

func (s *DashNGoImpl) UploadUsers(filter filters.Filter) []types.UserProfileWithAuth {
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.UserResource), false)
	if err != nil {
		slog.Error("failed to list files in directory for userListings", "err", err)
	}
	var userListings []types.UserProfileWithAuth
	var rawUser []byte
	userList := s.ListUsers(filter)
	currentUsers := make(map[string]*models.UserSearchHitDTO, 0)

	// Build current User Mapping
	for ndx, i := range userList {
		key := slug.Make(i.Login) + ".json"
		currentUsers[key] = userList[ndx]
	}

	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.UserResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawUser, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
			if val, ok := currentUsers[filepath.Base(file)]; ok {
				slog.Warn("User already exist, skipping", "username", val.Login)
				continue
			}
			var newUser models.AdminCreateUserForm

			// generate user password
			password := s.grafanaConf.GetUserSettings().GetPassword(file)

			data := make(map[string]interface{})
			if err = json.Unmarshal(rawUser, &data); err != nil {
				slog.Error("failed to unmarshall file", "filename", fileLocation, "err", err)
				continue
			}
			data["password"] = password

			// Get raw version of payload once more with password
			if rawUser, err = json.Marshal(data); err != nil {
				slog.Error("failed to marshall file to include password", "filename", fileLocation, "err", err)
			}

			if err = json.Unmarshal(rawUser, &newUser); err != nil {
				slog.Error("failed to unmarshall file", "filename", fileLocation, "err", err)
				continue
			}

			if newUser.Name == "admin" {
				slog.Info("Skipping admin user")
				continue
			}
			params := admin_users.NewAdminCreateUserParams()
			params.Body = &newUser
			userCreated, err := s.GetBasicAuthClient().AdminUsers.AdminCreateUser(&newUser)
			if err != nil {
				slog.Error("Failed to create user for file", "filename", fileLocation, "err", err)
				continue
			}
			resp, err := s.GetBasicAuthClient().Users.GetUserByID(userCreated.Payload.ID)
			if err != nil {
				slog.Error("unable to read user back from grafana", "username", newUser.Email, "userID", userCreated.GetPayload().ID)
				continue
			}
			userListings = append(userListings, types.UserProfileWithAuth{UserProfileDTO: *resp.GetPayload(), Password: string(newUser.Password)})
		}
	}

	return userListings
}

// ListUsers list all grafana users
func (s *DashNGoImpl) ListUsers(filter filters.Filter) []*models.UserSearchHitDTO {
	if !s.grafanaConf.IsBasicAuth() {
		log.Fatal("User listing requires basic auth to be configured.  Token based listing is not supported")
	}
	var filteredUsers []*models.UserSearchHitDTO
	params := users.NewSearchUsersParams()
	params.Page = ptr.Of(int64(1))
	params.Perpage = ptr.Of(int64(5000))
	usersList, err := s.GetClient().Users.SearchUsers(params)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, entry := range usersList.GetPayload() {
		if len(entry.AuthLabels) == 0 {
			filteredUsers = append(filteredUsers, entry)
		} else if filter.ValidateAll(map[filters.FilterType]string{filters.AuthLabel: entry.AuthLabels[0]}) {
			filteredUsers = append(filteredUsers, entry)
		}
	}
	sort.Slice(filteredUsers, func(i, j int) bool {
		return filteredUsers[i].ID < filteredUsers[j].ID
	})
	return filteredUsers
}

// DeleteAllUsers remove all users excluding admin or anything matching the filter
func (s *DashNGoImpl) DeleteAllUsers(filter filters.Filter) []string {
	userListing := s.ListUsers(filter)
	var deletedUsers []string
	for _, user := range userListing {
		if s.isAdminUser(user.ID, user.Name) {
			slog.Info("Skipping admin user")
			continue

		}
		params := admin_users.NewAdminDeleteUserParams()
		params.UserID = user.ID
		_, err := s.GetBasicAuthClient().AdminUsers.AdminDeleteUser(user.ID)
		if err == nil {
			deletedUsers = append(deletedUsers, user.Email)
		}
	}
	return deletedUsers
}

// PromoteUser promote the user to have Admin Access
func (s *DashNGoImpl) PromoteUser(userLogin string) (string, error) {
	// Get all users
	userListing := s.ListUsers(filters.NewBaseFilter())
	var user *models.UserSearchHitDTO
	for ndx, item := range userListing {
		if item.Email == userLogin {
			user = userListing[ndx]
			break
		}
	}

	if user == nil {
		return "", fmt.Errorf("user: '%s' could not be found", userLogin)
	}
	requestBody := &models.AdminUpdateUserPermissionsForm{IsGrafanaAdmin: true}

	msg, err := s.GetBasicAuthClient().AdminUsers.AdminUpdateUserPermissions(user.ID, requestBody)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to promote user: '%s'", userLogin)
		slog.Error("failed to promote user", "username", userLogin, "err", err)
		return "", errors.New(errorMsg)
	}

	return msg.GetPayload().Message, nil
}

// getUserById get the user by ID
func (s *DashNGoImpl) getUserById(userId int64) (*models.UserProfileDTO, error) {
	resp, err := s.GetClient().Users.GetUserByID(userId)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}
