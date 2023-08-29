package service

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	gapi "github.com/esnet/grafana-swagger-api-golang"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/admin_users"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/users"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/pretty"
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

func (s *DashNGoImpl) DownloadUsers(filter filters.Filter) []string {
	var (
		userData []byte
		err      error
	)

	userListing := s.ListUsers(filter)
	var importedUsers []string

	userPath := buildResourceFolder("", config.UserResource)
	for ndx, user := range userListing {
		if s.isAdmin(user.ID, user.Name) {
			log.Info("Skipping admin super user")
			continue
		}
		fileName := filepath.Join(userPath, fmt.Sprintf("%s.json", GetSlug(user.Login)))
		userData, err = json.Marshal(&userListing[ndx])
		if err != nil {
			log.Errorf("could not serialize user object for userId: %d", user.ID)
			continue
		}
		if err = s.storage.WriteFile(fileName, pretty.Pretty(userData)); err != nil {
			log.WithError(err).Errorf("for %s\n", user.Login)
		} else {
			importedUsers = append(importedUsers, fileName)
		}

	}

	return importedUsers

}

func (s *DashNGoImpl) isAdmin(id int64, name string) bool {
	return id == 1 || name == "admin"
}

func (s *DashNGoImpl) UploadUsers(filter filters.Filter) []models.UserProfileDTO {
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.UserResource), false)
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for userListings")
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
				log.WithError(err).Errorf("failed to read file: %s", fileLocation)
				continue
			}
			if val, ok := currentUsers[filepath.Base(file)]; ok {
				log.Warnf("User %s already exist, skipping", val.Login)
				continue
			}
			var newUser models.AdminCreateUserForm

			//generate user password
			password := DefaultUserPassword(file)

			var data = make(map[string]interface{}, 0)
			if err = json.Unmarshal(rawUser, &data); err != nil {
				log.WithError(err).Errorf("failed to unmarshall file: %s", fileLocation)
				continue
			}
			data["password"] = password

			//Get raw version of payload once more with password
			if rawUser, err = json.Marshal(data); err != nil {
				log.WithError(err).Errorf("failed to marshall file: %s to include password", fileLocation)
			}

			if err = json.Unmarshal(rawUser, &newUser); err != nil {
				log.WithError(err).Errorf("failed to unmarshall file: %s", fileLocation)
				continue
			}

			if newUser.Name == "admin" {
				log.Info("Skipping admin user")
				continue
			}
			params := admin_users.NewAdminCreateUserParams()
			params.Body = &newUser
			userCreated, err := s.client.AdminUsers.AdminCreateUser(params, s.getAdminAuth())
			if err != nil {
				log.WithError(err).Errorf("Failed to create user for file: %s", fileLocation)
				continue
			}
			p := users.NewGetUserByIDParams()
			p.UserID = userCreated.Payload.ID
			resp, err := s.client.Users.GetUserByID(p, s.getAdminAuth())
			if err != nil {
				log.Errorf("unable to read user: %s, ID: %d back from grafana", newUser.Email, userCreated.Payload.ID)
				continue
			}
			userListings = append(userListings, *resp.Payload)
		}
	}

	return userListings
}

// ListUsers list all grafana users
func (s *DashNGoImpl) ListUsers(filter filters.Filter) []*models.UserSearchHitDTO {
	var filteredUsers []*models.UserSearchHitDTO
	params := users.NewSearchUsersParams()
	params.Page = gapi.ToPtr(int64(1))
	params.Perpage = gapi.ToPtr(int64(5000))
	usersList, err := s.extended.SearchUsers(params)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, entry := range usersList {
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
		if s.isAdmin(user.ID, user.Name) {
			log.Info("Skipping admin user")
			continue

		}
		params := admin_users.NewAdminDeleteUserParams()
		params.UserID = user.ID
		_, err := s.client.AdminUsers.AdminDeleteUser(params, s.getAdminAuth())
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

	msg, err := s.client.AdminUsers.AdminUpdateUserPermissions(promoteUserParam, s.getAdminAuth())
	if err != nil {
		errorMsg := fmt.Sprintf("failed to promote user: '%s'", userLogin)
		log.Error(errorMsg)
		return "", errors.New(errorMsg)
	}

	return msg.GetPayload().Message, nil

}
