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
//
// --name scopes every subcommand to a single routing tree by metadata.name;
// omitted, each subcommand operates on every routing tree found.
func newAlertingRoutingTreeCommand() simplecobra.Commander {
	description := "Manage Alerting Notification Routing Trees"
	return &domain.SimpleCommand{
		NameP: "routingtree",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"routingtrees", "routing-tree", "routing-trees", "rt"}
			cmd.PersistentFlags().String("name", "", "filter by routing tree name; omit to operate on every routing tree")
		},
		CommandsList: []simplecobra.Commander{
			newListRoutingTreeCmd(),
			newDownloadRoutingTreeCmd(),
			newClearRoutingTreeCmd(),
			newUploadRoutingTreeCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

// getRoutingTreeNameFilter reads the --name flag shared by every routing tree subcommand.
func getRoutingTreeNameFilter(cmd *cobra.Command) string {
	name, _ := cmd.Flags().GetString("name")
	return name
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
			name := getRoutingTreeNameFilter(cd.CobraCommand)
			slog.Info("Listing routing trees for context",
				slog.String("name", name),
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			data, err := rootCmd.GrafanaSvc().ListRoutingTrees(name)
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
			name := getRoutingTreeNameFilter(cd.CobraCommand)
			slog.Info("Downloading routing trees for context",
				slog.String("name", name),
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			files, err := rootCmd.GrafanaSvc().DownloadRoutingTrees(name)
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

// newClearRoutingTreeCmd deletes every matching routing tree. When the server
// rejects the delete (the org's protected/default tree typically cannot be
// removed), it is reset to only the default receiver (grafana-default-email)
// instead of being skipped.
func newClearRoutingTreeCmd() simplecobra.Commander {
	description := "Clear all routing trees for the given Organization. The protected/default tree is reset to the default receiver instead of being deleted"
	return &domain.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			name := getRoutingTreeNameFilter(cd.CobraCommand)
			slog.Info("Clearing routing trees for context",
				slog.String("name", name),
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			cleared, err := rootCmd.GrafanaSvc().ClearRoutingTrees(name)
			if err != nil {
				slog.Error("unable to clear routing trees", slog.Any("err", err))
				return nil
			}
			if len(cleared) == 0 {
				slog.Info("No routing trees found")
				return nil
			}
			rootCmd.TableObj.AppendHeader(table.Row{"name"})
			for _, treeName := range cleared {
				rootCmd.TableObj.AppendRow(table.Row{treeName})
			}
			rootCmd.Render(cd.CobraCommand, cleared)
			return nil
		},
	}
}

// newUploadRoutingTreeCmd creates or updates routing trees from local storage,
// then removes any server-side routing tree that has no matching local file
// (falling back to resetting it when the server won't allow the delete), so
// the server ends up matching local storage exactly.
func newUploadRoutingTreeCmd() simplecobra.Commander {
	description := "Upload all routing trees for the given Organization, removing any server-side routing tree not present locally"
	return &domain.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			name := getRoutingTreeNameFilter(cd.CobraCommand)
			slog.Warn("Uploading will make Grafana match your local routing trees exactly: any routing tree on the " +
				"server that does not have a corresponding local file will be deleted (or reset to the default " +
				"receiver, if it cannot be deleted). Make sure your local copy is up to date before continuing.")
			slog.Info("Uploading routing trees for context",
				slog.String("name", name),
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			uploaded, err := rootCmd.GrafanaSvc().UploadRoutingTrees(name)
			if err != nil {
				slog.Error("unable to upload routing trees", slog.Any("err", err))
				return nil
			}
			if len(uploaded) == 0 {
				slog.Info("No routing trees found")
				return nil
			}
			rootCmd.TableObj.AppendHeader(table.Row{"name"})
			for _, treeName := range uploaded {
				rootCmd.TableObj.AppendRow(table.Row{treeName})
			}
			rootCmd.Render(cd.CobraCommand, uploaded)
			return nil
		},
	}
}
