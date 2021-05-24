package api

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/netsage-project/grafana-dashboard-manager/config"
	"github.com/netsage-project/sdk"
	"github.com/sirupsen/logrus"
)

func validateUserAPI() {
	if config.GetDefaultGrafanaConfig().APIToken != "" {
		logrus.Error("The Admin HTTP API does not currently work with an API Token.  Please check your config")
		os.Exit(1)
	}

}

//ListUsers list all grafana users
func ListUsers(client *sdk.Client) []sdk.User {
	ctx := context.Background()
	validateUserAPI()
	users, err := client.GetAllUsers(ctx)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	return users
}

//PromoteUser promote the user to have Admin Access
func PromoteUser(client *sdk.Client, userLogin string) (*sdk.StatusMessage, error) {

	validateUserAPI()
	ctx := context.Background()
	//Get all users
	users := ListUsers(client)
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
	msg, err := client.UpdateUserPermissions(ctx, role, user.ID)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to promote user: '%s'", userLogin)
		logrus.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}

	return &msg, nil

}
