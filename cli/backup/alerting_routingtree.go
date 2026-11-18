package backup

import (
	"context"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// newAlertingRoutingTreeCommand manages Notification Policy routing trees via
// the App Platform (K8s-style) notifications.alerting.grafana.app API. It
// replaces the older "policy" command: the legacy /api/v1/provisioning/policies
// endpoint only exposes a single flat policy tree per org and silently drops
// nested sub-routes, whereas this endpoint supports multiple named routing
// trees and returns the full recursive route structure.
func newAlertingRoutingTreeCommand() simplecobra.Commander {
	description := "Manage Alerting Notification Routing Trees"
	return &domain.SimpleCommand{
		NameP: "routingtree",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"routingtrees", "routing-tree", "routing-trees", "rt"}
		},
		CommandsList: []simplecobra.Commander{
			newListRoutingTreeCmd(),
			newDownloadRoutingTreeCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newListRoutingTreeCmd() simplecobra.Commander {
	description := "List all routing trees for the given Organization"
	return &domain.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			slog.Info("Listing routing trees for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			data, err := rootCmd.GrafanaSvc().ListRoutingTrees()
			if err != nil {
				slog.Error("unable to list routing trees", slog.Any("err", err))
				return nil
			}
			if len(data) == 0 {
				slog.Info("No routing trees found")
				return nil
			}
			rootCmd.TableObj.AppendHeader(table.Row{"name", "default receiver", "routes", "gdg api version"})
			for _, tree := range data {
				rootCmd.TableObj.AppendRow(table.Row{tree.Name(), tree.DefaultReceiver(), tree.RouteCount(), tree.GdgApiVersion})
			}
			rootCmd.Render(cd.CobraCommand, data)
			return nil
		},
	}
}

func newDownloadRoutingTreeCmd() simplecobra.Commander {
	description := "Download all routing trees for the given Organization"
	return &domain.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			slog.Info("Downloading routing trees for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			files, err := rootCmd.GrafanaSvc().DownloadRoutingTrees()
			if err != nil {
				slog.Error("unable to download routing trees", slog.Any("err", err))
				return nil
			}
			if len(files) == 0 {
				slog.Info("No routing trees found")
				return nil
			}
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			for _, file := range files {
				rootCmd.TableObj.AppendRow(table.Row{file})
			}
			rootCmd.Render(cd.CobraCommand, files)
			return nil
		},
	}
}
