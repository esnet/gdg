package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/esnet/gdg/internal/service/filters/v1"
	v2 "github.com/esnet/gdg/internal/service/filters/v2"
	"github.com/tidwall/gjson"

	"github.com/gosimple/slug"

	"github.com/esnet/gdg/internal/tools/encode"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	customTypes "github.com/esnet/gdg/internal/types"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/tidwall/pretty"
	"golang.org/x/exp/maps"
)

func setupDashReaders(filterObj filters.V2Filter) {
	obj := customTypes.NestedHit{}
	err := filterObj.RegisterReader(reflect.TypeOf(&obj), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.(*customTypes.NestedHit)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.FolderFilter:
			return val.NestedPath, nil
		case filters.TagsFilter:
			return val.Tags, nil
		case filters.DashFilter:
			return slug.Make(val.Title), nil

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, aborting.")
	}
	err = filterObj.RegisterReader(reflect.TypeOf([]byte{}), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.FolderFilter:
			{
				r := gjson.GetBytes(val, "folderTitle")
				if !r.Exists() {
					return "General", nil
				}
				return r.String(), nil
			}
		case filters.TagsFilter:
			{
				r := gjson.GetBytes(val, "tags")
				if !r.Exists() || !r.IsArray() {
					return nil, fmt.Errorf("no valid title found")
				}
				ar := r.Array()
				data := lo.Map(ar, func(item gjson.Result, index int) string {
					return item.String()
				})
				return data, nil

			}
			// return val.Tags, nil
		case filters.DashFilter:
			{
				r := gjson.GetBytes(val, "title")
				if !r.Exists() || r.String() == "" {
					return nil, fmt.Errorf("no valid title found")
				}
				return r.String(), nil
			}
		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
}

func NewDashboardFilterV2(entries ...string) filters.V2Filter {
	if len(entries) != 3 {
		log.Fatalf("Unable to create a valid Dashboard Filter, aborting.")
	}
	folderFilter := entries[0]
	dashboardFilter := entries[1]
	tagsFilter := entries[2]
	var tagObj []string
	if tagsFilter != "" {
		err := json.Unmarshal([]byte(tagsFilter), &tagObj)
		if err != nil {
			log.Fatalf("Unable to create a valid Dashboard Filter, aborting.")
		}
		tagsFilter = "[]"
	}
	filterObj := v2.NewBaseFilter()
	setupDashReaders(filterObj)
	// Setup Readers

	err := filterObj.RegisterDataProcessor(filters.FolderFilter, filters.ProcessorEntity{
		Name: "folderQuoteRegEx",
		Processor: func(item any) (any, error) {
			switch w := item.(type) {
			case string:
				slog.Debug("folder quote filter applied to string")
				quoteRegex, _ := regexp.Compile("['\"]+")
				w = quoteRegex.ReplaceAllString(w, "")
				return w, nil
			case []string:
				slog.Debug("folder quote filter applied to []string")
				return lo.Map(w, func(i string, index int) string {
					quoteRegex, _ := regexp.Compile("['\"]+")
					i = quoteRegex.ReplaceAllString(i, "")
					return i
				}), nil
			}
			return item, nil
		},
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, aborting.")
	}
	var folderArr []string
	if folderFilter != "" {
		folderArr = []string{folderFilter}
	} else {
		folderArr = config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders()
	}
	filterObj.AddValidation(filters.FolderFilter, func(value any, expected any) error {
		val, expressions, convErr := v2.GetMismatchParams[string, []string](value, expected, filters.FolderFilter)
		if convErr != nil {
			return convErr
		}
		for _, exp := range expressions {
			r, ReErr := regexp.Compile(exp)
			if ReErr != nil {
				return fmt.Errorf("invalid regex: %s", exp)
			}
			if r.MatchString(val) {
				return nil
			}
		}

		return fmt.Errorf("invalid folder filter. Expected: %v", expressions)
	}, folderArr)
	filterObj.AddValidation(filters.DashFilter, func(value any, expected any) error {
		val, exp, convErr := v2.GetParams[string](value, expected, filters.DashFilter)
		if convErr != nil {
			return convErr
		}
		if val == "" {
			return nil
		}
		if exp == val {
			return fmt.Errorf("failed validation test val:%s  expected: %s", val, exp)
		}
		return nil
	}, dashboardFilter)

	filterObj.AddValidation(filters.TagsFilter, func(value any, expected any) error {
		return nil
	}, tagObj)

	return filterObj
}

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

	filterObj := v1.NewBaseFilter()
	filterObj.AddFilter(filters.FolderFilter, folderFilter)
	filterObj.AddFilter(filters.DashFilter, dashboardFilter)
	filterObj.AddFilter(filters.TagsFilter, tagsFilter)
	quoteRegex, _ := regexp.Compile("['\"]+")
	filterObj.AddRegex(filters.FolderFilter, quoteRegex)
	// Add Folder Validation
	filterObj.AddValidation(filters.FolderFilter, func(i any) bool {
		val, ok := i.(map[filters.FilterType]string)
		if !ok {
			return ok
		}
		// Check folder
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

	// Add DashValidation
	filterObj.AddValidation(filters.DashFilter, func(i any) bool {
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

// validateFolderRegex accepts a list of regular expression and a folder name or path.  Returns true if any of the regex matches
func validateFolderRegex(acceptList []string, folder string) bool {
	for _, pattern := range acceptList {
		p, err := regexp.Compile(pattern)
		if err != nil {
			// fallback on exact string match
			if pattern == folder {
				return true
			}
			continue
		}
		if p.MatchString(folder) {
			return true
		}
	}
	return false
}

// ListDashboards List all dashboards optionally filtered by folder name. If folderFilters
// is blank, defaults to the configured Monitored folders
func (s *DashNGoImpl) ListDashboards(filterReq filters.V2Filter) []*customTypes.NestedHit {
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilterV2("", "", "")
	}

	boardLinks := make([]*customTypes.NestedHit, 0)
	deduplicatedLinks := make(map[int64]*customTypes.NestedHit)

	var page int64 = 1
	var limit int64 = 5000 // Upper bound of Grafana API call

	tagsParams := make([]string, 0)
	tagExpected := filterReq.GetExpectedValue(filters.TagsFilter)
	if val, ok := tagExpected.([]string); ok {
		tagsParams = append(tagsParams, val...)
	}
	watchedFolders := s.grafanaConf.GetMonitoredFolders()

	retrieve := func(tag string) {
		for {
			searchParams := search.NewSearchParams()
			if tag != "" {
				searchParams.Tag = []string{tag}
			}
			searchParams.Limit = ptr.Of(limit)
			searchParams.Page = ptr.Of(page)
			searchParams.Type = ptr.Of(searchTypeDashboard)

			pageBoardLinks, err := s.GetClient().Search.Search(searchParams)
			if err != nil {
				log.Fatal("Failed to retrieve dashboards", err)
			}
			boardLinks = append(boardLinks,
				lo.Map(pageBoardLinks.GetPayload(), func(item *models.Hit, index int) *customTypes.NestedHit {
					return &customTypes.NestedHit{Hit: item}
				})...)
			if int64(len(pageBoardLinks.GetPayload())) < limit {
				break
			}
			page += 1
		}
	}
	if len(tagsParams) == 0 {
		retrieve("")
	} else {
		// need to iterate over all tags since grafana API filters on AND (&&) instead of OR (||)
		for _, tag := range tagsParams {
			retrieve(tag)
			slog.Info("retrieving dashboard by tag", slog.String("tag", tag))
		}
	}

	folderUid := getFolderUIDEntityMap(s.ListFolders(nil))
	folderFilters, err := filterReq.GetExpectedStringSlice(filters.FolderFilter)
	if err != nil {
		folderFilters = s.grafanaConf.GetMonitoredFolders()
	}
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
		folderMatch := link.FolderTitle
		if folderMatch == "" {
			folderMatch = DefaultFolderName
		}
		folderMatch = getNestedFolder(folderMatch, link.FolderUID, folderUid)
		link.NestedPath = folderMatch

		// accepts all folders
		if config.Config().GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters {
			// if folder filter parameter is enabled, apply given filter
			overlap := lo.Intersect(watchedFolders, folderFilters)
			if len(overlap) != len(folderFilters) {
				// reuse filters below since they are intended to be ignored when ignore filters  enabled
				config.Config().GetDefaultGrafanaConfig().MonitoredFoldersOverride = nil
				// set monitored folders from CLI param filter
				config.Config().GetDefaultGrafanaConfig().MonitoredFolders = folderFilters
				if validateFolderRegex(folderFilters, folderMatch) { // ensure folder matches CLI param filter
					validFolder = true
				}
			} else {
				validFolder = true
			}
		} else if filterReq.Validate(filters.FolderFilter, link) { // validateFolderRegex(folderFilters, folderMatch) { // ensure folder matches con
			validFolder = true
		} else if slices.Contains(folderFilters, DefaultFolderName) && link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
			validFolder = true
		}

		if !validFolder {
			slog.Debug("Skipping dashboard, as it failed the filter check", "title", link.Title, "folder", link.NestedPath)
			continue
		}

		validUid = filterReq.GetExpectedString(filters.DashFilter) == "" || link.Slug == filterReq.GetExpectedString(filters.DashFilter)
		if link.FolderID == 0 && string(link.Type) == searchTypeDashboard {
			link.FolderTitle = DefaultFolderName
		}
		// check folder

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

// ListDashboards List all dashboards optionally filtered by folder name. If folderFilters
// is blank, defaults to the configured Monitored folders
func (s *DashNGoImpl) ListDashboardsLegacy(filterReq filters.Filter) []*customTypes.NestedHit {
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}

	boardLinks := make([]*customTypes.NestedHit, 0)
	deduplicatedLinks := make(map[int64]*customTypes.NestedHit)

	var page int64 = 1
	var limit int64 = 5000 // Upper bound of Grafana API call

	tagsParams := make([]string, 0)
	tagsParams = append(tagsParams, filterReq.GetEntity(filters.TagsFilter)...)
	watchedFolders := s.grafanaConf.GetMonitoredFolders()

	retrieve := func(tag string) {
		for {
			searchParams := search.NewSearchParams()
			if tag != "" {
				searchParams.Tag = []string{tag}
			}
			searchParams.Limit = ptr.Of(limit)
			searchParams.Page = ptr.Of(page)
			searchParams.Type = ptr.Of(searchTypeDashboard)

			pageBoardLinks, err := s.GetClient().Search.Search(searchParams)
			if err != nil {
				log.Fatal("Failed to retrieve dashboards", err)
			}
			boardLinks = append(boardLinks,
				lo.Map(pageBoardLinks.GetPayload(), func(item *models.Hit, index int) *customTypes.NestedHit {
					return &customTypes.NestedHit{Hit: item}
				})...)
			if int64(len(pageBoardLinks.GetPayload())) < limit {
				break
			}
			page += 1
		}
	}
	if len(tagsParams) == 0 {
		retrieve("")
	} else {
		// need to iterate over all tags since grafana API filters on AND (&&) instead of OR (||)
		for _, tag := range tagsParams {
			retrieve(tag)
			slog.Info("retrieving dashboard by tag", slog.String("tag", tag))
		}
	}

	folderUid := getFolderUIDEntityMap(s.ListFolders(nil))
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
		folderMatch := link.FolderTitle
		if folderMatch == "" {
			folderMatch = DefaultFolderName
		}
		folderMatch = getNestedFolder(folderMatch, link.FolderUID, folderUid)
		link.NestedPath = folderMatch

		// accepts all folders
		if config.Config().GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters {
			// if folder filter parameter is enabled, apply given filter
			overlap := lo.Intersect(watchedFolders, folderFilters)
			if len(overlap) != len(folderFilters) {
				// reuse filters below since they are intended to be ignored when ignore filters  enabled
				config.Config().GetDefaultGrafanaConfig().MonitoredFoldersOverride = nil
				// set monitored folders from CLI param filter
				config.Config().GetDefaultGrafanaConfig().MonitoredFolders = folderFilters
				if validateFolderRegex(folderFilters, folderMatch) { // ensure folder matches CLI param filter
					validFolder = true
				}
			} else {
				validFolder = true
			}
		} else if validateFolderRegex(folderFilters, folderMatch) { // ensure folder matches con
			validFolder = true
		} else if slices.Contains(folderFilters, DefaultFolderName) && link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
			validFolder = true
		}

		if !validFolder {
			slog.Debug("Skipping dashboard, as it failed the filter check", "title", link.Title, "folder", link.NestedPath)
			continue
		}
		validUid = filterReq.GetFilter(filters.DashFilter) == "" || link.Slug == filterReq.GetFilter(filters.DashFilter)
		if link.FolderID == 0 && string(link.Type) == searchTypeDashboard {
			link.FolderTitle = DefaultFolderName
		}
		// check folder

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
func (s *DashNGoImpl) DownloadDashboards(filter filters.V2Filter) []string {
	var (
		boardLinks []*customTypes.NestedHit
		rawBoard   []byte
		err        error
		metaData   *dashboards.GetDashboardByUIDOK
	)

	boardLinks = s.ListDashboards(filter)

	var boards []string
	for _, link := range boardLinks {
		if string(link.Type) != searchTypeDashboard {
			slog.Debug("Ignoring dashboard-folder", "folder", link.Title)
			continue
		}

		if metaData, err = s.GetClient().Dashboards.GetDashboardByUID(link.UID); err != nil {
			slog.Error("unable to get Dashboard by UID", "err", err, "Dashboard-URI", link.URI)
			continue
		}

		rawBoard, err = json.Marshal(metaData.GetPayload().Dashboard)
		if err != nil {
			slog.Error("unable to serialize dashboard", "dashboard", link.UID)
			continue
		}

		// fileName := buildDashboardFileName(link.NestedPath, metaData.GetPayload().Meta.Slug, folderUidMap, s.isLocal(), s.globalConf.ClearOutput)
		fileName := fmt.Sprintf("%s/%s.json", BuildResourceFolder(link.NestedPath, config.DashboardResource, s.isLocal(), s.globalConf.ClearOutput), metaData.GetPayload().Meta.Slug)
		if err = s.storage.WriteFile(fileName, pretty.Pretty(rawBoard)); err != nil {
			slog.Error("Unable to save dashboard to file\n", "err", err, "dashboard", metaData.GetPayload().Meta.Slug)
		} else {
			boards = append(boards, fileName)
		}

	}
	return boards
}

// getNestedFolder use this if calling from within the service, returns the nested folder path for a given folder
func getNestedFolder(folderTitle, folderUID string, folderUidMap map[string]*customTypes.NestedHit) string {
	folderPath := encode.Encode(folderTitle)
	currentFolderUid := folderUID
	for currentFolderUid != "" {
		parent, ok := folderUidMap[currentFolderUid]
		if ok && parent.FolderUID != "" {
			currentFolderUid = parent.FolderUID
			folderPath = fmt.Sprintf("%s/%s", encode.Encode(parent.FolderTitle), folderPath)
		} else {
			currentFolderUid = ""
		}

	}
	return folderPath
}

// buildDashboardFileName for a given dashboard, a full nested folder path is constructed
//func buildDashboardFileName(folderPath, dbSlug string, folderUidMap map[string]*customTypes.NestedHit, createDestination, clearOutput bool) string {
//	fileName := fmt.Sprintf("%s/%s.json", BuildResourceFolder(folderPath, config.DashboardResource, createDestination, clearOutput), dbSlug)
//	return fileName
//}

// createFolders Creates a new folder with the given name.  If nested, each sub folder that does not exist is also created
func (s *DashNGoImpl) createdFolders(folderName string) (map[string]string, error) {
	folderUidMap := getFolderUIDEntityMap(s.ListFolders(nil))

	namedUIDMap := getFolderMapping(s.ListFolders(NewFolderFilter()),
		func(db *customTypes.NestedHit) string {
			return getNestedFolder(db.Title, db.FolderUID, folderUidMap)
		},
		func(fld *customTypes.NestedHit) *customTypes.NestedHit { return fld },
	)

	cratedBaseFolder := func(createFolder string, parent string) (string, error) {
		request := &models.CreateFolderCommand{
			Title:     createFolder,
			ParentUID: parent,
		}
		res, err := s.GetClient().Folders.CreateFolder(request)
		if err != nil {
			return "", err
		}
		return res.GetPayload().UID, nil
	}
	newFoldersMap := make(map[string]string)

	folderPath := strings.Builder{}
	parentUid := ""
	const pathSeparator = string(os.PathSeparator)
	if strings.Contains(folderName, pathSeparator) {
		elements := strings.Split(folderName, pathSeparator)
		for ndx, folder := range elements {
			var (
				cnt     int
				pathErr error
			)
			// folder = encode.Decode(folder)
			if ndx == 0 {
				cnt, pathErr = folderPath.WriteString(folder)
			} else {
				cnt, pathErr = folderPath.WriteString(fmt.Sprintf("/%s", folder))
			}

			if pathErr != nil || cnt <= 0 {
				log.Fatal("unable to update folder path, critical logic error")
			}
			if val, ok := namedUIDMap[folderPath.String()]; ok {
				parentUid = val.UID
			} else {
				uid, err := cratedBaseFolder(encode.Decode(folder), parentUid)
				if err != nil {
					return newFoldersMap, err
				}
				newFoldersMap[folderPath.String()] = uid
				parentUid = uid
			}
		}
	} else { // Handles simple case
		data, err := cratedBaseFolder(folderName, "")
		if err == nil {
			newFoldersMap[folderName] = data
		}
		return newFoldersMap, err
	}

	return newFoldersMap, nil
}

// UploadDashboards finds all the dashboards in the configured location and exports them to grafana.
// if the folder doesn't exist, it'll be created.
func (s *DashNGoImpl) UploadDashboards(filterReq filters.Filter) ([]string, error) {
	var (
		rawBoard   []byte
		folderName string
		folderUid  string
		dashFiles  []string
	)
	dashboardPath := config.Config().GetDefaultGrafanaConfig().GetPath(config.DashboardResource)
	filesInDir, err := s.storage.FindAllFiles(dashboardPath, true)
	if err != nil {
		return nil, fmt.Errorf("unable to find any dashFiles to export from storage engine, err: %w", err)
	}

	currentDashboards := s.ListDashboardsLegacy(filterReq)

	folderUidMap := s.getFolderNameUIDMap(s.ListFolders(NewFolderFilter()))

	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}

	// validFolders := filterReq.GetEntity(filters.FolderFilter)
	alreadyProcessed := make(map[any]bool)

	for _, file := range filesInDir {
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")

		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json dashFiles are supported, skipping", "filename", file)
			continue
		}

		if rawBoard, err = s.storage.ReadFile(file); err != nil {
			slog.Warn("Unable to read file", "filename", file, "err", err)
			continue
		}
		board := make(map[string]any)
		if err = json.Unmarshal(rawBoard, &board); err != nil {
			slog.Warn("Failed to unmarshall file", "filename", file)
			continue
		}
		if _, ok := alreadyProcessed[board["uid"]]; ok {
			return nil, fmt.Errorf("Board with same UID was already processed.  Please check your backup folder. This may occur if you pulled the data multiple times with configuration of: nested folder enabled and disabled, uid: %v, title: %v", board["uid"], slug.Make((board["title"]).(string)))
		} else {
			alreadyProcessed[board["uid"]] = true
		}

		// Extract Folder Name based on dashboardPath
		folderName, err = getFolderFromResourcePath(file, config.DashboardResource, s.storage.GetPrefix())
		if err != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default")
		}

		if folderName == "" {
			folderName = DefaultFolderName
		}
		folderUidMap, err = s.validateDashUploadEntity(filterReq, folderName, baseFile, &folderUid, folderUidMap, board["tags"])
		if err != nil {
			slog.Warn("validation failed, skipping", "file", file, "err", err)
			continue
		}

		// zero out ID.  Can't create a new dashboard if an ID already exists.
		delete(board, "id")
		importDashReq := &models.ImportDashboardRequest{
			FolderUID: folderUid,
			Overwrite: true,
			Dashboard: board,
		}

		if _, exportError := s.GetClient().Dashboards.ImportDashboard(importDashReq); exportError != nil {
			slog.Info("error on Exporting dashboard", "dashboard-filename", file, "err", exportError)
			continue
		} else {
			dashFiles = append(dashFiles, file)
		}

	}

	for _, item := range currentDashboards {
		if ok := alreadyProcessed[item.UID]; !ok {
			slog.Info("Deleting Dashboard not found in backup", "folder", item.FolderTitle, "dashboard", item.Title)
			err := s.deleteDashboard(item.Hit)
			if err != nil {
				slog.Error("Unable to delete dashboard", "folder", item.FolderTitle, "dashboard", item.Title)
			}
		}
	}
	return dashFiles, nil
}

// TODO: Migrate to be part of filter
func (s *DashNGoImpl) validateDashUploadEntity(filterReq filters.Filter, folderName string, baseFile string, folderUid *string, folderUidMap map[string]string, rawBoardTags any) (map[string]string, error) {
	// Extract Tags
	tagFilter := filterReq.GetFilter(filters.TagsFilter)
	if filterVal := tagFilter; filterVal != "[]" && filterVal != "" {
		var boardTags []string
		for _, val := range rawBoardTags.([]any) {
			boardTags = append(boardTags, val.(string))
		}
		var requestedSlices []string
		err := json.Unmarshal([]byte(filterVal), &requestedSlices)
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
			return folderUidMap, fmt.Errorf("dashboard fails to pass tag filter: tagFilter: %s", tagFilter)
		}

	}
	validateMap := map[filters.FilterType]string{filters.FolderFilter: folderName, filters.DashFilter: baseFile}
	// If folder OR slug is filtered, then skip if it doesn't match
	if !s.grafanaConf.GetDashboardSettings().IgnoreFilters && (!filterReq.ValidateAll(validateMap) || !validateFolderRegex(filterReq.GetEntity(filters.FolderFilter), folderName)) {
		return folderUidMap, errors.New("dashboard fails to pass filter")
	}

	if folderName == DefaultFolderName {
		*folderUid = ""
	} else {
		if val, ok := folderUidMap[folderName]; ok {
			// folderId = val
			*folderUid = val
		} else {
			if s.grafanaConf.GetDashboardSettings().IgnoreFilters || filterReq.ValidateAll(validateMap) {
				newFolders, folderErr := s.createdFolders(folderName)
				if folderErr != nil {
					log.Panic("Unable to create required folder")
				} else {
					maps.Copy(folderUidMap, newFolders)
					*folderUid = folderUidMap[folderName]
				}
			}
		}
	}
	return folderUidMap, nil
}

// deleteDashboard removes a dashboard from grafana.  If the dashboard doesn't exist,
// an error is returned.
//
// Parameters:
// item - dashboard to be deleted
//
// Returns:
// error - error returned from the grafana API
func (s *DashNGoImpl) deleteDashboard(item *models.Hit) error {
	_, err := s.GetClient().Dashboards.DeleteDashboardByUID(item.UID)
	return err
}

// DeleteAllDashboards clears all current dashboards being monitored.  Any folder not white listed
// will not be affected
func (s *DashNGoImpl) DeleteAllDashboards(filter filters.Filter) []string {
	dashboardListing := make([]string, 0)

	items := s.ListDashboardsLegacy(filter)
	for _, item := range items {
		if filter.ValidateAll(map[filters.FilterType]string{filters.FolderFilter: item.FolderTitle, filters.DashFilter: item.Slug}) {
			err := s.deleteDashboard(item.Hit)
			if err == nil {
				dashboardListing = append(dashboardListing, item.Title)
			} else {
				slog.Warn("Unable to remove dashboard", slog.String("title", item.Title), slog.String("uid", item.UID))
			}
		}
	}
	return dashboardListing
}
