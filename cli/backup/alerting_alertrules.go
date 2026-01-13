package backup

import (
	"context"
	"log"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/tools/ptr"
	"github.com/go-openapi/strfmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var ignoreAlertRuleFilters bool

func getAlertRulesFilter(cfg *domain.GDGAppConfiguration, grafanaService service.GrafanaService) filters.V2Filter {
	if ignoreAlertRuleFilters {
		return nil
	}
	return service.NewAlertRuleFilter(cfg, grafanaService)
}

func newAlertingRulesCommand() simplecobra.Commander {
	description := "Manage Alerting Rules"
	return &support.SimpleCommand{
		NameP: "rules",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"rule", "alert-rules", "alert-rule"}
			cmd.PersistentFlags().BoolVar(&ignoreAlertRuleFilters, "no-filters", false, "Default to false, but if passed then will only operate on the list of folders listed in the configuration file")
		},
		CommandsList: []simplecobra.Commander{
			newListAlertRulesCmd(),
			newDownloadAlertRulesCmd(),
			newClearAlertRulesCmd(),
			newUploadAlertRulesCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newUploadAlertRulesCmd() simplecobra.Commander {
	description := "Upload all alert rules for the given Organization"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"uid"})
			slog.Info("Uploading all alert rules for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			err := rootCmd.GrafanaSvc().UploadAlertRules(getAlertRulesFilter(rootCmd.ConfigSvc(), rootCmd.GrafanaSvc()))
			if err != nil {
				log.Fatal("unable to upload Orgs rule alerts", slog.Any("err", err))
			}
			slog.Info("Rules have been successfully uploaded to grafana")
			return nil
		},
	}
}

func newClearAlertRulesCmd() simplecobra.Commander {
	description := "Clear all alert rules for the given Organization"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Deleting all alert rules for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			files, err := rootCmd.GrafanaSvc().ClearAlertRules(getAlertRulesFilter(rootCmd.ConfigSvc(), rootCmd.GrafanaSvc()))
			if err != nil {
				log.Fatal("unable to deleting Orgs rule alerts", slog.Any("err", err))
			}
			if len(files) > 0 {
				rootCmd.TableObj.AppendHeader(table.Row{"title"})
				for _, link := range files {
					rootCmd.TableObj.AppendRow(table.Row{link})
				}
				rootCmd.Render(cd.CobraCommand, files)
			} else {
				slog.Info("No Alerting rules were found")
			}

			return nil
		},
	}
}

func newListAlertRulesCmd() simplecobra.Commander {
	description := "List all alert rules for the given Organization"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"name", "uid", "folder", "ruleGroup", "For"})
			slog.Info("Listing alert rules for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			rules, err := rootCmd.GrafanaSvc().ListAlertRules(getAlertRulesFilter(rootCmd.ConfigSvc(), rootCmd.GrafanaSvc()))
			if err != nil {
				log.Fatal("unable to retrieve Orgs rule alerts", slog.Any("err", err))
			}
			if len(rules) == 0 {
				slog.Info("No alert rules found")
			} else {
				for _, link := range rules {
					rootCmd.TableObj.AppendRow(table.Row{
						ptr.ValueOrDefault(link.Title, ""),
						link.UID,
						link.NestedPath,
						ptr.ValueOrDefault(link.RuleGroup, ""),
						ptr.ValueOrDefault(link.For, strfmt.Duration(0)),
					})
				}
				rootCmd.Render(cd.CobraCommand, rules)
			}
			return nil
		},
	}
}

func newDownloadAlertRulesCmd() simplecobra.Commander {
	description := "Download all alert rules for the given Organization"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"alert-rule"})
			slog.Info("Downloading alert rules for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			files, err := rootCmd.GrafanaSvc().DownloadAlertRules(getAlertRulesFilter(rootCmd.ConfigSvc(), rootCmd.GrafanaSvc()))
			if err != nil {
				log.Fatal("unable to retrieve Orgs rule alerts", slog.Any("err", err))
			}
			if err != nil {
				slog.Error("unable to download alert rules")
			} else {
				for _, link := range files {
					rootCmd.TableObj.AppendRow(table.Row{link})
				}
				rootCmd.Render(cd.CobraCommand, files)
			}
			return nil
		},
	}
}
