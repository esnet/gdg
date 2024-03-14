package tools

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service/types"
	"github.com/spf13/cobra"
	"log/slog"
)

var lintStrictFlag bool
var lintVerboseFlag bool
var lintAutofixFlag bool

func newDashboardCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP:        "dashboards",
		Short:        "Utility for Grafana Dashboards",
		Long:         "Utility for Grafana Dashboards",
		CommandsList: []simplecobra.Commander{newDashboardLintCmd()},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"dash", "dashboard"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newDashboardLintCmd() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "lint",
		Short: "lint all or single dashboard",
		Long:  "lint all or a single dashboard",
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			dashboard := cmd
			dashboard.PersistentFlags().BoolVarP(&lintStrictFlag, "strict", "", false, "Strict Linting")
			dashboard.PersistentFlags().BoolVarP(&lintVerboseFlag, "verbose", "", false, "Verbose Linting")
			dashboard.PersistentFlags().BoolVarP(&lintAutofixFlag, "autofix", "", false, "AutoFix Linting (Beta)")
			dashboard.PersistentFlags().StringP("dashboard", "d", "", "filter by dashboard slug")
			dashboard.PersistentFlags().StringP("folder", "f", "", "filter by folderName")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("This is mainly provided as a convenience GDG, if you find yourself relying on this tool heavily, please have a look at: github.com/grafana/dashboard-linter/")
			dashboard, _ := cd.CobraCommand.Flags().GetString("dashboard")
			folder, _ := cd.CobraCommand.Flags().GetString("folder")
			filterReq := types.LintRequest{
				StrictFlag:    lintStrictFlag,
				VerboseFlag:   lintVerboseFlag,
				AutoFix:       lintAutofixFlag,
				DashboardSlug: dashboard,
				FolderName:    folder,
			}
			rootCmd.GrafanaSvc().LintDashboards(filterReq)
			return nil
		},
	}
}
