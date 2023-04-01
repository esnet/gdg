package api

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/api/filters"
	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/gdg/config"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/library_elements"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"os"
	"strings"
)

type LibraryElementsApi interface {
	ListLibraryElements(filter filters.Filter) []*models.LibraryElementDTO
	ListLibraryElementsConnections(filter filters.Filter, connectionID string) []*models.DashboardFullWithMeta
	ImportLibraryElements(filter filters.Filter) []string
	ExportLibraryElements(filter filters.Filter) []string
	DeleteAllLibraryElements(filter filters.Filter) []string
}

var (
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
		dashboard, err := s.getDashboardByUid(filter, item.ConnectionUID)
		if err != nil {
			log.WithField("UID", item.ConnectionUID).Errorf("failed to retrieve linked Dashboard")
		}
		results = append(results, dashboard)
	}

	return results
}

func (s *DashNGoImpl) ListLibraryElements(filter filters.Filter) []*models.LibraryElementDTO {
	ignoreFilters := apphelpers.GetCtxDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters
	folderFilter := NewFolderFilter()
	if ignoreFilters {
		folderFilter = nil
	}

	folderNameMap := getFolderNameIDMap(s.ListFolder(folderFilter))
	values := maps.Values(folderNameMap)
	var buf = strings.Builder{}
	//Check to see if General should be included
	//If Ignore Filters OR General is in monitored list, add 0 folder
	if (!ignoreFilters && slices.Contains(apphelpers.GetCtxDefaultGrafanaConfig().GetMonitoredFolders(), DefaultFolderName)) || ignoreFilters {
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
	params.Kind = &listLibraryPanels
	libraryElements, err := s.client.LibraryElements.GetLibraryElements(params, s.getAuth())
	if err != nil {
		log.WithError(err).Fatal("Unable to list Library Elements")

	}
	return libraryElements.GetPayload().Result.Elements
}

func (s *DashNGoImpl) ImportLibraryElements(filter filters.Filter) []string {
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
			log.Errorf("%s for %s\n", err, item.Name)
			continue
		}
		folderName := DefaultFolderName

		if val, ok := folderMap[item.FolderID]; ok {
			folderName = val
		}

		libraryPath := fmt.Sprintf("%s/%s.json", buildResourceFolder(folderName, config.LibraryElementResource), slug.Make(item.Name))

		if err = s.storage.WriteFile(libraryPath, dsPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err, slug.Make(item.Name))
		} else {
			dataFiles = append(dataFiles, libraryPath)
		}
	}
	return dataFiles
}

func (s *DashNGoImpl) ExportLibraryElements(filter filters.Filter) []string {
	var (
		exported          []string = make([]string, 0)
		rawLibraryElement []byte
		folderName        string
		libraryUID        string
	)

	log.Infof("Reading files from folder: %s", apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.LibraryElementResource))
	filesInDir, err := s.storage.FindAllFiles(apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.LibraryElementResource), true)

	currentLibElements := s.ListLibraryElements(filter)
	libMapping := make(map[string]*models.LibraryElementDTO, 0)
	//Build a mapping by UID
	for ndx, item := range currentLibElements {
		libMapping[item.UID] = currentLibElements[ndx]
	}

	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for library elements")
	}

	for _, file := range filesInDir {
		fileLocation := file
		if strings.HasSuffix(file, ".json") {
			if rawLibraryElement, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file: %s", fileLocation)
				continue
			}

			Results := gjson.GetManyBytes(rawLibraryElement, "meta.folderName", "uid")

			if Results[0].Exists() {
				folderName = Results[0].String()
			} else {
				log.Errorf("Unable to determine folder name of library component, skipping %s", file)
				continue
			}
			//Get UID
			if Results[1].Exists() {
				libraryUID = Results[1].String()
			} else {
				log.Errorf("Unable to determine the library panel UID, %s, attempting to export anyways", file)
				//continue
			}

			if _, ok := libMapping[libraryUID]; ok {
				log.Warnf("Library already exists, skipping %s", file)
				continue
			}

			if !slices.Contains(apphelpers.GetCtxDefaultGrafanaConfig().GetMonitoredFolders(), folderName) {
				log.WithField("folder", folderName).WithField("file", file).Warn("Skipping since requested file is not in a folder gdg is configured to manage")
				continue
			}
			var newLibraryRequest models.CreateLibraryElementCommand

			if err = json.Unmarshal(rawLibraryElement, &newLibraryRequest); err != nil {
				log.WithError(err).Errorf("failed to unmarshall file: %s", fileLocation)
				continue
			}

			params := library_elements.NewCreateLibraryElementParams()
			params.Body = &newLibraryRequest
			entity, err := s.client.LibraryElements.CreateLibraryElement(params, s.getAuth())
			if err != nil {
				log.WithError(err).Errorf("Failed to create library element")
			} else {
				exported = append(exported, entity.Payload.Result.Name)
			}
		}
	}
	return exported
}

func (s *DashNGoImpl) DeleteAllLibraryElements(filter filters.Filter) []string {
	var entries []string
	libraryElements := s.ListLibraryElements(filter)
	for _, element := range libraryElements {

		params := library_elements.NewDeleteLibraryElementByUIDParams()
		params.SetLibraryElementUID(element.UID)
		_, err := s.client.LibraryElements.DeleteLibraryElementByUID(params, s.getAuth())
		if err != nil {
			var logEntry *log.Entry
			if serr, ok := err.(*library_elements.DeleteLibraryElementByUIDForbidden); ok {
				logEntry = log.WithField("ErrorMessage", *serr.GetPayload().Message)
			} else {
				log.WithError(err)
			}
			logEntry.Errorf("Failed to delete library panel titled: %s", element.Name)
			continue
		}
		entries = append(entries, element.Name)
	}

	return entries
}
