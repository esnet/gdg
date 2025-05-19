package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/internal/tools/encode"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/library_elements"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	listLibraryPanels int64 = 1
	listLibraryVars   int64 = 2
)

func (s *DashNGoImpl) ListLibraryElementsConnections(filter filters.Filter, connectionID string) []*models.DashboardFullWithMeta {
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

func (s *DashNGoImpl) ListLibraryElements(filter filters.Filter) []*models.LibraryElementDTO {
	ignoreFilters := s.grafanaConf.GetDashboardSettings().IgnoreFilters
	folderFilter := NewFolderFilter()
	if ignoreFilters {
		folderFilter = nil
		filter = nil
	}

	// folderUidMap := s.getFolderNameUIDMap(s.ListFolders(NewFolderFilter()))
	folderNameMap := getFolderNameIDMap(s.ListFolders(folderFilter))
	values := maps.Values(folderNameMap)
	buf := strings.Builder{}
	// Check to see if General should be included
	// If Ignore Filters OR General is in monitored list, add 0 folder
	if (!ignoreFilters && slices.Contains(s.grafanaConf.GetMonitoredFolders(), DefaultFolderName)) || ignoreFilters {
		buf.WriteString("0,")
	} else {
		buf.WriteString("")
	}
	for _, i := range values {
		buf.WriteString(fmt.Sprintf("%d,", i))
	}
	folderList := buf.String()[:len(buf.String())-1]

	params := library_elements.NewGetLibraryElementsParams()
	params.FolderFilter = &folderList
	params.Kind = ptr.Of(listLibraryPanels)
	libraryElements, err := s.GetClient().LibraryElements.GetLibraryElements(params)
	if err != nil {
		log.Fatalf("Unable to list Library Elements %v", err)
	}
	var data []*models.LibraryElementDTO
	for _, val := range libraryElements.GetPayload().Result.Elements {
		var nestedPath string
		if val.FolderUID == "" {
			nestedPath = DefaultFolderName
		} else {
			fld, err := s.getFolderByUid(val.FolderUID)
			if err != nil {
				slog.Error("unable to get forder to validate resource")
				continue
			}
			nestedPath = fld.NestedPath
		}

		if filter == nil || filter.ValidateAll(map[filters.FilterType]string{filters.FolderFilter: nestedPath}) {
			data = append(data, val)
		}
	}

	return data
}

// DownloadLibraryElements downloads all the Library Elements
func (s *DashNGoImpl) DownloadLibraryElements(filter filters.Filter) []string {
	var (
		listing   []*models.LibraryElementDTO
		dsPacked  []byte
		err       error
		dataFiles []string
	)

	folderMap := reverseLookUp(s.getFolderNameUIDMap(s.ListFolders(nil)))
	listing = s.ListLibraryElements(filter)
	for _, item := range listing {
		if dsPacked, err = json.MarshalIndent(item, "", "	"); err != nil {
			slog.Error("Unable to serialize object", "err", err, "library-element", item.Name)
			continue
		}
		folderName := DefaultFolderName

		if val, ok := folderMap[item.FolderUID]; ok {
			folderName = val
		}

		libraryPath := fmt.Sprintf("%s/%s.json", BuildResourceFolder(folderName, config.LibraryElementResource, s.isLocal(), s.globalConf.ClearOutput), slug.Make(item.Name))

		if err = s.storage.WriteFile(libraryPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "err", err, "library-element", slug.Make(item.Name))
		} else {
			dataFiles = append(dataFiles, libraryPath)
		}
	}
	return dataFiles
}

// UploadLibraryElements uploads all the Library Elements
func (s *DashNGoImpl) UploadLibraryElements(filterReq filters.Filter) []string {
	var (
		exported          []string = make([]string, 0)
		rawLibraryElement []byte
		folderUid         string
		libraryUID        string
		folderName        string
	)

	slog.Info("Reading files from folder", "folder", config.Config().GetDefaultGrafanaConfig().GetPath(config.LibraryElementResource))
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.LibraryElementResource), true)
	if err != nil {
		slog.Error("failed to list files in directory for library elements", "err", err)
	}

	folderUidMap := s.getFolderNameUIDMap(s.ListFolders(NewFolderFilter()))

	_ = folderUidMap
	currentLibElements := s.ListLibraryElements(filterReq)
	libMapping := make(map[string]*models.LibraryElementDTO)
	// Build a mapping by UID
	for ndx, item := range currentLibElements {
		libMapping[item.UID] = currentLibElements[ndx]
	}
	ignoreFilters := s.grafanaConf.GetDashboardSettings().IgnoreFilters

	for _, file := range filesInDir {
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")

		if strings.HasSuffix(file, ".json") {
			if rawLibraryElement, err = s.storage.ReadFile(file); err != nil {
				slog.Error("failed to read file", "file", file, "err", err)
				continue
			}

			// Extract Folder Name based on dashboardPath
			folderName, err = getFolderFromResourcePath(file, config.LibraryElementResource, s.storage.GetPrefix())
			if err != nil {
				slog.Warn("unable to determine dashboard folder name, falling back on default")
			}
			if folderName == "" {
				folderName = DefaultFolderName
			}
			if folderName != DefaultFolderName {
				Results := gjson.GetBytes(rawLibraryElement, "uid")
				if Results.Exists() {
					folderUid = Results.String()
				} else {
					slog.Error("Unable to determine folder name of library component, using default folder", "filename", file)
					folderUid = ""
				}

			} else {
				folderUid = ""
			}
			Results := gjson.GetBytes(rawLibraryElement, "uid")
			// Get UID
			if Results.Exists() {
				libraryUID = Results.String()
			} else {
				slog.Error("Unable to determine the library panel UID, attempting to export anyways", "filename", file)
			}

			if _, ok := libMapping[libraryUID]; ok {
				slog.Warn("Library already exists, skipping", "filename", file)
				continue
			}
			// validateMap := map[filters.FilterType]string{filters.FolderFilter: folderName, filters.DashFilter: baseFile}

			if folderName == DefaultFolderName {
				folderUid = ""
			} else {
				if val, ok := folderUidMap[folderName]; ok {
					// folderId = val
					folderUid = val
				} else {
					// if filterReq.ValidateAll(validateMap) {
					newFolders, folderErr := s.createdFolders(folderName)
					if folderErr != nil {
						log.Panic("Unable to create required folder")
					} else {
						maps.Copy(folderUidMap, newFolders)
						folderUid = folderUidMap[encode.Decode(folderName)]
					}
					//}
				}
			}
			if !ignoreFilters && !validateFolderRegex(s.grafanaConf.GetMonitoredFolders(), folderName) {
				slog.Warn("Skipping since requested file is not in a folder gdg is configured to manage", "folder", folderUid, "file", file)
				continue
			}
			var newLibraryRequest models.CreateLibraryElementCommand
			if err = json.Unmarshal(rawLibraryElement, &newLibraryRequest); err != nil {
				slog.Error("failed to unmarshall file", "filename", file, "err", err)
				continue
			}
			if folderUid != "" {
				newLibraryRequest.FolderUID = folderUid
			}

			entity, err := s.GetClient().LibraryElements.CreateLibraryElement(&newLibraryRequest)
			if err != nil {
				slog.Error("Failed to create library element", "err", err, "resource", file)
			} else {
				exported = append(exported, entity.Payload.Result.Name)
			}
		}
	}
	return exported
}

// DeleteAllLibraryElements deletes all the Library Elements
func (s *DashNGoImpl) DeleteAllLibraryElements(filter filters.Filter) []string {
	var entries []string
	libraryElements := s.ListLibraryElements(filter)
	for _, element := range libraryElements {

		_, err := s.GetClient().LibraryElements.DeleteLibraryElementByUID(element.UID)
		if err != nil {
			logEntries := make([]any, 0)
			var serr *library_elements.DeleteLibraryElementByUIDForbidden
			if errors.As(err, &serr) {
				logEntries = append(logEntries, []any{"ErrorMessage", *serr.GetPayload().Message}...)
			}

			logEntries = append(logEntries, []any{"panel", element.Name}...)
			slog.Error("Failed to delete library panel", logEntries...)
			continue
		}
		entries = append(entries, element.Name)
	}

	return entries
}
