package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/config"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/pretty"
	"os"
	"path/filepath"
	"strings"
)

func (s *DashNGoImpl) ImportUsers() []string {
	var (
		userData []byte
	)
	ctx := context.Background()
	users, err := s.GetAdminClient().GetAllUsers(ctx)
	if err != nil {
		log.Fatal(err)
	}
	importedUsers := []string{}

	userPath := buildResourceFolder("", config.UserResource)
	for ndx, user := range users {
		if s.isAdmin(user) {
			log.Info("Skipping admin super user")
			continue
		}
		fileName := filepath.Join(userPath, fmt.Sprintf("%s.json", GetSlug(user.Login)))
		userData, err = json.Marshal(&users[ndx])
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

//Skips the admin super user
func (s *DashNGoImpl) isAdmin(user sdk.User) bool {
	return user.ID == 1 || user.Name == "admin"
}

func (s *DashNGoImpl) ExportUsers() []sdk.User {
	ctx := context.Background()
	filesInDir, err := s.storage.FindAllFiles(getResourcePath(config.UserResource), false)
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for users")
	}
	var users []sdk.User
	var rawUser []byte
	h := sha256.New()
	for _, file := range filesInDir {
		fileLocation := filepath.Join(getResourcePath(config.UserResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawUser, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file: %s", fileLocation)
				continue
			}
			var newUser sdk.User

			//generate user password
			password := func() string {
				h.Write([]byte(file))
				hash := h.Sum(nil)
				password := fmt.Sprintf("%x", hash)
				return password
			}()

			var data map[string]interface{} = make(map[string]interface{}, 0)
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

			if s.isAdmin(newUser) {
				log.Info("Skipping admin user")
				continue
			}
			_, err = s.GetAdminClient().CreateUser(ctx, newUser)
			if err != nil {
				log.WithError(err).Errorf("Failed to create user for file: %s", fileLocation)
				continue
			}
			users = append(users, newUser)
		}
	}

	return users
}

//ListUsers list all grafana users
func (s *DashNGoImpl) ListUsers() []sdk.User {
	ctx := context.Background()
	users, err := s.GetAdminClient().GetAllUsers(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return users
}

// DeleteAllUsers
func (s *DashNGoImpl) DeleteAllUsers() []string {
	ctx := context.Background()
	users := s.ListUsers()
	var deletedUsers []string
	for _, user := range users {
		if s.isAdmin(user) {
			log.Info("Skipping admin user")
			continue

		}
		_, err := s.GetAdminClient().DeleteUser(ctx, user.ID)
		if err == nil {
			deletedUsers = append(deletedUsers, user.Email)
		}
	}
	return deletedUsers

}

//PromoteUser promote the user to have Admin Access
func (s *DashNGoImpl) PromoteUser(userLogin string) (*sdk.StatusMessage, error) {

	ctx := context.Background()
	//Get all users
	users := s.ListUsers()
	var user *sdk.User
	for ndx, item := range users {
		if item.Login == userLogin {
			user = &users[ndx]
			break
		}

	}

	if user == nil {
		return nil, fmt.Errorf("user: '%s' could not be found", userLogin)
	}

	role := sdk.UserPermissions{
		IsGrafanaAdmin: true,
	}
	msg, err := s.GetAdminClient().UpdateUserPermissions(ctx, role, user.ID)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to promote user: '%s'", userLogin)
		log.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}

	return &msg, nil

}
