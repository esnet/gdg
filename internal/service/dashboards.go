package service

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/tools"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/pretty"
	"golang.org/x/exp/slices"
	"log"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thoas/go-funk"
)

// DashboardsApi Contract definition
type DashboardsApi interface {
	ListDashboards(filter filters.Filter) []*models.Hit
	DownloadDashboards(filter filters.Filter) []string
	UploadDashboards(filter filters.Filter)
	DeleteAllDashboards(filter filters.Filter) []string
}

// getDashboardByUid retrieve a dashboard given a particular uid.
func (s *DashNGoImpl) getDashboardByUid(uid string) (*models.DashboardFullWithMeta, error) {
	params := dashboards.NewGetDashboardByUIDParams()
	params.UID = uid
	data, err := s.client.Dashboards.GetDashboardByUID(params, s.getAuth())
	if err != nil {
		return nil, err
	}
	return data.GetPayload(), nil

}

func NewDashboardFilter(entries ...string) filters.Filter {
	if len(entries) != 3 {
		log.Fatalf("Unable to create a valid Dashboard Filter, aborting.")
	}
	folderFilter := entries[0]
	dashboardFilter := entries[1]
	tagsFilter := entries[2]

	filterObj := filters.NewBaseFilter()
	filterObj.AddFilter(filters.FolderFilter, folderFilter)
	filterObj.AddFilter(filters.DashFilter, dashboardFilter)
	filterObj.AddFilter(filters.TagsFilter, tagsFilter)
	quoteRegex, _ := regexp.Compile("['\"]+")
	filterObj.AddRegex(filters.TagsFilter, quoteRegex)
	filterObj.AddRegex(filters.FolderFilter, quoteRegex)
	//Add Folder Validation
	filterObj.AddValidation(filters.FolderFilter, func(i interface{}) bool {
		val, ok := i.(map[filters.FilterType]string)
		if !ok {
			return ok
		}
		//Check folder
		if folderFilter, ok = val[filters.FolderFilter]; ok {
			if filterObj.GetFilter(filters.FolderFilter) == "" {
				return true
			} else {
				return folderFilter == filterObj.GetFilter(filters.FolderFilter)
			}
		} else {
			return true
		}
	})

	//Add Tag Validation
	filterObj.AddValidation(filters.TagsFilter, func(i interface{}) bool {
		val, ok := i.(map[filters.FilterType]string)
		if !ok {
			return ok
		}

		//Check Tags
		if tagsFilter, ok = val[filters.TagsFilter]; ok {
			if filterObj.GetFilter(filters.TagsFilter) == "" {
				return true
			}
			return tagsFilter == filterObj.GetFilter(filters.TagsFilter)
		} else {
			return true
		}
		//Check Dashboard

	})
	//Add DashValidation
	filterObj.AddValidation(filters.DashFilter, func(i interface{}) bool {
		val, ok := i.(map[filters.FilterType]string)
		if !ok {
			return ok
		}

		if dashboardFilter, ok = val[filters.DashFilter]; ok {
			if filterObj.GetFilter(filters.DashFilter) == "" {
				return true
			}
			return dashboardFilter == filterObj.GetFilter(filters.DashFilter)
		} else {
			return true
		}

	})

	return filterObj
}

// ListDashboards List all dashboards optionally filtered by folder name. If folderFilters
// is blank, defaults to the configured Monitored folders
func (s *DashNGoImpl) ListDashboards(filterReq filters.Filter) []*models.Hit {
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}

	var boardsList = make([]*models.Hit, 0)
	var boardLinks = make([]*models.Hit, 0)

	var page uint = 1
	var limit uint = 5000 // Upper bound of Grafana API call

	var tagsParams = make([]string, 0)
	if !config.Config().GetDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
		tagsParams = append(tagsParams, filterReq.GetEntity(filters.TagsFilter)...)
	}

	for {
		searchParams := search.NewSearchParams()
		searchParams.Tag = tagsParams
		searchParams.Limit = tools.PtrOf(int64(limit))
		searchParams.Page = tools.PtrOf(int64(page))
		searchParams.Type = tools.PtrOf(searchTypeDashboard)

		pageBoardLinks, err := s.client.Search.Search(searchParams, s.getAuth())
		if err != nil {
			log.Fatal("Failed to retrieve dashboards", err)
		}
		boardLinks = append(boardLinks, pageBoardLinks.GetPayload()...)
		if len(pageBoardLinks.GetPayload()) < int(limit) {
			break
		}
		page += 1
	}

	folderFilters := filterReq.GetEntity(filters.FolderFilter)
	var validFolder bool
	var validUid bool
	for _, link := range boardLinks {
		validFolder = false
		if config.Config().GetDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
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
		link.Slug = updateSlug(link.URI)
		validUid = filterReq.GetFilter(filters.DashFilter) == "" || link.Slug == filterReq.GetFilter(filters.DashFilter)
		if link.FolderID == 0 {

			link.FolderTitle = DefaultFolderName
		}

		if validFolder && validUid {
			boardsList = append(boardsList, link)
		}
	}

	return boardsList

}

// DownloadDashboards saves all dashboards matching query to configured location
func (s *DashNGoImpl) DownloadDashboards(filter filters.Filter) []string {
	var (
		boardLinks []*models.Hit
		rawBoard   []byte
		err        error
		metaData   *dashboards.GetDashboardByUIDOK
	)

	boardLinks = s.ListDashboards(filter)
	var boards []string
	for _, link := range boardLinks {
		dp := dashboards.NewGetDashboardByUIDParams()
		dp.UID = link.UID

		if metaData, err = s.client.Dashboards.GetDashboardByUID(dp, s.getAuth()); err != nil {
			slog.Error("unable to get Dashboard by UID", "err", err, "Dashboard-URI", link.URI)
			continue
		}

		rawBoard, err = json.Marshal(metaData.Payload.Dashboard)
		if err != nil {
			slog.Error("unable to serialize dashboard", "dashboard", dp.UID)
			continue
		}

		fileName := fmt.Sprintf("%s/%s.json", BuildResourceFolder(link.FolderTitle, config.DashboardResource), metaData.Payload.Meta.Slug)
		if err = s.storage.WriteFile(fileName, pretty.Pretty(rawBoard)); err != nil {
			slog.Error("Unable to save dashboard to file\n", "err", err, "dashboard", metaData.Payload.Meta.Slug)
		} else {
			boards = append(boards, fileName)
		}

	}
	return boards
}

// createFolder Creates a new folder with the given name.
func (s *DashNGoImpl) createdFolder(folderName string) (int64, error) {
	createdFolderRequest := folders.NewCreateFolderParams()
	createdFolderRequest.Body = &models.CreateFolderCommand{
		Title: folderName,
	}
	folder, err := s.client.Folders.CreateFolder(createdFolderRequest, s.getAuth())
	if err != nil {
		return 0, err
	}
	return folder.GetPayload().ID, nil

}

// UploadDashboards finds all the dashboards in the configured location and exports them to grafana.
// if the folder doesn't exist, it'll be created.
func (s *DashNGoImpl) UploadDashboards(filterReq filters.Filter) {

	var (
		rawBoard   []byte
		folderName string
		folderId   int64
	)
	path := config.Config().GetDefaultGrafanaConfig().GetPath(config.DashboardResource)
	filesInDir, err := s.storage.FindAllFiles(path, true)
	if err != nil {
		log.Fatalf("unable to find any files to export from storage engine, err: %v", err)
	}
	//Delete all dashboards that match prior to import
	s.DeleteAllDashboards(filterReq)

	folderMap := getFolderNameIDMap(s.ListFolder(NewFolderFilter()))

	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}
	validFolders := filterReq.GetEntity(filters.FolderFilter)
	for _, file := range filesInDir {
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")

		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json files are supported, skipping", "filename", file)
			continue
		}

		if rawBoard, err = s.storage.ReadFile(file); err != nil {
			slog.Warn("Unable to read file", "filename", file, "err", err)
			continue
		}
		var board = make(map[string]interface{})
		if err = json.Unmarshal(rawBoard, &board); err != nil {
			slog.Warn("Failed to unmarshall file", "filename", file)
			continue
		}

		//Extract Folder Name based on path
		folderName, err = getFolderFromResourcePath(s.grafanaConf.Storage, file, config.DashboardResource)
		if err != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default")
		}

		if folderName == "" || folderName == DefaultFolderName {
			folderId = DefaultFolderId
			folderName = DefaultFolderName
		}
		if !slices.Contains(validFolders, folderName) && !config.Config().GetDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
			slog.Debug("Skipping file since it doesn't match any valid folders", "filename", file)
			continue
		}
		validateMap := map[filters.FilterType]string{filters.FolderFilter: folderName, filters.DashFilter: baseFile}
		//If folder OR slug is filtered, then skip if it doesn't match
		if !filterReq.ValidateAll(validateMap) {
			continue
		}

		if folderName == DefaultFolderName {
			folderId = DefaultFolderId
		} else {

			if val, ok := folderMap[folderName]; ok {
				folderId = val
			} else {
				if filterReq.ValidateAll(validateMap) {
					id, folderErr := s.createdFolder(folderName)
					if folderErr != nil {
						log.Panic("Unable to create required folder")
					} else {
						folderMap[folderName] = id
						folderId = id
					}
				}
			}
		}

		data := make(map[string]interface{})

		err = json.Unmarshal(rawBoard, &data)
		if err != nil {
			slog.Error("Unable to marshall data to valid JSON, skipping import", slog.Any("data", rawBoard))
			continue
		}
		//zero out ID.  Can't create a new dashboard if an ID already exists.
		delete(data, "id")
		importDashReq := dashboards.NewImportDashboardParams()
		importDashReq.Body = &models.ImportDashboardRequest{
			FolderID:  folderId,
			Overwrite: true,
			Dashboard: data,
		}

		if _, exportError := s.client.Dashboards.ImportDashboard(importDashReq, s.getAuth()); exportError != nil {
			slog.Info("error on Exporting dashboard", "dashboard-filename", file, "err", exportError)
			continue
		}

	}
}

// DeleteAllDashboards clears all current dashboards being monitored.  Any folder not white listed
// will not be affected
func (s *DashNGoImpl) DeleteAllDashboards(filter filters.Filter) []string {
	var dashboardListing = make([]string, 0)

	items := s.ListDashboards(filter)
	for _, item := range items {
		if filter.ValidateAll(map[filters.FilterType]string{filters.FolderFilter: item.FolderTitle, filters.DashFilter: item.Slug}) {
			dp := dashboards.NewDeleteDashboardByUIDParams()
			dp.UID = item.UID
			_, err := s.client.Dashboards.DeleteDashboardByUID(dp, s.getAuth())
			if err == nil {
				dashboardListing = append(dashboardListing, item.Title)
			}
		}
	}
	return dashboardListing

}
