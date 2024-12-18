package backup

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
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

func logWarning() {
	slog.Warn("GDG does not manage the 'email receiver' entity.  It has a very odd behavior compared to all " +
		"other entities. If you need to manage email contacts, please create a new contact.  GDG will ignore the default contact.")
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
			rootCmd.TableObj.AppendHeader(table.Row{"uid", "name", "type", "settings"})
			contactPoints, err := rootCmd.GrafanaSvc().ListContactPoints()
			slog.Info("Listing contact points for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			logWarning()
			if err != nil {
				log.Fatal("unable to retrieve Orgs contact points", slog.Any("err", err))
			}
			if len(contactPoints) == 0 {
				slog.Info("No contact points found")
			} else {
				for _, link := range contactPoints {
					rawBytes, err := json.Marshal(link.Settings)
					if err != nil {
						slog.Warn("unable to marshall settings to valid JSON")
					}
					typeVal := ""
					if link.Type != nil {
						typeVal = *link.Type
					}
					rootCmd.TableObj.AppendRow(table.Row{link.UID, link.Name, typeVal, string(rawBytes)})
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
			slog.Info("Download contact points for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))
			logWarning()
			file, err := rootCmd.GrafanaSvc().DownloadContactPoints()
			if err != nil {
				slog.Error("unable to download contact points")
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
			slog.Info("Upload contact points for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))
			newItems, err := rootCmd.GrafanaSvc().UploadContactPoints()
			logWarning()
			if err != nil {
				slog.Error("unable to upload contact points", slog.Any("err", err))
			} else {
				slog.Info("contact points successfully uploaded")
				rootCmd.TableObj.AppendHeader(table.Row{"name"})
				for _, item := range newItems {
					rootCmd.TableObj.AppendRow(table.Row{item})
				}

				rootCmd.Render(cd.CobraCommand, newItems)
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
			slog.Info("Clear contact points for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))
			removedItems, err := rootCmd.GrafanaSvc().ClearContactPoints()
			logWarning()
			if err != nil {
				slog.Error("unable to clear Contact Points")
			} else {
				slog.Info("Contact Points successfully removed")
				rootCmd.TableObj.AppendHeader(table.Row{"name"})
				for _, item := range removedItems {
					rootCmd.TableObj.AppendRow(table.Row{item})
				}
			}
			return nil
		},
	}
}
