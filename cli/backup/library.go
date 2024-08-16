package backup

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
)

func newLibraryElementsCommand() simplecobra.Commander {
	description := "Manage Library Elements"
	return &support.SimpleCommand{
		NameP: "libraryelements",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"lib", "library"}
		},
		CommandsList: []simplecobra.Commander{
			newLibraryElementsListCmd(),
			newLibraryElementsClearCmd(),
			newLibraryElementsDownloadCmd(),
			newLibraryElementsUploadCmd(),
			newLibraryElementsListConnectionsCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newLibraryElementsClearCmd() simplecobra.Commander {
	description := "delete all Library elements from grafana"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			//filter := getLibraryGlobalFlags(cli)
			deletedLibrarys := rootCmd.GrafanaSvc().DeleteAllLibraryElements(nil)
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			for _, file := range deletedLibrarys {
				rootCmd.TableObj.AppendRow(table.Row{"library", file})
			}
			if len(deletedLibrarys) == 0 {
				slog.Info("No library were found.  0 libraries removed")

			} else {
				slog.Info("libraries were deleted", "count", len(deletedLibrarys))
				rootCmd.Render(cd.CobraCommand, deletedLibrarys)
			}
			return nil
		},
	}
}
func newLibraryElementsListCmd() simplecobra.Commander {
	description := "List all library Elements"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"id", "UID", "Folder", "Name", "Type"})

			elements := rootCmd.GrafanaSvc().ListLibraryElements(nil)

			slog.Info("Listing library for context", "context", config.Config().GetGDGConfig().GetContext())
			for _, link := range elements {
				rootCmd.TableObj.AppendRow(table.Row{link.ID, link.UID, link.Meta.FolderName, link.Name, link.Type})

			}
			if len(elements) > 0 {
				rootCmd.Render(cd.CobraCommand, elements)
			} else {
				slog.Info("No library found")
			}

			return nil
		},
	}
}
func newLibraryElementsDownloadCmd() simplecobra.Commander {
	description := "Download all library from grafana to local file system"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Downloading library for context", "context", config.Config().GetGDGConfig().GetContext())
			savedFiles := rootCmd.GrafanaSvc().DownloadLibraryElements(nil)
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			for _, file := range savedFiles {
				rootCmd.TableObj.AppendRow(table.Row{"library", file})
			}
			rootCmd.Render(cd.CobraCommand, savedFiles)
			return nil
		},
	}
}
func newLibraryElementsUploadCmd() simplecobra.Commander {
	description := "upload all library to grafana"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("exporting lib elements")
			libraryFilter := filters.NewBaseFilter()
			elements := rootCmd.GrafanaSvc().UploadLibraryElements(libraryFilter)
			rootCmd.TableObj.AppendHeader(table.Row{"Name"})
			if len(elements) > 0 {
				for _, link := range elements {
					rootCmd.TableObj.AppendRow(table.Row{link})
				}
				rootCmd.Render(cd.CobraCommand, elements)
			} else {
				slog.Info("No library found")
			}
			return nil
		},
	}
}

func newLibraryElementsListConnectionsCmd() simplecobra.Commander {
	description := "List all library Connection given a valid library Connection UID"
	return &support.SimpleCommand{
		NameP: "list-connections",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) != 1 {
				log.Fatal("Wrong number of arguments, requires library element UUID")
			}
			rootCmd.TableObj.AppendHeader(table.Row{"id", "UID", "Slug", "Title", "Folder"})

			libElmentUid := args[0]
			elements := rootCmd.GrafanaSvc().ListLibraryElementsConnections(nil, libElmentUid)
			slog.Info("Listing library connections for context", "context", config.Config().GetGDGConfig().GetContext())
			for _, link := range elements {
				dash := link.Dashboard.(map[string]interface{})
				rootCmd.TableObj.AppendRow(table.Row{dash["id"], dash["uid"], link.Meta.Slug, dash["title"], link.Meta.FolderTitle})
			}
			if len(elements) > 0 {
				rootCmd.Render(cd.CobraCommand, elements)
			} else {
				slog.Info("No library found")
			}
			return nil
		},
	}
}
