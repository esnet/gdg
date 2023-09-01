package tools

import (
	"errors"
	"fmt"
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

var orgCmd = &cobra.Command{
	Use:     "organizations",
	Aliases: []string{"org", "orgs"},
	Short:   "Manage Organizations",
	Long:    `Manage Grafana Organizations.`,
}

var setOrgCmd = &cobra.Command{
	Use:   "set",
	Short: "set <OrgId>, 0 removes filter",
	Long:  `set <OrgId>, 0 removes filter`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires an Org ID and name")
		}
		return nil
	},

	Run: func(command *cobra.Command, args []string) {
		ctx := args[0]
		orgId, err := strconv.ParseInt(ctx, 10, 64)
		if err != nil {
			log.Fatal("invalid Org ID, could not parse value to a numeric value")
		}
		err = cmd.GetGrafanaSvc().SetOrganization(orgId)
		if err != nil {
			log.WithError(err).Fatal("unable to set Org ID")
		}
		log.Infof("Succesfully set Org ID for context: %s", config.Config().AppConfig.GetContext())
	},
}

var getUserOrgCmd = &cobra.Command{
	Use:   "userOrg",
	Short: "display org associated with user",
	Long:  `display org associated with user`,
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Listing organizations for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"id", "name"})
		org := cmd.GetGrafanaSvc().GetUserOrganization()
		if org == nil {
			log.Info("No organizations found")
		} else {
			cmd.TableObj.AppendRow(table.Row{org.ID, org.Name})
			cmd.TableObj.Render()
		}

	},
}

var getTokenOrgCmd = &cobra.Command{
	Use:   "tokenOrg",
	Short: "display org associated with token",
	Long:  `display org associated with token`,
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Display token organization for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"id", "name"})
		org := cmd.GetGrafanaSvc().GetTokenOrganization()
		if org == nil {
			log.Info("No organizations found")
		} else {
			cmd.TableObj.AppendRow(table.Row{org.ID, org.Name})
			cmd.TableObj.Render()
		}

	},
}

var listUsers = &cobra.Command{
	Use:   "listUsers",
	Short: "listUsers <orgId>",
	Long:  `list an Organization users`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires an orgId to be specified")
		}
		return nil
	},
	Run: func(command *cobra.Command, args []string) {
		orgId, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal("unable to parse orgId to numeric value")
		}
		log.Infof("Listing org users for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"id", "login", "orgId", "name", "email", "role"})
		users := cmd.GetGrafanaSvc().ListOrgUsers(orgId)
		if len(users) == 0 {
			log.Info("No users found")
		} else {
			for _, user := range users {
				cmd.TableObj.AppendRow(table.Row{user.UserID, user.Login, user.OrgID, user.Name, user.Email, user.Role})
			}
			cmd.TableObj.Render()
		}

	},
}

var updateUserRole = &cobra.Command{
	Use:   "updateUserRole",
	Short: "updateUserRole <orgId> <userId> <role>",
	Long:  `updateUserRole an Organization users`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 3 {
			return fmt.Errorf("requires the following parameters to be specified: [<orgId> <userId> <role>]\nValid roles are: [admin, editor, viewer]")
		}
		return nil
	},
	Run: func(command *cobra.Command, args []string) {
		orgId, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal("unable to parse orgId to numeric value")
		}
		userId, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Fatal("unable to parse userId to numeric value")
		}
		log.Infof("Listing org users for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"login", "orgId", "name", "email", "role"})
		err = cmd.GetGrafanaSvc().UpdateUserInOrg(args[2], userId, orgId)
		if err != nil {
			log.Error("Unable to update Org user")
		} else {
			log.Infof("User has been updated")
		}
	},
}

var addUserRole = &cobra.Command{
	Use:   "addUser",
	Short: "addUser <orgId> <userId> <role>",
	Long:  `addUser to an Organization users`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 3 {
			return fmt.Errorf("requires the following parameters to be specified: [<orgId> <userId> <role>]\nValid roles are: [admin, editor, viewer]")
		}
		return nil
	},
	Run: func(command *cobra.Command, args []string) {
		orgId, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal("unable to parse orgId to numeric value")
		}
		userId, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Fatal("unable to parse userId to numeric value")
		}
		log.Infof("Add user to org for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"login", "orgId", "name", "email", "role"})
		err = cmd.GetGrafanaSvc().AddUserToOrg(args[2], userId, orgId)
		if err != nil {
			log.Error("Unable to add user to Org")
		} else {
			log.Infof("User has been add to Org")
		}
	},
}

var deleteUserFromOrg = &cobra.Command{
	Use:   "deleteUser",
	Short: "deleteUser <orgId> <userId>",
	Long:  `deleteUser removes a user from the given Organization (This will NOT delete the actual user from Grafana)`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires the following parameters to be specified: [<orgId> <userId>]")
		}
		return nil
	},
	Run: func(command *cobra.Command, args []string) {
		orgId, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatal("unable to parse orgId to numeric value")
		}
		userId, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Fatal("unable to parse userId to numeric value")
		}
		log.Infof("Update org for context: '%s'", config.Config().AppConfig.GetContext())
		err = cmd.GetGrafanaSvc().DeleteUserFromOrg(userId, orgId)
		if err != nil {
			log.Error("Unable to remove user from Org")
		} else {
			log.Infof("User has been removed from Org: %s", args[0])
		}
	},
}

func init() {
	toolsCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(setOrgCmd)
	orgCmd.AddCommand(getUserOrgCmd)
	orgCmd.AddCommand(getTokenOrgCmd)
	//Users
	orgCmd.AddCommand(listUsers)
	orgCmd.AddCommand(updateUserRole)
	orgCmd.AddCommand(addUserRole)
	orgCmd.AddCommand(deleteUserFromOrg)
}
