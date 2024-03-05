package backup

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/gosimple/slug"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"log/slog"
	"sort"
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
			cmd.PersistentFlags().StringP("org-name", "o", "", "when set to true, bypass confirmation prompts")
		},

		InitCFunc: func(cd *simplecobra.Commandeer, r *support.RootCommand) error {
			r.GrafanaSvc().InitOrganizations()
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
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			filter := service.NewOrganizationFilter(parseOrganizationGlobalFlags(cd.CobraCommand)...)
			slog.Info("Listing organizations for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"id", "organization Name", "org slug ID"})
			listOrganizations := rootCmd.GrafanaSvc().ListOrganizations(filter)
			sort.Slice(listOrganizations, func(a, b int) bool {
				return listOrganizations[a].ID < listOrganizations[b].ID
			})
			if len(listOrganizations) == 0 {
				slog.Info("No organizations found")
			} else {
				for _, org := range listOrganizations {
					rootCmd.TableObj.AppendRow(table.Row{org.ID, org.Name, slug.Make(org.Name)})
				}
				rootCmd.TableObj.Render()
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
			slog.Info("Downloading organizations for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			filter := service.NewOrganizationFilter(parseOrganizationGlobalFlags(cd.CobraCommand)...)
			listOrganizations := rootCmd.GrafanaSvc().DownloadOrganizations(filter)
			if len(listOrganizations) == 0 {
				slog.Info("No organizations found")
			} else {
				for _, org := range listOrganizations {
					rootCmd.TableObj.AppendRow(table.Row{org})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}

}
func newOrganizationsUploadCmd() simplecobra.Commander {
	description := "upload Orgs to grafana"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Uploading Folders for context: ", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			filter := service.NewOrganizationFilter(parseOrganizationGlobalFlags(cd.CobraCommand)...)
			folders := rootCmd.GrafanaSvc().UploadOrganizations(filter)
			if len(folders) == 0 {
				slog.Info("No Orgs were uploaded")
			} else {
				for _, folder := range folders {
					rootCmd.TableObj.AppendRow(table.Row{folder})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}

}
