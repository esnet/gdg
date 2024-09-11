package tools

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newOrgUsersCommand() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "users",
		Short: "Manage organization users",
		Long:  "Manager organization users",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"user"}
		},
		CommandsList: []simplecobra.Commander{
			newListUsers(),
			newAddUserRoleCmd(),
			newDeleteUserRoleCmd(),
			newGetUserOrgCmd(),
			newUpdateUserRoleCmd(),
		},
	}
}

func newGetUserOrgCmd() simplecobra.Commander {
	description := "display org associated with user"
	return &support.SimpleCommand{
		NameP: "currentOrg",
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
				rootCmd.Render(cd.CobraCommand, map[string]interface{}{"id": org.ID, "name": org.Name})
			}
			return nil
		},
	}
}

func newListUsers() simplecobra.Commander {
	description := "list <orgId> list an Organization users"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"listUsers"}
		},
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
				rootCmd.Render(cd.CobraCommand, users)
			}
			return nil
		},
	}
}

func newUpdateUserRoleCmd() simplecobra.Commander {
	description := "updateRole <orgSlugName> <userId> <role>"
	return &support.SimpleCommand{
		NameP: "updateRole",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"updateUserRole"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("requires the following parameters to be specified: [<orgId> <userId> <role>]\nValid roles are: [%s]", strings.Join(getBasicRoles(), ", "))
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
	description := "add <orgSlugName> <userId> <role>"
	return &support.SimpleCommand{
		NameP: "add",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"addUser", "addUsers"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("requires the following parameters to be specified: [<orgSlugName> <userId> <role>]\nValid roles are: [%s]", strings.Join(getBasicRoles(), ", "))
			}
			orgSlug := args[0]
			userId, err := strconv.ParseInt(args[1], 10, 64)
			role := args[2]
			if err != nil {
				log.Fatal("unable to parse userId to numeric value")
			}
			slog.Info("Add user to org for context",
				slog.Any("context", config.Config().GetGDGConfig().GetContext()),
				slog.Any("organization", config.Config().GetDefaultGrafanaConfig().OrganizationName),
			)
			if !validBasicRole(role) {
				log.Fatalf("Invalid role specified, '%s'.  Valid roles are:[%s]", role, strings.Join(getBasicRoles(), ", "))
			}
			rootCmd.TableObj.AppendHeader(table.Row{"login", "orgId", "name", "email", "role"})
			err = rootCmd.GrafanaSvc().AddUserToOrg(role, orgSlug, userId)
			if err != nil {
				slog.Error("Unable to add user to Org", slog.Any("err", err.Error()))
			} else {
				slog.Info("User has been add to Org", slog.Any("userId", userId), slog.String("organization", orgSlug))
			}
			return nil
		},
	}
}

func newDeleteUserRoleCmd() simplecobra.Commander {
	description := "deleteUser <orgSlug> <userId> removes a user from the given Organization (This will NOT delete the actual user from Grafana)"
	return &support.SimpleCommand{
		NameP: "delete",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"deleteUser", "remove"}
		},
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
				slog.Error("Unable to remove user from Org", slog.Any("err", err.Error()))
			} else {
				slog.Info("User has been removed from Org", "userId", orgSlug)
			}
			return nil
		},
	}
}
