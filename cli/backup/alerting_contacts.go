package backup

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newAlertingContactCommand() simplecobra.Commander {
	description := "Manage Alerting ContactPoints "
	return &support.SimpleCommand{
		NameP: "contactpoint",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"contact", "contacts", "contactpoints"}
		},
		CommandsList: []simplecobra.Commander{
			newListContactPointsCmd(),
			newClearContactPointsCmd(),
			newUploadContactPointsCmd(),
			newDownloadContactPointsCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newListContactPointsCmd() simplecobra.Commander {
	description := "List all contact points for the given Organization"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"uid", "name", "slug", "type", "provenance", "settings"})
			contactPoints := rootCmd.GrafanaSvc().ListContactPoints()
			slog.Info("Listing contact points for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))
			if len(contactPoints) == 0 {
				slog.Info("No contact points found")
			} else {
				for _, link := range contactPoints {
					rawBytes, err := json.Marshal(link.Settings)
					if err != nil {
						slog.Warn("unable to marshall settings to valid JSON")
					}
					rootCmd.TableObj.AppendRow(table.Row{link.UID, link.Name, service.GetSlug(link.Name), link.Type, link.Provenance, string(rawBytes)})
				}
				rootCmd.Render(cd.CobraCommand, contactPoints)
			}
			return nil
		},
	}
}

func newDownloadContactPointsCmd() simplecobra.Commander {
	description := "Download all contact points for the given Organization"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			file, err := rootCmd.GrafanaSvc().DownloadContactPoints()
			slog.Info("Download contact points for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))
			if err != nil {
				slog.Error("unable to contact point")
			} else {
				slog.Info("contact points successfully downloaded", slog.Any("file", file))
			}
			return nil
		},
	}
}

func newUploadContactPointsCmd() simplecobra.Commander {
	description := "Upload all contact points for the given Organization"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			removedItems, err := rootCmd.GrafanaSvc().UploadContactPoints()
			slog.Info("Upload contact points for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))
			if err != nil {
				slog.Error("unable to upload contact points", slog.Any("err", err))
			} else {
				slog.Info("contact points successfully uploaded")
				rootCmd.TableObj.AppendHeader(table.Row{"name"})
				for _, item := range removedItems {
					rootCmd.TableObj.AppendRow(table.Row{item})
				}

				rootCmd.Render(cd.CobraCommand, removedItems)
			}
			return nil
		},
	}
}

func newClearContactPointsCmd() simplecobra.Commander {
	description := "Clear all contact points for the given Organization"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			removedItems, err := rootCmd.GrafanaSvc().ClearContactPoints()
			slog.Info("Clear contact points for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))
			if err != nil {
				slog.Error("unable to contact point")
			} else {
				slog.Info("contact points successfully removed")
				rootCmd.TableObj.AppendHeader(table.Row{"name"})
				for _, item := range removedItems {
					rootCmd.TableObj.AppendRow(table.Row{item})
				}
			}
			return nil
		},
	}
}
