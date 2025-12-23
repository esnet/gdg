package backup

import (
	"context"
	"log"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newAlertingNotificationCommand() simplecobra.Commander {
	description := "Manage Alerting Notification"
	return &support.SimpleCommand{
		NameP: "notifications",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"notification", "notify", "n"}
		},
		CommandsList: []simplecobra.Commander{
			newListAlertNotificationCmd(),
			newDownloadAlertNotificationCmd(),
			newClearAlertNotificationCmd(),
			newUploadAlertNotificationCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newUploadAlertNotificationCmd() simplecobra.Commander {
	description := "Upload all alert notification for the given Organization"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"uid"})
			slog.Info("Uploading all alert notification for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			files, err := rootCmd.GrafanaSvc().UploadAlertNotifications()
			if err != nil {
				log.Fatal("unable to upload Orgs notification alerts", slog.Any("err", err))
			}
			rootCmd.TableObj.AppendHeader(table.Row{"receiver", "matchers"})
			for _, link := range files.Routes {
				rootCmd.TableObj.AppendRow(table.Row{link.Receiver, link.ObjectMatchers})
			}
			rootCmd.Render(cd.CobraCommand, files)
			return nil
		},
	}
}

func newClearAlertNotificationCmd() simplecobra.Commander {
	description := "Clear all alert notification for the given Organization"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Deleting all alert notification for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			err := rootCmd.GrafanaSvc().ClearAlertNotifications()
			if err != nil {
				log.Fatal("unable to deleting Orgs notification alerts", slog.Any("err", err))
			}

			slog.Info("All notifications alerts have been cleared")
			return nil
		},
	}
}

func newListAlertNotificationCmd() simplecobra.Commander {
	description := "List all alert notification for the given Organization"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"Receiver", "ObjectMatchers"})
			slog.Info("Listing alert notification for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			data, err := rootCmd.GrafanaSvc().ListAlertNotifications()
			if err != nil {
				log.Fatal("unable to retrieve Orgs notification alerts", slog.Any("err", err))
			}
			if len(data.Routes) == 0 {
				slog.Info("No alert notifications found")
			} else {
				for _, link := range data.Routes {
					rootCmd.TableObj.AppendRow(table.Row{link.Receiver, link.ObjectMatchers})
				}
				rootCmd.Render(cd.CobraCommand, data)
			}
			return nil
		},
	}
}

func newDownloadAlertNotificationCmd() simplecobra.Commander {
	description := "Download all alert notification for the given Organization"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"uid", "folderUid", "ruleGroup", "Title", "provenance", "data"})
			slog.Info("Downloading alert notification for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			file, err := rootCmd.GrafanaSvc().DownloadAlertNotifications()
			if err != nil {
				log.Fatal("unable to retrieve Orgs notification alerts", slog.Any("err", err))
			}
			if err != nil {
				slog.Error("unable to download alert templates")
			} else {
				slog.Info("alert templates successfully downloaded", slog.Any("file", file))
			}
			return nil
		},
	}
}
