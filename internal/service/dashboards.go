package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"

	configDomain "github.com/esnet/gdg/internal/config/domain"

	"github.com/esnet/gdg/internal/service/domain"

	"github.com/esnet/gdg/internal/service/filters/v2"
	"github.com/tidwall/gjson"

	"github.com/gosimple/slug"

	"github.com/esnet/gdg/internal/tools/encode"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/tidwall/pretty"
	"golang.org/x/exp/maps"
)

const (
	NestedDashFolderName = "NestedDashFolderName"
)

func setupDashReaders(filterObj filters.V2Filter) {
	obj := domain.NestedHit{}
	err := filterObj.RegisterReader(reflect.TypeOf(&obj), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.(*domain.NestedHit)
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
		log.Fatalf("Unable to create a valid Dashboard Filter, object reader could not be created, aborting.")
	}
	err = filterObj.RegisterReader(reflect.TypeOf([]byte{}), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
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
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, json reader could not be created, aborting.")
	}
	err = filterObj.RegisterReader(reflect.TypeOf(map[string]any{}), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.FolderFilter:
			{
				return val[NestedDashFolderName], nil
			}
		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, map entity reader could not be created, aborting.")
	}
}

func addFolderFilter(filterReq filters.V2Filter, folderFilter string) {
	var folderArr []string
	if folderFilter != "" {
		config.Config().GetDefaultGrafanaConfig().SetFilterFolder(folderFilter)
		folderArr = []string{folderFilter}
	} else {
		config.Config().GetDefaultGrafanaConfig().ClearFilters()
		folderArr = config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders(false)
	}
	filterReq.AddValidation(filters.FolderFilter, func(value any, expected any) error {
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
}

func NewDashboardFilter(entries ...string) filters.V2Filter {
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
		// tagsFilter = "[]"

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

	addFolderFilter(filterObj, folderFilter)

	filterObj.AddValidation(filters.DashFilter, func(value any, expected any) error {
		val, exp, convErr := v2.GetParams[string](value, expected, filters.DashFilter)
		if convErr != nil {
			return convErr
		}
		if val == "" || exp == "" {
			return nil
		}
		if exp != slug.Make(val) {
			return fmt.Errorf("failed validation test val:%s  expected: %s", val, exp)
		}
		return nil
	}, dashboardFilter)

	filterObj.AddValidation(filters.TagsFilter, func(value any, expected any) error {
		val, exp, convErr := v2.GetParams[[]string](value, expected, filters.TagsFilter)

		if convErr != nil {
			return convErr
		}
		// no filter active, returning nil
		if len(exp) == 0 {
			return nil
		}
		for _, item := range exp {
			if slices.Contains(val, item) {
				return nil
			}
		}

		return fmt.Errorf("failed validation test val:%s  expected: %s", val, exp)
	}, tagObj)

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

// ListDashboards List all dashboards optionally filtered by folder name. If folderFilters
// is blank, defaults to the configured Monitored folders
func (s *DashNGoImpl) ListDashboards(filterReq filters.V2Filter) []*domain.NestedHit {
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}

	boardLinks := make([]*domain.NestedHit, 0)
	deduplicatedLinks := make(map[int64]*domain.NestedHit)

	var page int64 = 1
	var limit int64 = 5000 // Upper bound of Grafana API call

	tagsParams := make([]string, 0)
	tagExpected := filterReq.GetExpectedValue(filters.TagsFilter)
	if val, ok := tagExpected.([]string); ok {
		tagsParams = append(tagsParams, val...)
	}

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
				lo.Map(pageBoardLinks.GetPayload(), func(item *models.Hit, index int) *domain.NestedHit {
					return &domain.NestedHit{Hit: item}
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

	folderUidMap := getFolderUIDEntityMap(s.ListFolders(nil))
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
		folderMatch = getNestedFolder(folderMatch, link.FolderUID, folderUidMap)
		link.NestedPath = folderMatch

		// accepts all folders if no filter is set
		if s.grafanaConf.GetDashboardSettings().IgnoreFilters && !s.grafanaConf.IsFilterSet() {
			validFolder = true
		} else if filterReq.Validate(filters.FolderFilter, link) /* if no global ignore and filter is set, check folder validity */ {
			validFolder = true
		} else if slices.Contains(s.grafanaConf.GetMonitoredFolders(false), DefaultFolderName) && link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
			validFolder = true
		}

		if !validFolder {
			slog.Debug("Skipping dashboard, as it failed the filter check", "title", link.Title, "folder", link.NestedPath)
			continue
		}

		validUid = filterReq.Validate(filters.DashFilter, link)
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
		boardLinks []*domain.NestedHit
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
		fileName := fmt.Sprintf("%s/%s.json", BuildResourceFolder(link.NestedPath, configDomain.DashboardResource, s.isLocal(), s.globalConf.ClearOutput), metaData.GetPayload().Meta.Slug)
		if err = s.storage.WriteFile(fileName, pretty.Pretty(rawBoard)); err != nil {
			slog.Error("Unable to save dashboard to file\n", "err", err, "dashboard", metaData.GetPayload().Meta.Slug)
		} else {
			boards = append(boards, fileName)
		}

	}
	return boards
}

// getNestedFolder use this if calling from within the service, returns the nested folder path for a given folder
func getNestedFolder(folderTitle, folderUID string, folderUidMap map[string]*domain.NestedHit) string {
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

// TestCreatedFolders  entry point to allow for test to validate behavior independent of any other code path
func (s *DashNGoImpl) TestCreatedFolders(folderName string) (map[string]string, error) {
	return s.createdFolders(folderName)
}

// createFolders Creates a new folder with the given name.  If nested, each sub folder that does not exist is also created
func (s *DashNGoImpl) createdFolders(folderName string) (map[string]string, error) {
	namedUIDMap := getFolderMapping(s.ListFolders(NewFolderFilter()),
		func(db *domain.NestedHit) string {
			return db.NestedPath
		},
		func(fld *domain.NestedHit) *domain.NestedHit { return fld },
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
			if ndx == 0 {
				cnt, pathErr = folderPath.WriteString(folder)
			} else {
				cnt, pathErr = folderPath.WriteString(fmt.Sprintf("%s%s", pathSeparator, folder))
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
func (s *DashNGoImpl) UploadDashboards(filterReq filters.V2Filter) ([]string, error) {
	var (
		rawBoard   []byte
		folderName string
		folderUid  string
		dashFiles  []string
	)
	dashboardPath := config.Config().GetDefaultGrafanaConfig().GetPath(configDomain.DashboardResource, s.grafanaConf.GetOrganizationName())
	filesInDir, err := s.storage.FindAllFiles(dashboardPath, true)
	if err != nil {
		return nil, fmt.Errorf("unable to find any dashFiles to export from storage engine, err: %w", err)
	}
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}
	currentDashboards := s.ListDashboards(filterReq)

	folderUidMap := s.getFolderNameUIDMap(s.ListFolders(NewFolderFilter()))

	alreadyProcessed := make(map[any]bool)

	for _, file := range filesInDir {

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
			return nil, fmt.Errorf("board with same UID was already processed.  Please check your backup folder. This may occur if you pulled the data multiple times with configuration of: nested folder enabled and disabled, uid: %v, title: %v", board["uid"], slug.Make((board["title"]).(string)))
		} else {
			alreadyProcessed[board["uid"]] = true
		}

		// Extract Folder Name based on dashboardPath
		folderName, err = getFolderFromResourcePath(file, configDomain.DashboardResource, s.storage.GetPrefix(), s.grafanaConf.GetOrganizationName())
		if err != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default")
		}

		if folderName == "" {
			folderName = DefaultFolderName
		}
		folderUidMap, err = s.validateDashUploadEntity(filterReq, folderName, &folderUid, folderUidMap, rawBoard)
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

func (s *DashNGoImpl) baseFolderValidation(filterReq filters.V2Filter, folderName string, folderUid *string, folderUidMap map[string]string, rawBoard []byte) (map[string]string, error) {
	// if filter is set or ignore set is not set, apply folder filter, otherwise fall through
	if (s.grafanaConf.IsFilterSet() || !s.grafanaConf.GetDashboardSettings().IgnoreFilters) && !filterReq.Validate(filters.FolderFilter, map[string]any{NestedDashFolderName: folderName}) {
		return folderUidMap, errors.New("dashboard fails to pass folder filter")
	}

	if folderName == DefaultFolderName {
		*folderUid = ""
	} else {
		if val, ok := folderUidMap[folderName]; ok {
			*folderUid = val
		} else {
			newFolders, folderErr := s.createdFolders(folderName)
			if folderErr != nil {
				log.Panic("Unable to create required folder")
			} else {
				maps.Copy(folderUidMap, newFolders)
				*folderUid = folderUidMap[folderName]
			}
		}
	}
	return folderUidMap, nil
}

func (s *DashNGoImpl) validateDashUploadEntity(filterReq filters.V2Filter, folderName string, folderUid *string, folderUidMap map[string]string, rawBoard []byte) (map[string]string, error) {
	if !filterReq.Validate(filters.TagsFilter, rawBoard) {
		return folderUidMap, fmt.Errorf("dashboard fails to pass tag filter: tagFilter: %s", filterReq.GetExpectedString(filters.TagsFilter))
	}

	// always apply filter, ignore filter only applies to folders
	if !filterReq.Validate(filters.DashFilter, rawBoard) {
		return folderUidMap, errors.New("dashboard fails to pass dash filter")
	}

	return s.baseFolderValidation(filterReq, folderName, folderUid, folderUidMap, rawBoard)
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

// DeleteAllDashboards clears all current dashboards being monitored.  Any folder not whitelisted
// will not be affected
func (s *DashNGoImpl) DeleteAllDashboards(filter filters.V2Filter) []string {
	dashboardListing := make([]string, 0)

	items := s.ListDashboards(filter)
	for _, item := range items {
		// if filter.Validate(filters.FolderFilter, item) && filter.Validate(filters.DashFilter, item) {
		err := s.deleteDashboard(item.Hit)
		if err == nil {
			dashboardListing = append(dashboardListing, item.Title)
		} else {
			slog.Warn("Unable to remove dashboard", slog.String("title", item.Title), slog.String("uid", item.UID))
		}
		//}
	}
	return dashboardListing
}
