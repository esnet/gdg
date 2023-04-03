package api

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/gdg/config"
	gapi "github.com/esnet/grafana-swagger-api-golang"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/admin_users"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/api_keys"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/users"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/pretty"
	"os"
	"path/filepath"
	"strings"
)

// UsersApi Contract definition
type UsersApi interface {
	//User
	ListUsers() []*models.UserSearchHitDTO
	ImportUsers() []string
	ExportUsers() []models.UserProfileDTO
	PromoteUser(userLogin string) (string, error)
	DeleteAllUsers() []string
	//TODO: clean up and expose API Keys management via CLI
	CreateAdminAPIKey(name string) *models.NewAPIKeyResult
	DeleteAPIKey(id int64) error
	ClearAPIKeys() error
}

func (s *DashNGoImpl) ClearAPIKeys() error {
	p := api_keys.NewGetAPIkeysParams()
	keys, err := s.client.APIKeys.GetAPIkeys(p, s.getAuth())
	if err != nil {
		return err
	}
	for _, key := range keys.GetPayload() {
		err := s.DeleteAPIKey(key.ID)
		if err != nil {
			log.Warnf("Failed to delete API key %d named %s", key.ID, key.Name)
		}
	}

	return nil
}

func (s *DashNGoImpl) CreateAdminAPIKey(name string) *models.NewAPIKeyResult {
	p := api_keys.NewAddAPIkeyParams()
	p.Body = &models.AddCommand{
		Name: name,
		Role: "admin",
	}
	newKey, err := s.client.APIKeys.AddAPIkey(p, s.getAuth())
	if err != nil {
		log.Fatal("unable to create a new API Key")
	}
	return newKey.GetPayload()

}
func (s *DashNGoImpl) DeleteAPIKey(id int64) error {
	p := api_keys.NewDeleteAPIkeyParams()
	p.ID = id
	_, err := s.client.APIKeys.DeleteAPIkey(p, s.getAuth())
	if err != nil {
		return fmt.Errorf("failed to delete API Key: %d", id)
	}
	return nil

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

func (s *DashNGoImpl) ImportUsers() []string {
	var (
		userData []byte
		err      error
	)

	userListing := s.ListUsers()
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
		if err = s.storage.WriteFile(fileName, pretty.Pretty(userData), os.FileMode(int(0666))); err != nil {
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

func (s *DashNGoImpl) ExportUsers() []models.UserProfileDTO {
	filesInDir, err := s.storage.FindAllFiles(apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.UserResource), false)
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for userListings")
	}
	var userListings []models.UserProfileDTO
	var rawUser []byte
	userList := s.ListUsers()
	var currentUsers = make(map[string]*models.UserSearchHitDTO, 0)

	//Build current User Mapping
	for ndx, i := range userList {
		key := slug.Make(i.Login) + ".json"
		currentUsers[key] = userList[ndx]
	}

	for _, file := range filesInDir {
		fileLocation := filepath.Join(apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.UserResource), file)
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
func (s *DashNGoImpl) ListUsers() []*models.UserSearchHitDTO {
	params := users.NewSearchUsersParams()
	params.Page = gapi.ToPtr(int64(1))
	params.Perpage = gapi.ToPtr(int64(5000))
	usersList, err := s.extended.SearchUsers(params)
	if err != nil {
		log.Fatal(err.Error())
	}
	return usersList
}

// DeleteAllUsers
func (s *DashNGoImpl) DeleteAllUsers() []string {
	userListing := s.ListUsers()
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
	userListing := s.ListUsers()
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
