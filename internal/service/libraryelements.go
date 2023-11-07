package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/tools"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/library_elements"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/gosimple/slug"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"log"
	"log/slog"
	"strings"
)

type LibraryElementsApi interface {
	ListLibraryElements(filter filters.Filter) []*models.LibraryElementDTO
	ListLibraryElementsConnections(filter filters.Filter, connectionID string) []*models.DashboardFullWithMeta
	DownloadLibraryElements(filter filters.Filter) []string
	UploadLibraryElements(filter filters.Filter) []string
	DeleteAllLibraryElements(filter filters.Filter) []string
}

const (
	listLibraryPanels int64 = 1
	listLibraryVars   int64 = 2
)

func (s *DashNGoImpl) ListLibraryElementsConnections(filter filters.Filter, connectionID string) []*models.DashboardFullWithMeta {
	params := library_elements.NewGetLibraryElementConnectionsParams()
	params.SetLibraryElementUID(connectionID)
	payload, err := s.client.LibraryElements.GetLibraryElementConnections(params, s.getAuth())
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
	ignoreFilters := config.Config().GetDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters
	folderFilter := NewFolderFilter()
	if ignoreFilters {
		folderFilter = nil
	}

	folderNameMap := getFolderNameIDMap(s.ListFolder(folderFilter))
	values := maps.Values(folderNameMap)
	var buf = strings.Builder{}
	//Check to see if General should be included
	//If Ignore Filters OR General is in monitored list, add 0 folder
	if (!ignoreFilters && slices.Contains(config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders(), DefaultFolderName)) || ignoreFilters {
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
	params.Kind = tools.PtrOf(listLibraryPanels)
	libraryElements, err := s.client.LibraryElements.GetLibraryElements(params, s.getAuth())
	if err != nil {
		log.Fatalf("Unable to list Library Elements %v", err)

	}
	return libraryElements.GetPayload().Result.Elements
}

// DownloadLibraryElements downloads all the Library Elements
func (s *DashNGoImpl) DownloadLibraryElements(filter filters.Filter) []string {
	var (
		listing   []*models.LibraryElementDTO
		dsPacked  []byte
		err       error
		dataFiles []string
	)

	folderMap := reverseLookUp(getFolderNameIDMap(s.ListFolder(nil)))
	listing = s.ListLibraryElements(filter)
	for _, item := range listing {
		if dsPacked, err = json.MarshalIndent(item, "", "	"); err != nil {
			slog.Error("Unable to serialize object", "err", err, "library-element", item.Name)
			continue
		}
		folderName := DefaultFolderName

		if val, ok := folderMap[item.FolderID]; ok {
			folderName = val
		}

		libraryPath := fmt.Sprintf("%s/%s.json", buildResourceFolder(folderName, config.LibraryElementResource), slug.Make(item.Name))

		if err = s.storage.WriteFile(libraryPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "err", err, "library-element", slug.Make(item.Name))
		} else {
			dataFiles = append(dataFiles, libraryPath)
		}
	}
	return dataFiles
}

// UploadLibraryElements uploads all the Library Elements
func (s *DashNGoImpl) UploadLibraryElements(filter filters.Filter) []string {
	var (
		exported          []string = make([]string, 0)
		rawLibraryElement []byte
		folderName        string
		libraryUID        string
	)

	slog.Info("Reading files from folder", "folder", config.Config().GetDefaultGrafanaConfig().GetPath(config.LibraryElementResource))
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.LibraryElementResource), true)

	currentLibElements := s.ListLibraryElements(filter)
	libMapping := make(map[string]*models.LibraryElementDTO, 0)
	//Build a mapping by UID
	for ndx, item := range currentLibElements {
		libMapping[item.UID] = currentLibElements[ndx]
	}

	if err != nil {
		slog.Error("failed to list files in directory for library elements", "err", err)
	}

	for _, file := range filesInDir {
		fileLocation := file
		if strings.HasSuffix(file, ".json") {
			if rawLibraryElement, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "file", fileLocation, "err", err)
				continue
			}

			Results := gjson.GetManyBytes(rawLibraryElement, "meta.folderName", "uid")

			if Results[0].Exists() {
				folderName = Results[0].String()
			} else {
				slog.Error("Unable to determine folder name of library component, skipping.", "filename", file)
				continue
			}
			//Get UID
			if Results[1].Exists() {
				libraryUID = Results[1].String()
			} else {
				slog.Error("Unable to determine the library panel UID, attempting to export anyways", "filename", file)
			}

			if _, ok := libMapping[libraryUID]; ok {
				slog.Warn("Library already exists, skipping", "filename", file)
				continue
			}

			if !slices.Contains(config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders(), folderName) {
				slog.Warn("Skipping since requested file is not in a folder gdg is configured to manage", "folder", folderName, "file", file)
				continue
			}
			var newLibraryRequest models.CreateLibraryElementCommand

			if err = json.Unmarshal(rawLibraryElement, &newLibraryRequest); err != nil {
				slog.Error("failed to unmarshall file", "filename", fileLocation, "err", err)
				continue
			}

			params := library_elements.NewCreateLibraryElementParams()
			params.Body = &newLibraryRequest
			entity, err := s.client.LibraryElements.CreateLibraryElement(params, s.getAuth())
			if err != nil {
				slog.Error("Failed to create library element", "err", err)
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

		params := library_elements.NewDeleteLibraryElementByUIDParams()
		params.SetLibraryElementUID(element.UID)
		_, err := s.client.LibraryElements.DeleteLibraryElementByUID(params, s.getAuth())
		if err != nil {
			logEntries := make([]interface{}, 0)
			var serr *library_elements.DeleteLibraryElementByUIDForbidden
			if errors.As(err, &serr) {
				logEntries = append(logEntries, []interface{}{"ErrorMessage", *serr.GetPayload().Message}...)
			}

			logEntries = append(logEntries, []interface{}{"panel", element.Name}...)
			slog.Error("Failed to delete library panel", logEntries...)
			continue
		}
		entries = append(entries, element.Name)
	}

	return entries
}
