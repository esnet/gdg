package service

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/tools"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/admin_users"
	"github.com/grafana/grafana-openapi-client-go/client/signed_in_user"
	"github.com/grafana/grafana-openapi-client-go/client/users"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/pretty"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
)

// UsersApi Contract definition
type UsersApi interface {
	//User
	ListUsers(filter filters.Filter) []*models.UserSearchHitDTO
	DownloadUsers(filter filters.Filter) []string
	UploadUsers(filter filters.Filter) []models.UserProfileDTO
	PromoteUser(userLogin string) (string, error)
	DeleteAllUsers(filter filters.Filter) []string
	GetUserInfo() (*models.UserProfileDTO, error)
}

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

func DefaultUserPassword(username string) string {
	if username == "admin" {
		return ""
	}

	username = username + ".json"
	//generate user password
	h := sha256.New()
	password := func() string {
		h.Write([]byte(username))
		hash := h.Sum(nil)
		password := fmt.Sprintf("%x", hash)
		return password
	}()

	return password
}

// GetUserInfo get signed-in user info, requires Basic authentication
func (s *DashNGoImpl) GetUserInfo() (*models.UserProfileDTO, error) {
	p := signed_in_user.NewGetSignedInUserParams()
	userInfo, err := s.client.SignedInUser.GetSignedInUser(p, s.getBasicAuth())
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

	userPath := BuildResourceFolder("", config.UserResource)
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

func (s *DashNGoImpl) UploadUsers(filter filters.Filter) []models.UserProfileDTO {
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.UserResource), false)
	if err != nil {
		slog.Error("failed to list files in directory for userListings", "err", err)
	}
	var userListings []models.UserProfileDTO
	var rawUser []byte
	userList := s.ListUsers(filter)
	var currentUsers = make(map[string]*models.UserSearchHitDTO, 0)

	//Build current User Mapping
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

			//generate user password
			password := DefaultUserPassword(file)

			var data = make(map[string]interface{}, 0)
			if err = json.Unmarshal(rawUser, &data); err != nil {
				slog.Error("failed to unmarshall file", "filename", fileLocation, "err", err)
				continue
			}
			data["password"] = password

			//Get raw version of payload once more with password
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
			userCreated, err := s.client.AdminUsers.AdminCreateUser(params, s.getBasicAuth())
			if err != nil {
				slog.Error("Failed to create user for file", "filename", fileLocation, "err", err)
				continue
			}
			p := users.NewGetUserByIDParams()
			p.UserID = userCreated.Payload.ID
			resp, err := s.client.Users.GetUserByID(p, s.getBasicAuth())
			if err != nil {
				slog.Error("unable to read user back from grafana", "username", newUser.Email, "userID", userCreated.GetPayload().ID)
				continue
			}
			userListings = append(userListings, *resp.Payload)
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
	params.Page = tools.PtrOf(int64(1))
	params.Perpage = tools.PtrOf(int64(5000))
	usersList, err := s.client.Users.SearchUsers(params, s.getAuth())
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
		_, err := s.client.AdminUsers.AdminDeleteUser(params, s.getBasicAuth())
		if err == nil {
			deletedUsers = append(deletedUsers, user.Email)
		}
	}
	return deletedUsers

}

// PromoteUser promote the user to have Admin Access
func (s *DashNGoImpl) PromoteUser(userLogin string) (string, error) {

	//Get all users
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

	promoteUserParam := admin_users.NewAdminUpdateUserPermissionsParams()
	promoteUserParam.UserID = user.ID
	promoteUserParam.Body = &models.AdminUpdateUserPermissionsForm{
		IsGrafanaAdmin: true,
	}

	msg, err := s.client.AdminUsers.AdminUpdateUserPermissions(promoteUserParam, s.getBasicAuth())
	if err != nil {
		errorMsg := fmt.Sprintf("failed to promote user: '%s'", userLogin)
		slog.Error("failed to promote user", "username", userLogin, "err", err)
		return "", errors.New(errorMsg)
	}

	return msg.GetPayload().Message, nil

}

// getUserById get the user by ID
func (s *DashNGoImpl) getUserById(userId int64) (*models.UserProfileDTO, error) {
	p := users.NewGetUserByIDParams()
	p.UserID = userId
	resp, err := s.client.Users.GetUserByID(p, s.getAuth())
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}
