package tools

import (
	"context"
	"errors"
	"fmt"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

func newOrgCommand() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "organizations",
		Short: "Manage organizations",
		Long:  "Manage organizations",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()

		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"org", "orgs"}
		},
		CommandsList: []simplecobra.Commander{
			newSetOrgCmd(),
			newGetUserOrgCmd(),
			newGetTokenOrgCmd(),
			//Users
			newListUsers(),
			newUpdateUserRoleCmd(),
			newAddUserRoleCmd(),
			newDeleteUserRoleCmd(),
		},
	}

}

func newSetOrgCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "set",
		Short: "Set <OrgId>, 0 removes filter",
		Long:  "Set <OrgId>, 0	removes filter",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 1 {
				return errors.New("requires an Org ID and name")
			}
			OrgId := args[0]
			orgId, err := strconv.ParseInt(OrgId, 10, 64)
			if err != nil {
				log.Fatal("invalid Org ID, could not parse value to a numeric value")
			}
			err = rootCmd.GrafanaSvc().SetOrganization(orgId)
			if err != nil {
				log.WithError(err).Fatal("unable to set Org ID")
			}
			log.Infof("Succesfully set Org ID for context: %s", config.Config().AppConfig.GetContext())
			return nil

		},
	}

}

func newGetUserOrgCmd() simplecobra.Commander {
	description := "display org associated with user"
	return &support.SimpleCommand{
		NameP: "userOrg",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Listing organizations for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "name"})
			org := rootCmd.GrafanaSvc().GetUserOrganization()
			if org == nil {
				log.Info("No organizations found")
			} else {
				rootCmd.TableObj.AppendRow(table.Row{org.ID, org.Name})
				rootCmd.TableObj.Render()
			}
			return nil

		},
	}

}

func newGetTokenOrgCmd() simplecobra.Commander {
	description := "display org associated with token"
	return &support.SimpleCommand{
		NameP: "tokenOrg",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {

			log.Infof("Display token organization for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "name"})
			org := rootCmd.GrafanaSvc().GetTokenOrganization()
			if org == nil {
				log.Info("No organizations found")
			} else {
				rootCmd.TableObj.AppendRow(table.Row{org.ID, org.Name})
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}

}

func newListUsers() simplecobra.Commander {
	description := "list an Organization users"
	return &support.SimpleCommand{
		NameP: "listUsers",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 1 {
				return errors.New("requires an orgId to be specified")
			}
			orgId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Fatal("unable to parse orgId to numeric value")
			}
			log.Infof("Listing org users for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "login", "orgId", "name", "email", "role"})
			users := rootCmd.GrafanaSvc().ListOrgUsers(orgId)
			if len(users) == 0 {
				log.Info("No users found")
			} else {
				for _, user := range users {
					rootCmd.TableObj.AppendRow(table.Row{user.UserID, user.Login, user.OrgID, user.Name, user.Email, user.Role})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}

}

func newUpdateUserRoleCmd() simplecobra.Commander {
	description := "updateUserRole <orgId> <userId> <role>"
	return &support.SimpleCommand{
		NameP: "updateUserRole",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("requires the following parameters to be specified: [<orgId> <userId> <role>]\nValid roles are: [admin, editor, viewer]")
			}
			orgId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Fatal("unable to parse orgId to numeric value")
			}
			userId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Fatal("unable to parse userId to numeric value")
			}
			log.Infof("Listing org users for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"login", "orgId", "name", "email", "role"})
			err = rootCmd.GrafanaSvc().UpdateUserInOrg(args[2], userId, orgId)
			if err != nil {
				log.Error("Unable to update Org user")
			} else {
				log.Infof("User has been updated")
			}
			return nil
		},
	}
}

func newAddUserRoleCmd() simplecobra.Commander {
	description := "addUser <orgId> <userId> <role>"
	return &support.SimpleCommand{
		NameP: "addUser",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("requires the following parameters to be specified: [<orgId> <userId> <role>]\nValid roles are: [admin, editor, viewer]")
			}
			orgId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Fatal("unable to parse orgId to numeric value")
			}
			userId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Fatal("unable to parse userId to numeric value")
			}
			log.Infof("Add user to org for context: '%s'", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"login", "orgId", "name", "email", "role"})
			err = rootCmd.GrafanaSvc().AddUserToOrg(args[2], userId, orgId)
			if err != nil {
				log.Error("Unable to add user to Org")
			} else {
				log.Infof("User has been add to Org")
			}
			return nil
		},
	}
}

func newDeleteUserRoleCmd() simplecobra.Commander {
	description := "deleteUser removes a user from the given Organization (This will NOT delete the actual user from Grafana)"
	return &support.SimpleCommand{
		NameP: "deleteUser",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("requires the following parameters to be specified: [<orgId> <userId>]")
			}
			orgId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Fatal("unable to parse orgId to numeric value")
			}
			userId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Fatal("unable to parse userId to numeric value")
			}
			log.Infof("Update org for context: '%s'", config.Config().AppConfig.GetContext())
			err = rootCmd.GrafanaSvc().DeleteUserFromOrg(userId, orgId)
			if err != nil {
				log.Error("Unable to remove user from Org")
			} else {
				log.Infof("User has been removed from Org: %s", args[0])
			}
			return nil
		},
	}
}
