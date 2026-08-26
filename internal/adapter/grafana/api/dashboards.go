package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"sort"

	"github.com/esnet/gdg/internal/adapter/filters/v2"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/encode"
	"github.com/tidwall/gjson"

	configDomain "github.com/esnet/gdg/internal/config/config_domain"

	"github.com/grafana/grafana-openapi-client-go/client/dashboards"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/tidwall/pretty"
)

const (
	NestedDashFolderName = "NestedDashFolderName"
)

func setupDashReaders(filterObj outbound.Filter) {
	err := v2.RegisterTypedReader[*domain.NestedHit](filterObj, func(ctx context.Context, filterType domain.FilterType, val *domain.NestedHit) (any, error) {
		switch filterType {
		case domain.FolderFilter:
			return val.NestedPath, nil
		case domain.TagsFilter:
			return val.Tags, nil
		case domain.DashFilter:
			return val.UID, nil
		default:
			return nil, fmt.Errorf("unsupported filter type: %s", filterType)
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, object reader could not be created, aborting.")
	}
	err = v2.RegisterTypedReader[[]byte](filterObj, func(ctx context.Context, filterType domain.FilterType, val []byte) (any, error) {
		switch filterType {
		case domain.TagsFilter:
			// Try top-level "tags" first, then the legacy Grafana envelope "dashboard.tags".
			r := gjson.GetBytes(val, "tags")
			if !r.Exists() || !r.IsArray() {
				r = gjson.GetBytes(val, "dashboard.tags")
			}
			if !r.Exists() || !r.IsArray() {
				return []string{}, nil
			}
			return lo.Map(r.Array(), func(item gjson.Result, _ int) string {
				return item.String()
			}), nil
		case domain.DashFilter:
			// Try top-level "uid" first (v1 raw JSON and v2 filterJSON), then the
			// legacy Grafana envelope "dashboard.uid", and finally "resource.metadata.name"
			// for DashboardV2Gdg files stored on disk.
			// Return an empty string (not an error) when no UID is found so the validation
			// function can decide: if no dash filter is configured it passes via its own
			// `val == "" || exp == ""` short-circuit; if a filter IS configured the empty
			// string correctly fails to match any UID.
			r := gjson.GetBytes(val, "uid")
			if !r.Exists() || r.String() == "" {
				r = gjson.GetBytes(val, "dashboard.uid")
			}
			if !r.Exists() || r.String() == "" {
				r = gjson.GetBytes(val, "resource.metadata.name")
			}
			if !r.Exists() || r.String() == "" {
				return "", nil
			}
			return r.String(), nil
		default:
			return nil, fmt.Errorf("unsupported filter type: %s", filterType)
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, json reader could not be created, aborting.")
	}
	err = v2.RegisterTypedReader[map[string]any](filterObj, func(ctx context.Context, filterType domain.FilterType, val map[string]any) (any, error) {
		switch filterType {
		case domain.FolderFilter:
			return val[NestedDashFolderName], nil
		default:
			return nil, fmt.Errorf("unsupported filter type: %s", filterType)
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, map entity reader could not be created, aborting.")
	}
}

func addFolderFilter(cfg *configDomain.GDGAppConfiguration, filterReq outbound.Filter, folderFilter string) {
	var folderArr []string
	if folderFilter != "" {
		cfg.GetDefaultGrafanaConfig().SetFilterFolder(folderFilter)
		folderArr = []string{folderFilter}
	} else {
		cfg.GetDefaultGrafanaConfig().ClearFilters()
		folderArr = cfg.GetDefaultGrafanaConfig().GetMonitoredFolders(false)
	}
	filterReq.AddValidation(domain.FolderFilter, func(ctx context.Context, value any, expected any) error {
		val, expressions, convErr := v2.GetMismatchParams[string, []string](value, expected, domain.FolderFilter)
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

func NewDashboardFilter(cfg *configDomain.GDGAppConfiguration, entries ...string) outbound.Filter {
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
	}
	filterObj := v2.NewBaseFilter()
	setupDashReaders(filterObj)
	// Setup Readers

	err := filterObj.RegisterDataProcessor(domain.FolderFilter, v2.FolderQuoteRegExProcessor)
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, aborting.")
	}

	addFolderFilter(cfg, filterObj, folderFilter)

	v2.RegisterTypedValidation[string](filterObj, domain.DashFilter, dashboardFilter, func(ctx context.Context, val, exp string) error {
		if val == "" || exp == "" {
			return nil
		}
		if exp != val {
			return fmt.Errorf("failed validation test val:%s  expected: %s", val, exp)
		}
		return nil
	})

	v2.RegisterTypedValidation[[]string](filterObj, domain.TagsFilter, tagObj, func(ctx context.Context, val, exp []string) error {
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
	})

	return filterObj
}

// listDashboardsV1 lists all dashboards via the legacy /api/dashboards endpoint.
// It is an unexported adapter-layer helper called by DashboardServiceImpl.
func (d *DashboardServiceImpl) listDashboardsV1(filterReq outbound.Filter) []*domain.NestedHit {
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter(d.gdgConfig, "", "", "")
	}

	boardLinks := make([]*domain.NestedHit, 0)
	deduplicatedLinks := make(map[int64]*domain.NestedHit)

	var page int64 = 1
	var limit int64 = 5000 // Upper bound of Grafana API call

	tagsParams := make([]string, 0)
	tagExpected := filterReq.GetExpectedValue(context.Background(), domain.TagsFilter)
	if val, ok := tagExpected.([]string); ok {
		tagsParams = append(tagsParams, val...)
	}

	retrieve := func(tag string) {
		for {
			searchParams := search.NewSearchParams()
			if tag != "" {
				searchParams.Tag = []string{tag}
			}
			searchParams.Limit = new(limit)
			searchParams.Page = new(page)
			searchParams.Type = new(domain.ApiConsts.SearchTypeDashboard)

			pageBoardLinks, err := d.GetClient().Search.Search(searchParams)
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

	folderUidMap := d.getFolderUIDEntityMap(nil)
	var validFolder bool
	var validUid bool
	for ndx, link := range boardLinks {
		link.Slug = d.resources.UpdateSlug(link.URI)
		_, ok := deduplicatedLinks[link.ID]
		if ok {
			slog.Debug("duplicate board, skipping ")
			continue
		}
		validFolder = false
		folderMatch := link.FolderTitle
		if folderMatch == "" {
			folderMatch = domain.ApiConsts.DefaultFolderName
		}
		folderMatch = getNestedFolder(folderMatch, link.FolderUID, folderUidMap)
		link.NestedPath = folderMatch

		// accepts all folders if no filter is set
		if d.grafanaConf.GetDashboardSettings().IgnoreFilters && !d.grafanaConf.IsFilterSet() {
			validFolder = true
		} else if filterReq.Validate(context.Background(), domain.FolderFilter, link) /* if no global ignore and filter is set, check folder validity */ {
			validFolder = true
		} else if slices.Contains(d.grafanaConf.GetMonitoredFolders(false), domain.ApiConsts.DefaultFolderName) && link.FolderID == 0 {
			link.FolderTitle = domain.ApiConsts.DefaultFolderName
			validFolder = true
		}

		if !validFolder {
			slog.Debug("Skipping dashboard, as it failed the filter check", "title", link.Title, "folder", link.NestedPath)
			continue
		}

		validUid = filterReq.Validate(context.Background(), domain.DashFilter, link)
		if link.FolderID == 0 && string(link.Type) == domain.ApiConsts.SearchTypeDashboard {
			link.FolderTitle = domain.ApiConsts.DefaultFolderName
		}
		// check folder

		if validUid {
			deduplicatedLinks[link.ID] = boardLinks[ndx]
		}
	}

	boardLinks = slices.Collect(maps.Values(deduplicatedLinks))
	sort.Slice(boardLinks, func(i, j int) bool {
		return boardLinks[i].ID < boardLinks[j].ID
	})

	return boardLinks
}

// downloadDashboardsV1 saves all dashboards via the legacy /api/dashboards endpoint.
// It is an unexported adapter-layer helper called by DashboardServiceImpl.
func (d *DashboardServiceImpl) downloadDashboardsV1(filter outbound.Filter) []string {
	var (
		boardLinks []*domain.NestedHit
		rawBoard   []byte
		err        error
		metaData   *dashboards.GetDashboardByUIDOK
	)

	boardLinks = d.listDashboardsV1(filter)

	var boards []string
	for _, link := range boardLinks {
		if string(link.Type) != domain.ApiConsts.SearchTypeDashboard {
			slog.Debug("Ignoring dashboard-folder", "folder", link.Title)
			continue
		}

		if metaData, err = d.GetClient().Dashboards.GetDashboardByUID(link.UID); err != nil {
			slog.Error("unable to get Dashboard by UID", "err", err, "Dashboard-URI", link.URI)
			continue
		}

		rawBoard, err = json.Marshal(metaData.GetPayload().Dashboard)
		if err != nil {
			slog.Error("unable to serialize dashboard", "dashboard", link.UID)
			continue
		}

		fileName := fmt.Sprintf("%s/%s.json", d.resources.BuildResourceFolder(d.grafanaConf, link.NestedPath, domain.DashboardResource, d.isLocal(), d.GetGlobals().ClearOutput), metaData.GetPayload().Meta.Slug)
		if err = d.storage.WriteFile(fileName, pretty.Pretty(rawBoard)); err != nil {
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

// createdFolders and createdFoldersWithBaseUID are promoted from baseService.

// UploadDashboards finds all the dashboards in the configured location and exports them to grafana.
// if the folder doesn't exist, it'll be created.
// uploadDashboardsV1 uploads dashboards via the legacy /api/dashboards endpoint.
// It is an unexported adapter-layer helper called by uploadDashboardsMixed, which
// pre-reads all files and passes only v1-format entries via the files map
// (path → raw bytes), eliminating redundant storage calls in this function.
// Pruning is handled by the caller via pruneDashboards.
func (d *DashboardServiceImpl) uploadDashboardsV1(filterReq outbound.Filter, files map[string][]byte) ([]string, error) {
	var (
		folderName string
		folderUid  string
		dashFiles  []string
	)
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter(d.gdgConfig, "", "", "")
	}

	folderUidMap := d.getFolderNameUIDMap(d.listFolders(NewFolderFilter(d.gdgConfig)))

	for file, rawBoard := range files {
		board := make(map[string]any)
		if err := json.Unmarshal(rawBoard, &board); err != nil {
			slog.Warn("Failed to unmarshall file", "filename", file)
			continue
		}

		// Extract Folder Name based on dashboardPath
		var folderErr error
		folderName, folderErr = d.resources.GetFolderFromResourcePath(d.grafanaConf, file, domain.DashboardResource, d.storage.GetPrefix(), d.grafanaConf.GetOrganizationName())
		if folderErr != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default")
		}

		if folderName == "" {
			folderName = domain.ApiConsts.DefaultFolderName
		}
		var validateErr error
		folderUidMap, validateErr = d.validateDashUploadEntity(filterReq, folderName, &folderUid, folderUidMap, rawBoard)
		if validateErr != nil {
			slog.Warn("validation failed, skipping", "file", file, "err", validateErr)
			continue
		}

		// zero out ID.  Can't create a new dashboard if an ID already exists.
		delete(board, "id")
		importDashReq := &models.ImportDashboardRequest{
			FolderUID: folderUid,
			Overwrite: true,
			Dashboard: board,
		}

		if _, exportError := d.GetClient().Dashboards.ImportDashboard(importDashReq); exportError != nil {
			slog.Info("error on Exporting dashboard", "dashboard-filename", file, "err", exportError)
			continue
		}
		dashFiles = append(dashFiles, file)
	}

	return dashFiles, nil
}

// baseFolderValidation validates the folder filter and resolves (or creates) the
// folder, returning the folder UID and the updated folderUidMap. The rawBoard
// parameter has been intentionally removed — tag/dash filter checks belong in
// validateDashUploadEntity which calls this function.
func (d *DashboardServiceImpl) baseFolderValidation(filterReq outbound.Filter, folderName string, folderUidMap map[string]string) (folderUID string, updatedMap map[string]string, err error) {
	if (d.grafanaConf.IsFilterSet() || !d.grafanaConf.GetDashboardSettings().IgnoreFilters) && !filterReq.Validate(context.Background(), domain.FolderFilter, map[string]any{NestedDashFolderName: folderName}) {
		return "", folderUidMap, errors.New("dashboard fails to pass folder filter")
	}

	if folderName == domain.ApiConsts.DefaultFolderName {
		return "", folderUidMap, nil
	}

	if val, ok := folderUidMap[folderName]; ok {
		return val, folderUidMap, nil
	}

	newFolders, folderErr := d.createdFolders(folderName)
	if folderErr != nil {
		log.Panic("Unable to create required folder")
	}
	maps.Copy(folderUidMap, newFolders)
	return folderUidMap[folderName], folderUidMap, nil
}

func (d *DashboardServiceImpl) validateDashUploadEntity(filterReq outbound.Filter, folderName string, folderUid *string, folderUidMap map[string]string, rawBoard []byte) (map[string]string, error) {
	if !filterReq.Validate(context.Background(), domain.TagsFilter, rawBoard) {
		return folderUidMap, fmt.Errorf("dashboard fails to pass tag filter: tagFilter: %s", filterReq.GetExpectedString(context.Background(), domain.TagsFilter))
	}

	// always apply filter, ignore filter only applies to folders
	if !filterReq.Validate(context.Background(), domain.DashFilter, rawBoard) {
		return folderUidMap, errors.New("dashboard fails to pass dash filter")
	}

	uid, updatedMap, err := d.baseFolderValidation(filterReq, folderName, folderUidMap)
	if err != nil {
		return folderUidMap, err
	}
	*folderUid = uid
	return updatedMap, nil
}

// deleteDashboard removes a dashboard from grafana.  If the dashboard doesn't exist,
// an error is returned.
//
// Parameters:
// item - dashboard to be deleted
//
// Returns:
// error - error returned from the grafana API
func (d *DashboardServiceImpl) deleteDashboard(item *models.Hit) error {
	_, err := d.GetClient().Dashboards.DeleteDashboardByUID(item.UID)
	return err
}

// deleteAllDashboardsV1 removes all monitored dashboards via the legacy /api/dashboards endpoint.
// It is an unexported adapter-layer helper called by DashboardServiceImpl.
func (d *DashboardServiceImpl) deleteAllDashboardsV1(filter outbound.Filter) []string {
	dashboardListing := make([]string, 0)

	items := d.listDashboardsV1(filter)
	for _, item := range items {
		// if filter.Validate(filters.FolderFilter, item) && filter.Validate(filters.DashFilter, item) {
		err := d.deleteDashboard(item.Hit)
		if err == nil {
			dashboardListing = append(dashboardListing, item.Title)
		} else {
			slog.Warn("Unable to remove dashboard", slog.String("title", item.Title), slog.String("uid", item.UID))
		}
		//}
	}
	return dashboardListing
}
