package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/config"

	"github.com/esnet/gdg/apphelpers"

	"github.com/tidwall/pretty"

	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"github.com/yalp/jsonpath"
)

// ListDashboards List all dashboards optionally filtered by folder name. If folderFilters
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

	// Fallback on defaults
	if filters == nil {
		filters = &DashboardFilter{}
	}

	var boardsList []sdk.FoundBoard = make([]sdk.FoundBoard, 0)
	var boardLinks []sdk.FoundBoard = make([]sdk.FoundBoard, 0)

	var page uint = 1
	var limit uint = 5000 // Upper bound of Grafana API call

	var tagsParams []sdk.SearchParam = []sdk.SearchParam{}
	if !apphelpers.GetCtxDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
		for _, t := range filters.GetTags() {
			tagsParams = append(tagsParams, sdk.SearchTag(t))
		}
	}
	var searchParams []sdk.SearchParam

	for {
		searchParams = append(tagsParams, sdk.SearchType(sdk.SearchTypeDashboard), sdk.SearchLimit(limit), sdk.SearchPage(page))
		pageBoardLinks, err := s.client.Search(ctx, searchParams...)
		if err != nil {
			log.Fatal("Failed to retrieve dashboards", err)
		}
		boardLinks = append(boardLinks, pageBoardLinks...)
		if len(pageBoardLinks) < int(limit) {
			break
		}
		page += 1
	}

	folderFilters := filters.GetFolders()
	var validFolder bool
	var validUid bool
	for _, link := range boardLinks {
		validFolder = false
		if apphelpers.GetCtxDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
			validFolder = true
		} else if funk.ContainsString(folderFilters, link.FolderTitle) {
			validFolder = true
		} else if funk.ContainsString(folderFilters, DefaultFolderName) && link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
			validFolder = true
		}
		if !validFolder {
			continue
		}
		link.Slug = updateSlug(link)
		validUid = filters.GetFilter("DashFilter") == "" || link.Slug == filters.GetFilter("DashFilter")
		if link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
		}

		if validFolder && validUid {
			boardsList = append(boardsList, link)
		}
	}

	return boardsList

}

// ImportDashboards saves all dashboards matching query to configured location
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

		fileName := fmt.Sprintf("%s/%s.json", buildResourceFolder(link.FolderTitle, config.DashboardResource), meta.Slug)
		if err = s.storage.WriteFile(fileName, pretty.Pretty(rawBoard), os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err, meta.Slug)
		} else {
			boards = append(boards, fileName)
		}

	}
	return boards
}

// ExportDashboards finds all the dashboards in the configured location and exports them to grafana.
// if the folder doesn't exist, it'll be created.
func (s *DashNGoImpl) ExportDashboards(filters Filter) {
	var (
		rawBoard   []byte
		folderName string = ""
		folderId   int
	)
	path := getResourcePath(config.DashboardResource)
	filesInDir, err := s.storage.FindAllFiles(path, true)
	if err != nil {
		log.WithError(err).Fatal("unable to find any files to export from storage engine")
	}
	ctx := context.Background()

	folderMap := getFolderNameIDMap(s.client, ctx)

	// Fallback on defaults
	if filters == nil {
		filters = &DashboardFilter{}
	}
	validFolders := filters.GetFolders()
	for _, file := range filesInDir {
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")
		if strings.HasSuffix(file, ".json") {
			if rawBoard, err = s.storage.ReadFile(file); err != nil {
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
			if !funk.Contains(validFolders, folderName) && !apphelpers.GetCtxDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
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

// DeleteAllDashboards clears all current dashboards being monitored.  Any folder not white listed
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
