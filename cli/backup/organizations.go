package backup

import (
	"context"
	"log/slog"
	"sort"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service"
	"github.com/gosimple/slug"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func parseOrganizationGlobalFlags(command *cobra.Command) []string {
	orgName, _ := command.Flags().GetString("org-name")
	return []string{orgName}
}

func newOrganizationsCommand() simplecobra.Commander {
	description := "Manage Grafana Organizations."
	return &support.SimpleCommand{
		NameP: "organizations",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"org", "orgs"}
			cmd.PersistentFlags().StringP("org-name", "o", "", "Filter by org name")
		},

		InitCFunc: func(cd *simplecobra.Commandeer, r *support.RootCommand) error {
			return nil
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		CommandsList: []simplecobra.Commander{
			newOrganizationsListCmd(),
			newOrganizationsDownloadCmd(),
			newOrganizationsUploadCmd(),
		},
	}
}

func newOrganizationsListCmd() simplecobra.Commander {
	description := "List Grafana Organizations."
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
			cmd.PersistentFlags().BoolP("with-preferences", "", false, "when set to true, Attempts to retrieve Orgs Preferences (Warning, this is slow due to Grafana current API design)")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			filter := service.NewOrganizationFilter(parseOrganizationGlobalFlags(cd.CobraCommand)...)
			includePreferences, _ := cd.CobraCommand.Flags().GetBool("with-preferences")

			headerRow := table.Row{"id", "organization Name", "org slug ID"}
			if includePreferences {
				headerRow = append(headerRow, table.Row{"HomeDashboardUID", "Theme", "WeekStart"}...)
			}
			rootCmd.TableObj.AppendHeader(headerRow)
			listOrganizations := rootCmd.GrafanaSvc().ListOrganizations(filter, includePreferences)
			slog.Info("Listing organizations for context",
				slog.Any("count", len(listOrganizations)),
				slog.Any("context", rootCmd.ConfigSvc().GetContext()))
			sort.Slice(listOrganizations, func(a, b int) bool {
				return listOrganizations[a].Organization.ID < listOrganizations[b].Organization.ID
			})
			if len(listOrganizations) == 0 {
				slog.Info("No organizations found")
			} else {
				for _, org := range listOrganizations {
					data := table.Row{
						org.Organization.ID,
						org.Organization.Name,
						slug.Make(org.Organization.Name),
					}
					if includePreferences {
						data = append(data, table.Row{
							org.Preferences.HomeDashboardUID,
							org.Preferences.Theme,
							org.Preferences.WeekStart,
						}...)
					}
					rootCmd.TableObj.AppendRow(data)

				}
				rootCmd.Render(cd.CobraCommand, listOrganizations)
			}
			return nil
		},
	}
}

func newOrganizationsDownloadCmd() simplecobra.Commander {
	description := "download Organizations"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Downloading organizations for context", "context", rootCmd.ConfigSvc().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			filter := service.NewOrganizationFilter(parseOrganizationGlobalFlags(cd.CobraCommand)...)
			listOrganizations := rootCmd.GrafanaSvc().DownloadOrganizations(filter)
			if len(listOrganizations) == 0 {
				slog.Info("No organizations found")
			} else {
				for _, org := range listOrganizations {
					rootCmd.TableObj.AppendRow(table.Row{org})
				}
				rootCmd.Render(cd.CobraCommand, listOrganizations)
			}
			return nil
		},
	}
}

func newOrganizationsUploadCmd() simplecobra.Commander {
	description := "upload Organizations to grafana"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Uploading Folders for context: ", "context", rootCmd.ConfigSvc().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			filter := service.NewOrganizationFilter(parseOrganizationGlobalFlags(cd.CobraCommand)...)
			organizations := rootCmd.GrafanaSvc().UploadOrganizations(filter)
			if len(organizations) == 0 {
				slog.Info("No Organizations were uploaded")
			} else {
				for _, folder := range organizations {
					rootCmd.TableObj.AppendRow(table.Row{folder})
				}
				rootCmd.Render(cd.CobraCommand, organizations)
			}
			return nil
		},
	}
}
