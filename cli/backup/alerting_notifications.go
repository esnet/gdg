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
	description := "Manage Alerting Notification Policies"
	return &support.SimpleCommand{
		NameP: "policy",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"policies", "notifications", "notification", "notify", "n"}
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
	description := "Upload all alert notification policies for the given Organization"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"uid"})
			slog.Info("Uploading all alert notification policies for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			files, err := rootCmd.GrafanaSvc().UploadAlertNotifications()
			if err != nil {
				log.Fatal("unable to upload Orgs notification policies alerts", slog.Any("err", err))
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
	description := "Clear all alert notification policies for the given Organization"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Deleting all alert notification policies for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			err := rootCmd.GrafanaSvc().ClearAlertNotifications()
			if err != nil {
				log.Fatal("unable to deleting Orgs notification policies alerts", slog.Any("err", err))
			}

			slog.Info("All notifications policies have been cleared")
			return nil
		},
	}
}

func newListAlertNotificationCmd() simplecobra.Commander {
	description := "List all alert notification policies for the given Organization"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"Receiver", "ObjectMatchers"})
			slog.Info("Listing alert notification policies for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			data, err := rootCmd.GrafanaSvc().ListAlertNotifications()
			if err != nil {
				log.Fatal("unable to retrieve Orgs notification policies", slog.Any("err", err))
			}
			if len(data.Routes) == 0 {
				slog.Info("No alert notifications policies found")
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
	description := "Download all alert notification policies for the given Organization"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"uid", "folderUid", "ruleGroup", "Title", "provenance", "data"})
			slog.Info("Downloading alert notification policies for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			file, err := rootCmd.GrafanaSvc().DownloadAlertNotifications()
			if err != nil {
				slog.Error("unable to download alert policies")
			} else {
				slog.Info("alert policies successfully downloaded", slog.Any("file", file))
			}
			return nil
		},
	}
}
