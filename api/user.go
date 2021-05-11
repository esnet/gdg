package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/netsage-project/sdk"
	"github.com/sirupsen/logrus"
)

//ListUsers list all grafana users
func ListUsers(client *sdk.Client) []sdk.User {
	ctx := context.Background()
	users, err := client.GetAllUsers(ctx)
	if err != nil {
		logrus.Panic("Unable to get users")
	}
	return users
}

//PromoteUser promote the user to have Admin Access
func PromoteUser(client *sdk.Client, userLogin string) (*sdk.StatusMessage, error) {

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
		errorMsg := fmt.Sprintf("failed to promot user: '%s'", userLogin)
		logrus.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}

	return &msg, nil

}
