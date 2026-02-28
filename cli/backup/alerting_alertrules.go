package backup

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/bep/simplecobra"
	domain3 "github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config/config_domain"
	domain2 "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/ptr"
	"github.com/go-openapi/strfmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// getAlertRulesFilter constructs alert rule filters from command-line flags and returns both the parsed
// AlertRuleFilterParams and a corresponding Filter for use in alert rule operations.
func getAlertRulesFilter(cfg *config_domain.GDGAppConfiguration, grafanaService ports.GrafanaService, command *cobra.Command) (domain2.AlertRuleFilterParams, ports.Filter) {
	f := domain2.AlertRuleFilterParams{}
	f.Folder, _ = command.Flags().GetString("folder")
	f.Label, _ = command.Flags().GetStringArray("label")
	f.IgnoreWatchedFolders, _ = command.Flags().GetBool("ignore-watched-folders")

	return f, api.NewAlertRuleFilter(cfg, grafanaService, f)
}

// newAlertingRulesCommand creates and returns a Commander that manages Alerting Rules. It supports subcommands for
// listing, downloading, clearing, and uploading alert rules. It provides persistent flags for filtering by
// watched folders and labels.
func newAlertingRulesCommand() simplecobra.Commander {
	description := "Manage Alerting Rules"
	return &domain3.SimpleCommand{
		NameP: "rules",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain3.RootCommand) {
			cmd.Aliases = []string{"rule", "alert-rules", "alert-rule"}
			cmd.PersistentFlags().Bool("ignore-watched-folders", false, "Default to false, but if passed then will only operate on the list of folders listed in the configuration file")
			cmd.PersistentFlags().String("folder", "", "Add a folder filter")
			cmd.PersistentFlags().StringArray("label", []string{}, "Filter by label name value pair. (Additive behavior dashboard includes: label1 AND label2).  ex --label env=staging")
		},
		CommandsList: []simplecobra.Commander{
			newListAlertRulesCmd(),
			newDownloadAlertRulesCmd(),
			newClearAlertRulesCmd(),
			newUploadAlertRulesCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain3.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

// newUploadAlertRulesCmd creates and returns a Commander that uploads all alert rules for the configured organization.
// It supports folder and label filtering and is aliased as "u".
func newUploadAlertRulesCmd() simplecobra.Commander {
	description := "Upload all alert rules for the given Organization"
	return &domain3.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain3.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain3.RootCommand, args []string) error {
			filtersList, rulesFilter := getAlertRulesFilter(rootCmd.ConfigSvc(), rootCmd.GrafanaSvc(), cd.CobraCommand)
			rootCmd.TableObj.AppendHeader(table.Row{"uid"})
			slog.Info("Uploading all alert rules for context",
				slog.String("folder", filtersList.Folder),
				slog.Any("labelFilter", filtersList.Label),
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			err := rootCmd.GrafanaSvc().UploadAlertRules(rulesFilter)
			if err != nil {
				slog.Error("unable to upload Org's rule alerts", slog.Any("err", err))
				return nil
			}
			slog.Info("Rules have been successfully uploaded to grafana")
			return nil
		},
	}
}

// newClearAlertRulesCmd creates and returns a Commander that clears all alert rules for the configured organization.
// It supports optional filtering by folder and label, and renders the deleted rule titles as a table or JSON output.
// The command is aliased as "c".
func newClearAlertRulesCmd() simplecobra.Commander {
	description := "Clear all alert rules for the given Organization"
	return &domain3.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain3.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain3.RootCommand, args []string) error {
			filtersList, rulesFilter := getAlertRulesFilter(rootCmd.ConfigSvc(), rootCmd.GrafanaSvc(), cd.CobraCommand)
			slog.Info("Deleting all alert rules for context",
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("folder", filtersList.Folder),
				slog.Any("labelFilter", filtersList.Label),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			files, err := rootCmd.GrafanaSvc().ClearAlertRules(rulesFilter)
			if err != nil {
				slog.Error("unable to deleting Org's rule alerts", slog.Any("err", err))
				return nil
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

// newListAlertRulesCmd creates and returns a Commander that lists all alert rules for the given Organization.
// It supports filtering by folder and label, and renders results as a table or JSON output.
// The command is aliased as "l".
func newListAlertRulesCmd() simplecobra.Commander {
	description := "List all alert rules for the given Organization"
	return &domain3.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain3.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain3.RootCommand, args []string) error {
			filtersList, rulesFilter := getAlertRulesFilter(rootCmd.ConfigSvc(), rootCmd.GrafanaSvc(), cd.CobraCommand)
			rootCmd.TableObj.AppendHeader(table.Row{"name", "uid", "folder", "ruleGroup", "Labels", "For"})
			slog.Info("Listing alert rules for context",
				slog.String("folder", filtersList.Folder),
				slog.Any("labelFilter", filtersList.Label),
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			rules, err := rootCmd.GrafanaSvc().ListAlertRules(rulesFilter)
			if err != nil {
				slog.Error("unable to retrieve Orgs rule alerts", slog.Any("err", err))
				return nil
			}
			if len(rules) == 0 {
				slog.Info("No alert rules found")
			} else {
				for _, link := range rules {
					var labels string
					if len(link.Labels) > 0 {
						raw, jsonErr := json.Marshal(link.Labels)
						if jsonErr != nil {
							slog.Warn("unable to marshal labels", slog.Any("err", jsonErr))
						}
						labels = string(raw)
					}

					rootCmd.TableObj.AppendRow(table.Row{
						ptr.ValueOrDefault(link.Title, ""),
						link.UID,
						link.NestedPath,
						ptr.ValueOrDefault(link.RuleGroup, ""),
						labels,
						ptr.ValueOrDefault(link.For, strfmt.Duration(0)),
					})
				}
				rootCmd.Render(cd.CobraCommand, rules)
			}
			return nil
		},
	}
}

// newDownloadAlertRulesCmd creates and returns a Commander that downloads all alert rules for the given Organization.
// It applies any configured filters (folder, label) and renders the resulting list of downloaded rule files.
func newDownloadAlertRulesCmd() simplecobra.Commander {
	description := "Download all alert rules for the given Organization"
	return &domain3.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *domain3.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain3.RootCommand, args []string) error {
			filtersList, rulesFilter := getAlertRulesFilter(rootCmd.ConfigSvc(), rootCmd.GrafanaSvc(), cd.CobraCommand)
			rootCmd.TableObj.AppendHeader(table.Row{"alert-rule"})
			slog.Info("Downloading alert rules for context",
				slog.String("folder", filtersList.Folder),
				slog.Any("labelFilter", filtersList.Label),
				slog.String("Organization", GetOrganizationName(rootCmd.ConfigSvc())),
				slog.String("context", rootCmd.ConfigSvc().GetContext()))

			files, err := rootCmd.GrafanaSvc().DownloadAlertRules(rulesFilter)
			if err != nil {
				slog.Error("unable to retrieve Org's rule alerts", slog.Any("err", err))
				return nil
			}
			for _, link := range files {
				rootCmd.TableObj.AppendRow(table.Row{link})
			}
			rootCmd.Render(cd.CobraCommand, files)

			return nil
		},
	}
}
