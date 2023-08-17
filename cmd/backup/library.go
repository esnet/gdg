package backup

import (
	"encoding/json"
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var libraryElements = &cobra.Command{
	Use:     "libraryelements",
	Aliases: []string{"lib", "library"},
	Short:   "Manage Library Elements",
	Long:    `Manage Library Elements.`,
}

var clearLibrary = &cobra.Command{
	Use:     "clear",
	Aliases: []string{"c"},
	Short:   "delete all Library elements from grafana",
	Long:    `delete all library elements from grafana`,
	Run: func(command *cobra.Command, args []string) {
		//filter := getLibraryGlobalFlags(cmd)
		deletedLibrarys := cmd.GetGrafanaSvc().DeleteAllLibraryElements(nil)
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range deletedLibrarys {
			cmd.TableObj.AppendRow(table.Row{"library", file})
		}
		if len(deletedLibrarys) == 0 {
			log.Info("No library were found.  0 librarys removed")

		} else {
			log.Infof("%d library were deleted", len(deletedLibrarys))
			cmd.TableObj.Render()
		}

	},
}

var uploadLibrary = &cobra.Command{
	Use:     "upload",
	Short:   "upload all library to grafana",
	Long:    `upload all library to grafana`,
	Aliases: []string{"u"},
	Run: func(command *cobra.Command, args []string) {
		log.Info("exporting lib elements")
		libraryFilter := filters.NewBaseFilter()
		elements := cmd.GetGrafanaSvc().UploadLibraryElements(libraryFilter)
		cmd.TableObj.AppendHeader(table.Row{"Name"})
		if len(elements) > 0 {
			for _, link := range elements {
				cmd.TableObj.AppendRow(table.Row{link})
			}
			cmd.TableObj.Render()
		} else {
			log.Info("No library found")
		}
	},
}

var downloadLibary = &cobra.Command{
	Use:     "download",
	Short:   "Download all library from grafana",
	Long:    `Download all library from grafana to local file system`,
	Aliases: []string{"d"},
	Run: func(command *cobra.Command, args []string) {
		log.Info("exporting lib elements")
		savedFiles := cmd.GetGrafanaSvc().DownloadLibraryElements(nil)
		log.Infof("Importing library for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			cmd.TableObj.AppendRow(table.Row{"library", file})
		}
		cmd.TableObj.Render()

	},
}

var listLibraries = &cobra.Command{
	Use:   "list",
	Short: "List all library",
	Long:  `List all library`,
	Run: func(command *cobra.Command, args []string) {
		cmd.TableObj.AppendHeader(table.Row{"id", "UID", "Folder", "Name", "Type"})

		elements := cmd.GetGrafanaSvc().ListLibraryElements(nil)
		log.Infof("Number of elements is: %d", len(elements))

		log.Infof("Listing library for context: '%s'", config.Config().AppConfig.GetContext())
		for _, link := range elements {
			cmd.TableObj.AppendRow(table.Row{link.ID, link.UID, link.Meta.FolderName, link.Name, link.Type})

		}
		if len(elements) > 0 {
			cmd.TableObj.Render()
		} else {
			log.Info("No library found")
		}

	},
}

var listLibraryConnections = &cobra.Command{
	Use:   "list-connections",
	Short: "List all library Connection given a valid library Connection UID",
	Long:  `List all library Connection for a given library connection UID`,
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(command *cobra.Command, args []string) {
		cmd.TableObj.AppendHeader(table.Row{"id", "UID", "Slug", "Title", "Folder"})

		libElmentUid := args[0]
		elements := cmd.GetGrafanaSvc().ListLibraryElementsConnections(nil, libElmentUid)
		log.Infof("Listing library connections for context: '%s'", config.Config().AppConfig.GetContext())
		for _, link := range elements {
			dash := link.Dashboard.(map[string]interface{})
			cmd.TableObj.AppendRow(table.Row{dash["id"].(json.Number), dash["uid"].(string), link.Meta.Slug, dash["title"].(string), link.Meta.FolderTitle})
		}
		if len(elements) > 0 {
			cmd.TableObj.Render()
		} else {
			log.Info("No library found")
		}
		/*


		 */

	},
}

func init() {
	backupCmd.AddCommand(libraryElements)
	libraryElements.AddCommand(downloadLibary)
	libraryElements.AddCommand(listLibraries)
	libraryElements.AddCommand(clearLibrary)
	libraryElements.AddCommand(uploadLibrary)
	libraryElements.AddCommand(listLibraryConnections)
}
