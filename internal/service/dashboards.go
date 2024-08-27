package service

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/service/types"
	"github.com/esnet/gdg/internal/tools"
	"github.com/grafana/dashboard-linter/lint"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/pretty"
	"github.com/zeitlinger/conflate"
	"golang.org/x/exp/maps"
	"log"
	"log/slog"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/thoas/go-funk"
)

func NewDashboardFilter(entries ...string) filters.Filter {
	if len(entries) != 3 {
		log.Fatalf("Unable to create a valid Dashboard Filter, aborting.")
	}
	folderFilter := entries[0]
	dashboardFilter := entries[1]
	tagsFilter := entries[2]
	if tagsFilter == "" {
		tagsFilter = "[]"
	}

	filterObj := filters.NewBaseFilter()
	filterObj.AddFilter(filters.FolderFilter, folderFilter)
	filterObj.AddFilter(filters.DashFilter, dashboardFilter)
	filterObj.AddFilter(filters.TagsFilter, tagsFilter)
	quoteRegex, _ := regexp.Compile("['\"]+")
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

func (s *DashNGoImpl) LintDashboards(req types.LintRequest) []string {
	var (
		rawBoard []byte
	)
	dashboardPath := config.Config().GetDefaultGrafanaConfig().GetPath(config.DashboardResource)
	filesInDir, err := s.storage.FindAllFiles(dashboardPath, true)
	if err != nil {
		log.Fatalf("unable to find any files to export from storage engine, err: %v", err)
	}
	filterReq := NewDashboardFilter(req.FolderName, req.DashboardSlug, "")
	validFolders := filterReq.GetEntity(filters.FolderFilter)
	for _, file := range filesInDir {
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")

		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json files are supported, skipping", "filename", file)
			continue
		}
		if req.DashboardSlug != "" && baseFile != req.DashboardSlug {
			slog.Debug("Skipping dashboard, does not match filter", slog.String("dashboard", req.DashboardSlug))
			continue
		}

		if rawBoard, err = s.storage.ReadFile(file); err != nil {
			slog.Warn("Unable to read file", "filename", file, "err", err)
			continue
		}
		if req.FolderName != "" {
			if !slices.Contains(validFolders, req.FolderName) && !config.Config().GetDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
				slog.Debug("Skipping file since it doesn't match any valid folders", "filename", file)
				continue
			}
		}

		dashboard, err := lint.NewDashboard(rawBoard)
		if err != nil {
			slog.Error("failed to parse dashboard", slog.Any("err", err))
			continue
		}
		lintConfigFlag := strings.ReplaceAll(file, ".json", ".lint")
		cfgLint := lint.NewConfigurationFile()
		if err := cfgLint.Load(lintConfigFlag); err != nil {
			slog.Error("Unable to load lintConfigFlag")
			continue
		}
		cfgLint.Verbose = req.VerboseFlag
		cfgLint.Autofix = req.AutoFix

		rules := lint.NewRuleSet()
		results, err := rules.Lint([]lint.Dashboard{dashboard})
		if err != nil {
			slog.Error("failed to lint dashboard", slog.Any("err", err))
			continue

		}
		if cfgLint.Autofix {
			changes := results.AutoFix(&dashboard)
			if changes > 0 {
				slog.Info("AutoFix possible")
				writeErr := s.writeLintedDashboard(dashboard, file, rawBoard)
				if writeErr != nil {
					slog.Error("unable to autofix linting issues for dashboard", slog.String("dashboard", file))
				}
			} else {
				slog.Error("AutoFix is not possible for dashboard.", slog.String("dashboard", file))
			}
		}

		slog.Info("Running Linter for Dashboard", slog.String("file", file))
		results.ReportByRule()

	}
	return nil
}

func (s *DashNGoImpl) writeLintedDashboard(dashboard lint.Dashboard, filename string, old []byte) error {
	newBytes, err := dashboard.Marshal()
	if err != nil {
		return err
	}
	c := conflate.New()
	err = c.AddData(old, newBytes)
	if err != nil {
		return err
	}
	b, err := c.MarshalJSON()
	if err != nil {
		return err
	}
	json := strings.ReplaceAll(string(b), "\"options\": null,", "\"options\": [],")
	return s.storage.WriteFile(filename, []byte(json))
}

// getDashboardByUid retrieve a dashboard given a particular uid.
func (s *DashNGoImpl) getDashboardByUid(uid string) (*models.DashboardFullWithMeta, error) {
	params := dashboards.NewGetDashboardByUIDParams()
	params.UID = uid
	data, err := s.GetClient().Dashboards.GetDashboardByUID(uid)
	if err != nil {
		return nil, err
	}
	return data.GetPayload(), nil

}

// ListDashboards List all dashboards optionally filtered by folder name. If folderFilters
// is blank, defaults to the configured Monitored folders
func (s *DashNGoImpl) ListDashboards(filterReq filters.Filter) []*models.Hit {
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}

	var boardLinks = make([]*models.Hit, 0)
	var deduplicatedLinks = make(map[int64]*models.Hit)

	var page int64 = 1
	var limit int64 = 5000 // Upper bound of Grafana API call

	var tagsParams = make([]string, 0)
	tagsParams = append(tagsParams, filterReq.GetEntity(filters.TagsFilter)...)

	retrieve := func(tag string) {
		for {
			searchParams := search.NewSearchParams()
			if tag != "" {
				searchParams.Tag = []string{tag}
			}
			searchParams.Limit = tools.PtrOf(limit)
			searchParams.Page = tools.PtrOf(page)
			searchParams.Type = tools.PtrOf(searchTypeDashboard)

			pageBoardLinks, err := s.GetClient().Search.Search(searchParams)
			if err != nil {
				log.Fatal("Failed to retrieve dashboards", err)
			}
			boardLinks = append(boardLinks, pageBoardLinks.GetPayload()...)
			if int64(len(pageBoardLinks.GetPayload())) < limit {
				break
			}
			page += 1
		}
	}
	if len(tagsParams) == 0 {
		retrieve("")
	} else {
		for _, tag := range tagsParams {
			slog.Info("retrieving dashboard by tag", slog.String("tag", tag))
			retrieve(tag)
		}
	}

	folderFilters := filterReq.GetEntity(filters.FolderFilter)
	var validFolder bool
	var validUid bool
	for ndx, link := range boardLinks {
		link.Slug = updateSlug(link.URI)
		_, ok := deduplicatedLinks[link.ID]
		if ok {
			slog.Debug("duplicate board, skipping ")
			continue
		}
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
		validUid = filterReq.GetFilter(filters.DashFilter) == "" || link.Slug == filterReq.GetFilter(filters.DashFilter)
		if link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
		}

		if validUid {
			deduplicatedLinks[link.ID] = boardLinks[ndx]
		}
	}

	boardLinks = maps.Values(deduplicatedLinks)
	sort.Slice(boardLinks, func(i, j int) bool {
		return boardLinks[i].ID < boardLinks[j].ID
	})

	return boardLinks

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
		if metaData, err = s.GetClient().Dashboards.GetDashboardByUID(link.UID); err != nil {
			slog.Error("unable to get Dashboard by UID", "err", err, "Dashboard-URI", link.URI)
			continue
		}

		rawBoard, err = json.Marshal(metaData.Payload.Dashboard)
		if err != nil {
			slog.Error("unable to serialize dashboard", "dashboard", link.UID)
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
func (s *DashNGoImpl) createdFolder(folderName string) (string, error) {
	request := &models.CreateFolderCommand{
		Title: folderName,
	}
	folder, err := s.GetClient().Folders.CreateFolder(request)
	if err != nil {
		return "", err
	}
	return folder.GetPayload().UID, nil

}

// UploadDashboards finds all the dashboards in the configured location and exports them to grafana.
// if the folder doesn't exist, it'll be created.
func (s *DashNGoImpl) UploadDashboards(filterReq filters.Filter) {

	var (
		rawBoard   []byte
		folderName string
		folderUid  string
	)
	path := config.Config().GetDefaultGrafanaConfig().GetPath(config.DashboardResource)
	filesInDir, err := s.storage.FindAllFiles(path, true)
	if err != nil {
		log.Fatalf("unable to find any files to export from storage engine, err: %v", err)
	}
	//Delete all dashboards that match prior to import
	s.DeleteAllDashboards(filterReq)

	folderUidMap := getFolderNameUIDMap(s.ListFolder(NewFolderFilter()))

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
		//Extract Tags
		if filterVal := filterReq.GetFilter(filters.TagsFilter); filterVal != "[]" {
			var boardTags []string
			for _, val := range board["tags"].([]interface{}) {
				boardTags = append(boardTags, val.(string))
			}
			var requestedSlices []string
			err = json.Unmarshal([]byte(filterVal), &requestedSlices)
			if err != nil {
				slog.Warn("unable to decode json of requested tags")
				requestedSlices = []string{}
			}
			valid := false
			for _, val := range requestedSlices {
				if slices.Contains(boardTags, val) {
					valid = true
					break
				}
			}
			if !valid {
				slog.Debug("board fails tag filter, ignoring board", slog.Any("title", board["title"]))
				continue
			}

		}

		//Extract Folder Name based on path
		folderName, err = getFolderFromResourcePath(s.grafanaConf.Storage, file, config.DashboardResource)
		if err != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default")
		}

		if folderName == "" || folderName == DefaultFolderName {
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
			folderUid = ""
		} else {

			if val, ok := folderUidMap[folderName]; ok {
				//folderId = val
				folderUid = val
			} else {
				if filterReq.ValidateAll(validateMap) {
					id, folderErr := s.createdFolder(folderName)
					if folderErr != nil {
						log.Panic("Unable to create required folder")
					} else {
						folderUidMap[folderName] = id
						folderUid = id
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
		importDashReq := &models.ImportDashboardRequest{
			FolderUID: folderUid,
			Overwrite: true,
			Dashboard: data,
		}

		if _, exportError := s.GetClient().Dashboards.ImportDashboard(importDashReq); exportError != nil {
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
			_, err := s.GetClient().Dashboards.DeleteDashboardByUID(item.UID)
			if err == nil {
				dashboardListing = append(dashboardListing, item.Title)
			} else {
				slog.Warn("Unable to remove dashboard", slog.String("title", item.Title), slog.String("uid", item.UID))
			}
		}
	}
	return dashboardListing

}
