package backup

import (
	"context"
	"fmt"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"log/slog"
)

func newAlertNotificationsCommand() simplecobra.Commander {
	description := "Manage alert notification channels"
	return &support.SimpleCommand{
		NameP: "alertnotifications",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"an", "alertnotifications"}
		},
		CommandsList: []simplecobra.Commander{
			newListAlertNotificationsCmd(),
			newDownloadAlertNotificationsCmd(),
			newUploadAlertNotificationsCmd(),
			newClearAlertNotificationsCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}

}

func newClearAlertNotificationsCmd() simplecobra.Commander {
	description := "delete all alert notification channels from grafana"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Warn("Alert Notifications will be deprecated as of Grafana 9.0, this API may no longer work soon")
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})

			slog.Info("Clearing all alert notification channels for context",
				"context", config.Config().AppConfig.GetContext())
			deleted := rootCmd.GrafanaSvc().DeleteAllAlertNotifications()
			for _, item := range deleted {
				rootCmd.TableObj.AppendRow(table.Row{"alertnotification", item})
			}
			if len(deleted) == 0 {
				slog.Info("No alert notification channels were found. 0 removed")
			} else {
				slog.Info("alert notification channels were deleted", "count", len(deleted))
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}

func newUploadAlertNotificationsCmd() simplecobra.Commander {
	description := "upload all alert notification channels to grafana"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Warn("Alert Notifications will be deprecated as of Grafana 9.0, this API may no longer work soon")
			rootCmd.TableObj.AppendHeader(table.Row{"name", "id", "UID"})

			slog.Info("Exporting alert notification channels for context",
				"context", config.Config().AppConfig.GetContext())
			rootCmd.GrafanaSvc().UploadAlertNotifications()
			items := rootCmd.GrafanaSvc().ListAlertNotifications()
			for _, item := range items {
				rootCmd.TableObj.AppendRow(table.Row{item.Name, item.ID, item.UID})
			}
			if len(items) > 0 {
				rootCmd.TableObj.Render()
			} else {
				slog.Info("No alert notification channels found")
			}
			return nil
		},
	}
}

func newDownloadAlertNotificationsCmd() simplecobra.Commander {
	description := "download all alert notification channels from grafana to local filesystem"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Warn("Alert Notifications will be deprecated as of Grafana 9.0, this API may no longer work soon")
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})

			slog.Info("Downloading alert notification channels for context",
				"context", config.Config().AppConfig.GetContext())

			savedFiles := rootCmd.GrafanaSvc().DownloadAlertNotifications()
			for _, file := range savedFiles {
				rootCmd.TableObj.AppendRow(table.Row{"alertnotification", file})
			}
			rootCmd.TableObj.Render()
			return nil
		},
	}
}

func newListAlertNotificationsCmd() simplecobra.Commander {
	description := "List all alert notification channels from grafana"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Warn("Alert Notifications will be deprecated as of Grafana 9.0, this API may no longer work soon")

			rootCmd.TableObj.AppendHeader(table.Row{"id", "name", "slug", "type", "default", "url"})
			alertnotifications := rootCmd.GrafanaSvc().ListAlertNotifications()

			slog.Info("Listing alert notifications channels for context",
				"context", config.Config().AppConfig.GetContext())

			if len(alertnotifications) == 0 {
				slog.Info("No alert notifications found")
			} else {
				for _, link := range alertnotifications {
					url := fmt.Sprintf("%s/alerting/notification/%d/edit", config.Config().GetDefaultGrafanaConfig().URL, link.ID)
					rootCmd.TableObj.AppendRow(table.Row{link.ID, link.Name, service.GetSlug(link.Name), link.Type, link.IsDefault, url})
				}
				rootCmd.TableObj.Render()
			}

			return nil
		},
	}
}
