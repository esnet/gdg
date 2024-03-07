package tools

import (
	"context"
	"errors"
	"fmt"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
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
		Short: "Set --orgSlugName --orgName to set user Org",
		Long:  "Set --orgSlugName --orgName to set user Org",
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.PersistentFlags().StringP("orgName", "o", "", "Set user Org by Name (not slug)")
			cmd.PersistentFlags().StringP("orgSlugName", "", "", "Set user Org by slug name")

		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			orgName, _ := cd.CobraCommand.Flags().GetString("orgName")
			slugName, _ := cd.CobraCommand.Flags().GetString("orgSlugName")
			if orgName != "" || slugName != "" {
				var useSlug = false
				if slugName != "" {
					useSlug = true
					orgName = slugName
				}
				err := rootCmd.GrafanaSvc().SetOrganizationByName(orgName, useSlug)
				if err != nil {
					log.Fatal("unable to set Org ID, ", err.Error())
				}
			}

			rootCmd.GrafanaSvc().InitOrganizations()
			userOrg := rootCmd.GrafanaSvc().GetUserOrganization()
			slog.Info("New Org is now set to", slog.String("orgName", userOrg.Name))

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
			slog.Info("Listing organizations for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "name"})
			org := rootCmd.GrafanaSvc().GetUserOrganization()
			if org == nil {
				slog.Info("No organizations found")
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

			slog.Info("Display token organization for context'", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "name"})
			org := rootCmd.GrafanaSvc().GetTokenOrganization()
			if org == nil {
				slog.Info("No tokens were found")
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
			slog.Info("Listing org users for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "login", "orgId", "name", "email", "role"})
			users := rootCmd.GrafanaSvc().ListOrgUsers(orgId)
			if len(users) == 0 {
				slog.Info("No users found")
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
	description := "updateUserRole <orgSlugName> <userId> <role>"
	return &support.SimpleCommand{
		NameP: "updateUserRole",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("requires the following parameters to be specified: [<orgId> <userId> <role>]\nValid roles are: [admin, editor, viewer]")
			}
			orgSlug := args[0]
			roleName := args[2]
			userId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Fatal("unable to parse userId to numeric value")
			}
			slog.Info("Listing org users for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"login", "orgId", "name", "email", "role"})
			err = rootCmd.GrafanaSvc().UpdateUserInOrg(roleName, orgSlug, userId)
			if err != nil {
				slog.Error("Unable to update Org user")
			} else {
				slog.Info("User has been updated")
			}
			return nil
		},
	}
}

func newAddUserRoleCmd() simplecobra.Commander {
	description := "addUser <orgSlugName> <userId> <role>"
	return &support.SimpleCommand{
		NameP: "addUser",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("requires the following parameters to be specified: [<orgSlugName> <userId> <role>]\nValid roles are: [admin, editor, viewer]")
			}
			orgSlug := args[0]
			userId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Fatal("unable to parse userId to numeric value")
			}
			slog.Info("Add user to org for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"login", "orgId", "name", "email", "role"})
			err = rootCmd.GrafanaSvc().AddUserToOrg(args[2], orgSlug, userId)
			if err != nil {
				slog.Error("Unable to add user to Org")
			} else {
				slog.Info("User has been add to Org")
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
				return fmt.Errorf("requires the following parameters to be specified: [<orgSlugName> <userId>]")
			}
			orgSlug := args[0]
			userId, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Fatal("unable to parse userId to numeric value")
			}
			slog.Info("Update org for context", "context", config.Config().GetGDGConfig().GetContext())
			err = rootCmd.GrafanaSvc().DeleteUserFromOrg(orgSlug, userId)
			if err != nil {
				slog.Error("Unable to remove user from Org")
			} else {
				slog.Info("User has been removed from Org", "userId", args[0])
			}
			return nil
		},
	}
}
