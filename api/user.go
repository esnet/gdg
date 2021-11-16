package api

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/grafana-tools/sdk"
	"github.com/netsage-project/gdg/apphelpers"
	"github.com/sirupsen/logrus"
)

func validateUserAPI(client *sdk.Client) {
	if client == nil || !apphelpers.GetCtxDefaultGrafanaConfig().AdminEnabled {
		logrus.Fatal("Missing Admin client, please check your config and ensure basic auth is configured")
		os.Exit(1)
	}
}

//ListUsers list all grafana users
func (s *DashNGoImpl) ListUsers() []sdk.User {
	ctx := context.Background()
	validateUserAPI(s.adminClient)
	users, err := s.adminClient.GetAllUsers(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	return users
}

//PromoteUser promote the user to have Admin Access
func (s *DashNGoImpl) PromoteUser(userLogin string) (*sdk.StatusMessage, error) {

	validateUserAPI(s.adminClient)
	ctx := context.Background()
	//Get all users
	users := s.ListUsers()
	var user *sdk.User
	for _, item := range users {
		if item.Login == userLogin {
			user = &item
			break
		}

	}

	if user == nil {
		return nil, fmt.Errorf("user: '%s' could not be found", userLogin)
	}

	role := sdk.UserPermissions{
		IsGrafanaAdmin: true,
	}
	msg, err := s.adminClient.UpdateUserPermissions(ctx, role, user.ID)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to promote user: '%s'", userLogin)
		logrus.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}

	return &msg, nil

}
