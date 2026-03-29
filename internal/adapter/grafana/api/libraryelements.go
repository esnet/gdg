package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"strings"

	"github.com/esnet/gdg/internal/adapter/filters/v2"
	"github.com/esnet/gdg/internal/adapter/grafana"
	"github.com/esnet/gdg/internal/adapter/grafana/common"
	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/tools"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/library_elements"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/maps"
)

const (
	listLibraryPanels int64 = 1
	listLibraryVars   int64 = 2
)

func setupLibElementsReaders(filterObj ports.Filter) {
	err := filterObj.RegisterReader(reflect.TypeFor[*domain.WithNested[models.LibraryElementDTO]](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
		val, ok := a.(*domain.WithNested[models.LibraryElementDTO])
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		_ = val
		switch filterType {
		case domain.FolderFilter:
			return val.NestedPath, nil
		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Library Elements Filter, obj entity filter failure, aborting.")
	}
	err = filterObj.RegisterReader(reflect.TypeFor[[]byte](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.FolderFilter:
			{
				foo := string(val)
				_ = foo
				r := gjson.GetBytes(val, "meta.folderName")
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
		log.Fatalf("Unable to create a valid Library Elements Filter, json filter failure, aborting.")
	}
	err = filterObj.RegisterReader(reflect.TypeFor[map[string]any](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
		val, ok := a.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.FolderFilter:
			{
				return val[NestedDashFolderName], nil
			}
		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Library Elements Filter, map filter failure, aborting.")
	}
}

func NewLibraryElementFilter(cfg *configDomain.GDGAppConfiguration) ports.Filter {
	filterObj := v2.NewBaseFilter()
	setupLibElementsReaders(filterObj)
	addFolderFilter(cfg, filterObj, "")

	return filterObj
}

func (s *DashNGoImpl) ListLibraryElementsConnections(filter ports.Filter, connectionID string) []*models.DashboardFullWithMeta {
	payload, err := s.GetClient().LibraryElements.GetLibraryElementConnections(connectionID)
	if err != nil {
		log.Fatalf("unable to retrieve a valid connection for %s", connectionID)
	}
	var results []*models.DashboardFullWithMeta

	for _, item := range payload.GetPayload().Result {
		dashboard, err := s.getDashboardByUid(item.ConnectionUID)
		if err != nil {
			slog.Error("failed to retrieve linked Dashboard", "uid", item.ConnectionUID)
		}
		results = append(results, dashboard)
	}

	return results
}

func (s *DashNGoImpl) ListLibraryElements(filter ports.Filter) []*domain.WithNested[models.LibraryElementDTO] {
	const limit int64 = 100
	var (
		page        int64 = 1
		allElements []*models.LibraryElementDTO
	)
	ignoreFilters := s.grafanaConf.GetDashboardSettings().IgnoreFilters
	if ignoreFilters {
		filter = nil
	} else if filter == nil {
		filter = NewLibraryElementFilter(s.gdgConfig)
	}

	// Fetch all lib elements
	for {
		params := library_elements.NewGetLibraryElementsParams()
		params.Kind = new(listLibraryPanels)
		params.Page = &page
		params.PerPage = new(limit)

		libraryElements, err := s.GetClient().LibraryElements.GetLibraryElements(params)
		if err != nil {
			log.Fatalf("Unable to list Library Elements %v", err)
		}
		allElements = append(allElements, libraryElements.GetPayload().Result.Elements...)
		if int64(len(libraryElements.GetPayload().Result.Elements)) < limit {
			break
		}
		page += 1
	}

	var newData []*domain.WithNested[models.LibraryElementDTO]
	for _, val := range allElements {
		var nestedPath string
		if val.FolderUID == "" {
			nestedPath = common.DefaultFolderName
		} else {
			fld, err := s.getFolderByUid(val.FolderUID)
			if err != nil {
				slog.Error("unable to get forder to validate resource")
				continue
			}
			nestedPath = fld.NestedPath
		}

		if ignoreFilters || filter.ValidateAll(context.Background(), map[string]any{NestedDashFolderName: nestedPath}) {
			newData = append(newData, &domain.WithNested[models.LibraryElementDTO]{
				Entity:     val,
				NestedPath: nestedPath,
			})
		}
	}

	return newData
}

// DownloadLibraryElements downloads all the Library Elements
func (s *DashNGoImpl) DownloadLibraryElements(filter ports.Filter) []string {
	var (
		listing   []*domain.WithNested[models.LibraryElementDTO]
		dsPacked  []byte
		err       error
		dataFiles []string
	)

	folderMap := tools.ReverseLookUp(s.getFolderNameUIDMap(s.ListFolders(nil)))
	listing = s.ListLibraryElements(filter)
	for _, item := range listing {
		if dsPacked, err = json.MarshalIndent(item, "", "	"); err != nil {
			slog.Error("Unable to serialize object", "err", err, "library-element", item.Entity.Name)
			continue
		}
		folderName := common.DefaultFolderName

		if val, ok := folderMap[item.Entity.FolderUID]; ok {
			folderName = val
		}

		libraryPath := fmt.Sprintf("%s/%s.json", resources.BuildResourceFolder(s.grafanaConf, folderName, domain.LibraryElementResource, s.isLocal(), s.GetGlobals().ClearOutput), slug.Make(item.Entity.Name))

		if err = s.storage.WriteFile(libraryPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "err", err, "library-element", slug.Make(item.Entity.Name))
		} else {
			dataFiles = append(dataFiles, libraryPath)
		}
	}
	return dataFiles
}

// UploadLibraryElements uploads all the Library Elements
func (s *DashNGoImpl) UploadLibraryElements(filterReq ports.Filter) []string {
	var (
		exported          = make([]string, 0)
		rawLibraryElement []byte
		folderUid         string
		libraryUID        string
		folderName        string
	)

	orgName := s.grafanaConf.GetOrganizationName()
	slog.Info("Reading files from folder", "folder", s.grafanaConf.GetPath(domain.LibraryElementResource, orgName))
	filesInDir, err := s.storage.FindAllFiles(s.grafanaConf.GetPath(domain.LibraryElementResource, orgName), true)
	if err != nil {
		slog.Error("failed to list files in directory for library elements", "err", err)
	}

	folderUidMap := s.getFolderNameUIDMap(s.ListFolders(nil))
	currentLibElements := s.ListLibraryElements(filterReq)
	libMapping := make(map[string]*domain.WithNested[models.LibraryElementDTO])
	// Build a mapping by UID
	for ndx, item := range currentLibElements {
		libMapping[item.Entity.UID] = currentLibElements[ndx]
	}
	ignoreFilters := s.grafanaConf.GetDashboardSettings().IgnoreFilters

	for _, file := range filesInDir {
		if !strings.HasSuffix(file, ".json") {
			slog.Debug("Skipping file with wrong extension", "file", file)
			continue
		}

		if rawLibraryElement, err = s.storage.ReadFile(file); err != nil {
			slog.Error("failed to read file", "file", file, "err", err)
			continue
		}

		// Extract Folder Name based on dashboardPath
		folderName, err = getFolderFromResourcePath(s.grafanaConf, file, domain.LibraryElementResource, s.storage.GetPrefix(), s.grafanaConf.GetOrganizationName())
		if err != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default")
			folderName = common.DefaultFolderName
		}
		if !ignoreFilters && !filterReq.Validate(context.Background(), domain.FolderFilter, map[string]any{NestedDashFolderName: folderName}) {
			slog.Warn("Skipping since requested file is not in a folder gdg is configured to manage", "folder", folderName, "file", file)
			continue
		}
		if folderName != common.DefaultFolderName {
			Results := gjson.GetBytes(rawLibraryElement, "Entity.folderUid")
			if Results.Exists() {
				folderUid = Results.String()
			} else {
				slog.Error("Unable to determine folder uid of library component, using default folder", "filename", file)
				folderUid = ""
			}
		} else {
			folderUid = ""
		}
		Results := gjson.GetBytes(rawLibraryElement, "Entity.uid")
		// Get UID
		if Results.Exists() {
			libraryUID = Results.String()
		} else {
			slog.Warn("Unable to determine the library panel UID, attempting to export anyways", "filename", file)
		}

		if _, ok := libMapping[libraryUID]; ok {
			slog.Warn("Library already exists, skipping", "filename", file)
			continue
		}
		if folderName == "" {
			folderName = common.DefaultFolderName
		}

		if folderName != common.DefaultFolderName {
			if val, ok := folderUidMap[folderName]; ok {
				folderUid = val
			} else {

				newFolders, folderErr := s.createdFolders(folderName)
				if folderErr != nil {
					log.Panic("Unable to create required folder")
				} else {
					maps.Copy(folderUidMap, newFolders)
					folderUid = folderUidMap[folderName]
				}

			}
		}

		var libraryRequest domain.WithNested[*models.LibraryElementDTO]
		if err = json.Unmarshal(rawLibraryElement, &libraryRequest); err != nil {
			slog.Error("failed to unmarshall file", "filename", file, "err", err)
			continue
		}
		newLibraryRequest := grafana.WithNestedToCreateLibraryElement(libraryRequest)
		if folderUid != "" {
			newLibraryRequest.FolderUID = folderUid
		}

		entity, grafanaErr := s.GetClient().LibraryElements.CreateLibraryElement(newLibraryRequest)
		if grafanaErr != nil {
			slog.Error("Failed to create library element", "err", grafanaErr, "resource", file)
		} else {
			exported = append(exported, fmt.Sprintf("%s/%s", folderName, entity.Payload.Result.Name))
		}
	}
	return exported
}

// DeleteAllLibraryElements deletes all the Library Elements
func (s *DashNGoImpl) DeleteAllLibraryElements(filter ports.Filter) []string {
	var entries []string
	libraryElements := s.ListLibraryElements(filter)
	for _, element := range libraryElements {

		_, err := s.GetClient().LibraryElements.DeleteLibraryElementByUID(element.Entity.UID)
		if err != nil {
			logEntries := make([]any, 0)
			var serr *library_elements.DeleteLibraryElementByUIDForbidden
			if errors.As(err, &serr) {
				logEntries = append(logEntries, []any{"ErrorMessage", *serr.GetPayload().Message}...)
			}

			logEntries = append(logEntries, []any{"panel", element.Entity.Name}...)
			slog.Error("Failed to delete library panel", logEntries...)
			continue
		}
		entries = append(entries, element.Entity.Name)
	}

	return entries
}
