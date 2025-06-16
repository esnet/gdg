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

func newAlertingTemplatesCommand() simplecobra.Commander {
	description := "Manage Alerting Templates"
	return &support.SimpleCommand{
		NameP: "templates",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"templates", "t"}
		},
		CommandsList: []simplecobra.Commander{
			newListAlertTemplatesCmd(),
			newDownloadAlertTemplatesCmd(),
			newClearAlertTemplatesCmd(),
			newUploadAlertTemplatesCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newUploadAlertTemplatesCmd() simplecobra.Commander {
	description := "Upload all alert templates for the given Organization"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"uid"})
			slog.Info("Uploading all alert templates for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			files, err := rootCmd.GrafanaSvc().UploadAlertTemplates()
			if err != nil {
				log.Fatal("unable to upload Orgs templates alerts", slog.Any("err", err))
			}
			rootCmd.TableObj.AppendHeader(table.Row{"title"})
			for _, link := range files {
				rootCmd.TableObj.AppendRow(table.Row{link})
			}
			rootCmd.Render(cd.CobraCommand, files)
			return nil
		},
	}
}

func newClearAlertTemplatesCmd() simplecobra.Commander {
	description := "Clear all alert templates for the given Organization"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Deleting all alert templates for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			files, err := rootCmd.GrafanaSvc().ClearAlertTemplates()
			if err != nil {
				log.Fatal("unable to deleting Orgs templates alerts", slog.Any("err", err))
			}
			rootCmd.TableObj.AppendHeader(table.Row{"name"})
			for _, link := range files {
				rootCmd.TableObj.AppendRow(table.Row{link})
			}
			rootCmd.Render(cd.CobraCommand, files)
			return nil
		},
	}
}

func newListAlertTemplatesCmd() simplecobra.Commander {
	description := "List all alert templates for the given Organization"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"name", "provenance", "template snippet", "version"})
			slog.Info("Listing alert templates for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			rules, err := rootCmd.GrafanaSvc().ListAlertTemplates()
			if err != nil {
				log.Fatal("unable to retrieve Orgs rule alerts", slog.Any("err", err))
			}
			if len(rules) == 0 {
				slog.Info("No alert rules found")
			} else {
				for _, link := range rules {
					if len(link.Template) > 50 {
						link.Template = link.Template[:50] + "..."
					}
					rootCmd.TableObj.AppendRow(table.Row{link.Name, link.Provenance, link.Template, link.Version})
				}
				rootCmd.Render(cd.CobraCommand, rules)
			}
			return nil
		},
	}
}

func newDownloadAlertTemplatesCmd() simplecobra.Commander {
	description := "Download all alert templates for the given Organization"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"uid", "folderUid", "ruleGroup", "Title", "provenance", "data"})
			slog.Info("Downloading alert templates for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			file, err := rootCmd.GrafanaSvc().DownloadAlertTemplates()
			if err != nil {
				log.Fatal("unable to retrieve Orgs templates alerts", slog.Any("err", err))
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
