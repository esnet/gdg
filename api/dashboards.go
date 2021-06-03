package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"github.com/tidwall/pretty"

	"github.com/netsage-project/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/yalp/jsonpath"
)

//ListDashboards: List all dashboards optionally filtered by folder name. If folderFilters
// is blank, defaults to the configured Monitored folders
func ListDashboards(client *sdk.Client, filters *DashboardFilter) []sdk.FoundBoard {
	ctx := context.Background()
	var boardsList []sdk.FoundBoard = make([]sdk.FoundBoard, 0)
	boardLinks, err := client.SearchDashboards(ctx, "", false)
	if err != nil {
		panic(err)
	}
	//Fallback on defaults
	if filters == nil {
		filters = &DashboardFilter{}
	}

	folderFilters := filters.GetFolders()
	var validFolder bool = false
	var validUid bool = false
	for _, link := range boardLinks {
		if funk.Contains(folderFilters, link.FolderTitle) {
			validFolder = true
		} else if funk.Contains(folderFilters, DefaultFolderName) && link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
			validFolder = true
		}
		if !validFolder {
			continue
		}
		updateSlug(&link)
		if filters.DashFilter != "" {
			if link.Slug == filters.DashFilter {
				validUid = true
			} else {
				validUid = false
			}
		} else {
			validUid = true
		}

		if validFolder && validUid {
			boardsList = append(boardsList, link)
		}

		validFolder, validUid = false, false

	}

	return boardsList

}

//ImportDashboards saves all dashboards matching query to configured location
func ImportDashboards(client *sdk.Client, filter DashboardFilter, conf *viper.Viper) []string {
	var (
		boardLinks []sdk.FoundBoard
		rawBoard   []byte
		meta       sdk.BoardProperties
		err        error
	)
	ctx := context.Background()

	boardLinks = ListDashboards(client, &filter)
	var boards []string = make([]string, 0)
	for _, link := range boardLinks {
		if rawBoard, meta, err = client.GetRawDashboardByUID(ctx, link.UID); err != nil {
			fmt.Fprintf(os.Stderr, "%s for %s\n", err, link.URI)
			continue
		}
		fileName := fmt.Sprintf("%s/%s.json", buildDashboardPath(conf, link.FolderTitle), meta.Slug)
		if err = ioutil.WriteFile(fileName, pretty.Pretty(rawBoard), os.FileMode(int(0666))); err != nil {
			fmt.Fprintf(os.Stderr, "%s for %s\n", err, meta.Slug)
		} else {
			boards = append(boards, fileName)
		}

	}
	return boards
}

//getFolderNameIDMap helper function to build a mapping for name to folderID
func getFolderNameIDMap(client *sdk.Client, ctx context.Context) map[string]int {

	folders, _ := client.GetAllFolders(ctx)
	var folderMap map[string]int = make(map[string]int, 0)
	for _, folder := range folders {
		folderMap[folder.Title] = folder.ID
	}
	return folderMap
}

//ExportDashboards finds all the dashboards in the configured location and exports them to grafana.
// if the folde doesn't exist, it'll be created.
func ExportDashboards(client *sdk.Client, filters DashboardFilter, conf *viper.Viper) {
	filesInDir := findAllFiles(getResourcePath(conf, "dashboard"))
	ctx := context.Background()
	var rawBoard []byte
	folderMap := getFolderNameIDMap(client, ctx)
	var err error
	var folderName string = ""
	var folderId int

	for _, file := range filesInDir {
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")
		if strings.HasSuffix(file, ".json") {
			if rawBoard, err = ioutil.ReadFile(file); err != nil {
				log.Println(err)
				continue
			}
			var board = make(map[string]interface{})
			if err = json.Unmarshal(rawBoard, &board); err != nil {
				log.Println(err)
				log.Printf("Failed to unmarshall file: %s", file)
				continue
			}

			elements := strings.Split(file, "/")
			if len(elements) >= 2 {
				folderName = elements[len(elements)-2]
			}
			if folderName == "" || folderName == DefaultFolderName {
				folderId = sdk.DefaultFolderId
				folderName = DefaultFolderName
			}

			if folderName == DefaultFolderName {
				folderId = sdk.DefaultFolderId
			} else {
				if val, ok := folderMap[folderName]; ok {
					folderId = val
				} else {
					createFolder := filters.ValidateFolder(folderName)
					validUid := filters.ValidateDashboard(baseFile)

					if createFolder && validUid {
						folder := sdk.Folder{Title: folderName}
						folder, err = client.CreateFolder(ctx, folder)
						if err != nil {
							panic(err)
						}
						folderMap[folderName] = folder.ID
						folderId = folder.ID
					}
				}
			}

			//If folder OR slug is filtered, then skip if it doesn't match
			if !filters.Validate(folderName, baseFile) {
				continue
			}

			title, err := jsonpath.Read(board, "$.title")

			rawTitle := fmt.Sprintf("%v", title)
			slugName := GetSlug(rawTitle)
			if _, err = client.DeleteDashboard(ctx, slugName); err != nil {
				log.Println(err)
				continue
			}
			defaultParams := sdk.SetDashboardParams{
				Overwrite: true,
				FolderID:  folderId,
			}
			request := sdk.RawBoardRequest{
				Parameters: defaultParams,
				Dashboard:  rawBoard,
			}

			_, err = client.SetRawDashboardWithParam(ctx, request)
			if err != nil {
				log.Printf("error on Exporting dashboard %s", rawTitle)
				continue
			}
		}
	}
}

//DeleteAllDashboards clears all current dashboards being monitored.  Any folder not white listed
// will not be affected
func DeleteAllDashboards(client *sdk.Client, filter DashboardFilter) []string {
	ctx := context.Background()
	var dashboards []string = make([]string, 0)

	items := ListDashboards(client, &filter)
	for _, item := range items {
		if filter.Validate(item.FolderTitle, item.Slug) {
			_, err := client.DeleteDashboardByUID(ctx, item.UID)
			if err == nil {
				dashboards = append(dashboards, item.Title)
			}
		}
	}
	return dashboards

}

var quoteRegex *regexp.Regexp

func init() {
	quoteRegex, _ = regexp.Compile("['\"]+")
}
