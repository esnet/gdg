package backup

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/internal/tools"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

var useFolderFilters bool

func getFolderFilter(cfg *domain.GDGAppConfiguration) filters.V2Filter {
	if !useFolderFilters {
		return nil
	}
	return service.NewFolderFilter(cfg)
}

func newFolderCommand() simplecobra.Commander {
	description := "Manage folder entities"
	return &support.SimpleCommand{
		NameP: "folders",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"fld", "folder", "f"}
			cmd.PersistentFlags().BoolVar(&useFolderFilters, "use-filters", false, "Default to false, but if passed then will only operate on the list of folders listed in the configuration file")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		CommandsList: []simplecobra.Commander{
			newFolderPermissionCommand(),
			newFolderListCmd(),
			newFolderClearCmd(),
			newFolderDownloadCmd(),
			newFolderUploadCmd(),
		},
	}
}

func newFolderClearCmd() simplecobra.Commander {
	description := "delete Folders from grafana"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c", "delete"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Deleting all Folders for context", "context", rootCmd.ConfigSvc().GetContext())
			if !skipConfirmAction {
				tools.GetUserConfirmation(fmt.Sprintf("WARNING: this will delete all folders in the monitored folders list: '%s' "+
					"(or all folders in your grafana instance if ignore_dashboard_filters is set to true).  Do you wish to "+
					"continue (y/n) ", strings.Join(rootCmd.ConfigSvc().GetDefaultGrafanaConfig().GetMonitoredFolders(false), ", "),
				), "", true)
			}
			rootCmd.TableObj.AppendHeader(table.Row{"title"})

			folders := rootCmd.GrafanaSvc().DeleteAllFolders(getFolderFilter(rootCmd.ConfigSvc()))
			if len(folders) == 0 {
				slog.Info("No Folders found")
			} else {
				for _, folder := range folders {
					rootCmd.TableObj.AppendRow(table.Row{folder})
				}
				rootCmd.Render(cd.CobraCommand, folders)
			}
			return nil
		},
	}
}

func newFolderListCmd() simplecobra.Commander {
	description := "List Folders"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Listing Folders for context", "context", rootCmd.ConfigSvc().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"uid", "title", "nestedPath"})
			folders := rootCmd.GrafanaSvc().ListFolders(getFolderFilter(rootCmd.ConfigSvc()))

			if len(folders) == 0 {
				slog.Info("No folders found")
			} else {
				for _, folder := range folders {
					row := table.Row{folder.UID, folder.Title, folder.NestedPath}
					rootCmd.TableObj.AppendRow(row)
				}
				rootCmd.Render(cd.CobraCommand, folders)
			}
			return nil
		},
	}
}

func newFolderDownloadCmd() simplecobra.Commander {
	description := "Download Folders from grafana"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Listing Folders for context", "context", rootCmd.ConfigSvc().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			folders := rootCmd.GrafanaSvc().DownloadFolders(getFolderFilter(rootCmd.ConfigSvc()))
			if len(folders) == 0 {
				slog.Info("No folders found")
			} else {
				for _, folder := range folders {
					rootCmd.TableObj.AppendRow(table.Row{folder})
				}
				rootCmd.Render(cd.CobraCommand, folders)
			}
			return nil
		},
	}
}

func newFolderUploadCmd() simplecobra.Commander {
	description := "upload Folders to grafana"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Uploading Folders for context", "context", rootCmd.ConfigSvc().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"file"})
			folders := rootCmd.GrafanaSvc().UploadFolders(getFolderFilter(rootCmd.ConfigSvc()))
			if len(folders) == 0 {
				slog.Info("No folders found")
			} else {
				for _, folder := range folders {
					rootCmd.TableObj.AppendRow(table.Row{folder})
				}
				rootCmd.Render(cd.CobraCommand, folders)
			}
			return nil
		},
	}
}
