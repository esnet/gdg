package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/pretty"

	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/yalp/jsonpath"
)



//ListDashboards List all dashboards optionally filtered by folder name. If folderFilters
// is blank, defaults to the configured Monitored folders
func (s *DashNGoImpl) ListDashboards(filters Filter) []sdk.FoundBoard {
	ctx := context.Background()
	orgs, err := s.client.GetAllOrgs(ctx)
	if err != nil {
		log.Warnf("Error getting organizations: %s", err.Error())
		orgs = make([]sdk.Org, 0)
	}
	if s.grafanaConf.Organization != "" {
		var ID uint
		for _, org := range orgs {
			log.Errorf("%d %s\n", org.ID, org.Name)
			if org.Name == s.grafanaConf.Organization {
				ID = org.ID
			}
		}
		if ID > 0 {
			status, err := s.client.SwitchActualUserContext(ctx, ID)
			if err != nil {
				log.Fatalf("%s for %v\n", err, status)
			}
		}
	}
	var boardsList []sdk.FoundBoard = make([]sdk.FoundBoard, 0)
	boardLinks, err := s.client.Search(ctx, sdk.SearchType(sdk.SearchTypeDashboard))
	if err != nil {
		log.Fatal("Failed to retrieve dashboards", err)
	}
	//Fallback on defaults
	if filters == nil {
		filters = &DashboardFilter{}
	}

	folderFilters := filters.GetFolders()
	var validFolder bool = false
	var validUid bool = false
	for _, link := range boardLinks {
		if apphelpers.GetCtxDefaultGrafanaConfig().IgnoreFilters {
			validFolder = true
		} else if funk.Contains(folderFilters, link.FolderTitle) {
			validFolder = true
		} else if funk.Contains(folderFilters, DefaultFolderName) && link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
			validFolder = true
		}
		if !validFolder {
			continue
		}
		updateSlug(&link)
		if filters.GetFilter("DashFilter") != "" {
			if link.Slug == filters.GetFilter("DashFilter") {
				validUid = true
			} else {
				validUid = false
			}
		} else {
			validUid = true
		}
		if link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
		}

		if validFolder && validUid {
			boardsList = append(boardsList, link)
		}

		validFolder, validUid = false, false

	}

	return boardsList

}

//ImportDashboards saves all dashboards matching query to configured location
func (s *DashNGoImpl) ImportDashboards(filter Filter) []string {
	var (
		boardLinks []sdk.FoundBoard
		rawBoard   []byte
		meta       sdk.BoardProperties
		err        error
	)
	ctx := context.Background()

	boardLinks = s.ListDashboards(filter)
	var boards []string = make([]string, 0)
	for _, link := range boardLinks {
		if rawBoard, meta, err = s.client.GetRawDashboardByUID(ctx, link.UID); err != nil {
			log.Errorf("%s for %s\n", err, link.URI)
			continue
		}
		fileName := fmt.Sprintf("%s/%s.json", buildDashboardPath(s.configRef, link.FolderTitle), meta.Slug)
		if err = ioutil.WriteFile(fileName, pretty.Pretty(rawBoard), os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err, meta.Slug)
		} else {
			boards = append(boards, fileName)
		}

	}
	return boards
}

//ExportDashboards finds all the dashboards in the configured location and exports them to grafana.
// if the folder doesn't exist, it'll be created.
func (s *DashNGoImpl) ExportDashboards(filters Filter) {
	path := getResourcePath(s.configRef, "dashboard")
	filesInDir := findAllFiles(path)
	ctx := context.Background()
	var rawBoard []byte
	folderMap := getFolderNameIDMap(s.client, ctx)
	var err error
	var folderName string = ""
	var folderId int

	//Fallback on defaults
	if filters == nil {
		filters = &DashboardFilter{}
	}

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
			elements := strings.Split(file, string(os.PathSeparator))
			if len(elements) >= 2 {
				folderName = elements[len(elements)-2]
			}
			if folderName == "" || folderName == DefaultFolderName {
				folderId = sdk.DefaultFolderId
				folderName = DefaultFolderName
			}
			if !funk.Contains(filters.GetFolders(), folderName) && !apphelpers.GetCtxDefaultGrafanaConfig().IgnoreFilters {
				log.Debugf("Skipping file %s, doesn't match any valid folders", file)
				continue
			}
			validateMap := map[string]string{FolderFilter: folderName, DashFilter: baseFile}

			if folderName == DefaultFolderName {
				folderId = sdk.DefaultFolderId
			} else {
				if val, ok := folderMap[folderName]; ok {
					folderId = val
				} else {
					if filters.Validate(validateMap) {
						folder := sdk.Folder{Title: folderName}
						folder, err = s.client.CreateFolder(ctx, folder)
						if err != nil {
							panic(err)
						}
						folderMap[folderName] = folder.ID
						folderId = folder.ID
					}
				}
			}

			//If folder OR slug is filtered, then skip if it doesn't match
			if !filters.Validate(validateMap) {
				continue
			}

			title, err := jsonpath.Read(board, "$.title")
			if err != nil {
				log.Warn("Could not get dashboard title")
			}

			rawTitle := fmt.Sprintf("%v", title)
			slugName := GetSlug(rawTitle)
			if _, err = s.client.DeleteDashboard(ctx, slugName); err != nil {
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

			_, err = s.client.SetRawDashboardWithParam(ctx, request)
			if err != nil {
				log.Printf("error on Exporting dashboard %s", rawTitle)
				continue
			}
		}
	}
}

//DeleteAllDashboards clears all current dashboards being monitored.  Any folder not white listed
// will not be affected
func (s *DashNGoImpl) DeleteAllDashboards(filter Filter) []string {
	ctx := context.Background()
	var dashboards []string = make([]string, 0)

	items := s.ListDashboards(filter)
	for _, item := range items {
		if filter.Validate(map[string]string{FolderFilter: item.FolderTitle, DashFilter: item.Slug}) {
			_, err := s.client.DeleteDashboardByUID(ctx, item.UID)
			if err == nil {
				dashboards = append(dashboards, item.Title)
			}
		}
	}
	return dashboards

}
