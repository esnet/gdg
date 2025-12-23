package tools

import (
	"context"
	"errors"
	"log"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
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
			newGetTokenOrgCmd(),
			// Users
			newOrgUsersCommand(),
			// Preferences
			newOrgPreferenceCommand(),
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
			if orgName == "" && slugName == "" {
				return errors.New("must set either --orgName or --orgSlugName flag")
			}
			if orgName != "" || slugName != "" {
				useSlug := false
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

func newGetTokenOrgCmd() simplecobra.Commander {
	description := "display org associated with token"
	return &support.SimpleCommand{
		NameP: "tokenOrg",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Display token organization for context'", "context", rootCmd.ConfigSvc().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "name"})
			org := rootCmd.GrafanaSvc().GetTokenOrganization()
			if org == nil {
				slog.Info("No tokens were found")
			} else {
				rootCmd.TableObj.AppendRow(table.Row{org.ID, org.Name})
				rootCmd.Render(cd.CobraCommand, map[string]any{"id": org.ID, "name": org.Name})
			}
			return nil
		},
	}
}
