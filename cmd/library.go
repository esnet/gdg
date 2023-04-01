package cmd

import (
	"encoding/json"
	"github.com/esnet/gdg/api/filters"
	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
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
	Use:   "clear",
	Short: "delete all Library elements",
	Long:  `clear all library elements`,
	Run: func(cmd *cobra.Command, args []string) {
		//filter := getLibraryGlobalFlags(cmd)
		deletedLibrarys := client.DeleteAllLibraryElements(nil)
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range deletedLibrarys {
			tableObj.AppendRow(table.Row{"library", file})
		}
		if len(deletedLibrarys) == 0 {
			log.Info("No library were found.  0 librarys removed")

		} else {
			log.Infof("%d library were deleted", len(deletedLibrarys))
			tableObj.Render()
		}

	},
}

var exportLibrary = &cobra.Command{
	Use:   "export",
	Short: "export all library",
	Long:  `export all library`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("exporting lib elements")
		libraryFilter := filters.NewBaseFilter()
		elements := client.ExportLibraryElements(libraryFilter)
		tableObj.AppendHeader(table.Row{"Name"})
		if len(elements) > 0 {
			for _, link := range elements {
				tableObj.AppendRow(table.Row{link})
			}
			tableObj.Render()
		} else {
			log.Info("No library found")
		}
	},
}

var importLibrary = &cobra.Command{
	Use:   "import",
	Short: "Import all library",
	Long:  `Import all library from grafana to local file system`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("exporting lib elements")
		//filter := parseDashboardGlobalFlags(cmd)
		savedFiles := client.ImportLibraryElements(nil)
		log.Infof("Importing library for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"type", "filename"})
		for _, file := range savedFiles {
			tableObj.AppendRow(table.Row{"library", file})
		}
		tableObj.Render()

	},
}

var listLibraries = &cobra.Command{
	Use:   "list",
	Short: "List all library",
	Long:  `List all library`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "UID", "Folder", "Name", "Type"})

		elements := client.ListLibraryElements(nil)
		log.Infof("Number of elements is: %d", len(elements))

		log.Infof("Listing library for context: '%s'", apphelpers.GetContext())
		for _, link := range elements {
			tableObj.AppendRow(table.Row{link.ID, link.UID, link.Meta.FolderName, link.Name, link.Type})

		}
		if len(elements) > 0 {
			tableObj.Render()
		} else {
			log.Info("No library found")
		}

	},
}

var listLibraryConnections = &cobra.Command{
	Use:   "list-connections",
	Short: "List all library Connection given a valid library Connection UID",
	Long:  `List all library Connection`,
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"id", "UID", "Slug", "Title", "Folder"})

		libElmentUid := args[0]
		elements := client.ListLibraryElementsConnections(nil, libElmentUid)
		log.Infof("Listing library for context: '%s'", apphelpers.GetContext())
		for _, link := range elements {
			dash := link.Dashboard.(map[string]interface{})
			tableObj.AppendRow(table.Row{dash["id"].(json.Number), dash["uid"].(string), link.Meta.Slug, dash["title"].(string), link.Meta.FolderTitle})
		}
		if len(elements) > 0 {
			tableObj.Render()
		} else {
			log.Info("No library found")
		}
		/*


		 */

	},
}

func init() {
	rootCmd.AddCommand(libraryElements)
	libraryElements.AddCommand(importLibrary)
	libraryElements.AddCommand(listLibraries)
	libraryElements.AddCommand(clearLibrary)
	libraryElements.AddCommand(exportLibrary)
	libraryElements.AddCommand(listLibraryConnections)
}
