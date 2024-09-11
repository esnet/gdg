package tools

import (
	"context"
	"log"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newOrgPreferenceCommand() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "preferences",
		Short: "Update organization preferences",
		Long:  "Update organization preferences",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"preference", "pref", "p", "prefs"}
		},
		CommandsList: []simplecobra.Commander{
			// Preferences
			newGetOrgPreferenceCmd(),
			newUpdateOrgPreferenceCmd(),
		},
	}
}

func newUpdateOrgPreferenceCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "set",
		Short: "Set --orgName [--homeDashUid, --theme, --weekstart] to set Org preferences",
		Long:  "Set --orgName [--homeDashUid, --theme, --weekstart] to set Org preferences",
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.PersistentFlags().StringP("orgName", "", "", "Organization Name")
			cmd.PersistentFlags().StringP("homeDashUid", "", "", "UID for the home dashboard")
			cmd.PersistentFlags().StringP("theme", "", "", "light, dark")
			cmd.PersistentFlags().StringP("weekstart", "", "", "day of the week (sunday, monday, etc)")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("update the org preferences")
			org, _ := cd.CobraCommand.Flags().GetString("orgName")
			home, _ := cd.CobraCommand.Flags().GetString("homeDashUid")
			theme, _ := cd.CobraCommand.Flags().GetString("theme")
			weekstart, _ := cd.CobraCommand.Flags().GetString("weekstart")
			if org == "" {
				log.Fatal("--orgName is a required parameter")
			}
			if home != "" && theme != "" && weekstart == "" {
				log.Fatal("At least one of [--homeDashUid, --theme, --weekstart] needs to be set")
			}

			preferences, err := rootCmd.GrafanaSvc().GetOrgPreferences(org)
			if err != nil {
				log.Fatal(err.Error())
			}
			if home != "" {
				preferences.HomeDashboardUID = home
			}
			if theme != "" {
				preferences.Theme = theme
			}
			if weekstart != "" {
				preferences.WeekStart = weekstart
			}

			err = rootCmd.GrafanaSvc().UploadOrgPreferences(org, preferences)
			if err != nil {
				log.Fatalf("Failed to update org preferences, %v", err)
			}
			slog.Info("Preferences update for organization", slog.Any("organization", org))

			return nil
		},
	}
}

func newGetOrgPreferenceCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "get",
		Short: "get <orgName> returns org preferences",
		Long:  "get <orgName> returns org preferences",
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.PersistentFlags().StringP("orgName", "", "", "Organization Name")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			orgName, _ := cd.CobraCommand.Flags().GetString("orgName")

			pref, err := rootCmd.GrafanaSvc().GetOrgPreferences(orgName)
			if err != nil {
				log.Fatal(err.Error())
			}

			rootCmd.TableObj.AppendHeader(table.Row{"field", "value"})
			rootCmd.TableObj.AppendRow(table.Row{"HomeDashboardUID", pref.HomeDashboardUID})
			rootCmd.TableObj.AppendRow(table.Row{"Theme", pref.Theme})
			rootCmd.TableObj.AppendRow(table.Row{"WeekStart", pref.WeekStart})

			rootCmd.Render(cd.CobraCommand, pref)

			return nil
		},
	}
}
