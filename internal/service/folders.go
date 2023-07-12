package service

import (
	"encoding/json"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/folders"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/search"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"strings"
)

// FoldersApi Contract definition
type FoldersApi interface {
	ListFolder(filter filters.Filter) []*models.Hit
	ImportFolder(filter filters.Filter) []string
	ExportFolder(filter filters.Filter) []string
	DeleteAllFolder(filter filters.Filter) []string
}

func NewFolderFilter() filters.Filter {
	filterObj := filters.NewBaseFilter()
	filterObj.AddValidation(filters.FolderFilter, func(i interface{}) bool {
		val, ok := i.(map[filters.FilterType]string)
		if !ok {
			return ok
		}
		//Check folder
		if folderFilter, ok := val[filters.FolderFilter]; ok {
			return slices.Contains(config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders(), folderFilter)
		} else {
			return true
		}
	})
	return filterObj

}

func (s *DashNGoImpl) ListFolder(filter filters.Filter) []*models.Hit {
	var result = make([]*models.Hit, 0)
	if config.Config().GetDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
		filter = nil
	}
	p := search.NewSearchParams()
	p.Type = &searchTypeFolder
	folderListing, err := s.client.Search.Search(p, s.getAuth())
	folderListing.GetPayload()
	if err != nil {
		log.Fatal("unable to retrieve folder list.")
	}
	for ndx, val := range folderListing.GetPayload() {
		if filter == nil {
			result = append(result, folderListing.GetPayload()[ndx])
		} else if filter.ValidateAll(map[filters.FilterType]string{filters.FolderFilter: val.Title}) {
			result = append(result, folderListing.GetPayload()[ndx])
		}
	}

	return result

}

// ImportFolder
func (s *DashNGoImpl) ImportFolder(filter filters.Filter) []string {
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	folderListing := s.ListFolder(filter)
	for _, folder := range folderListing {
		if dsPacked, err = json.MarshalIndent(folder, "", "	"); err != nil {
			log.Errorf("%s for %s\n", err, folder.Title)
			continue
		}
		dsPath := buildResourcePath(slug.Make(folder.Title), config.FolderResource)
		if err = s.storage.WriteFile(dsPath, dsPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err.Error(), slug.Make(folder.Title))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles
}

func (s *DashNGoImpl) ExportFolder(filter filters.Filter) []string {
	var (
		result    []string
		rawFolder []byte
	)
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.FolderResource), false)
	if err != nil {
		log.WithError(err).Fatal("Failed to read folders imports")
	}
	folderItems := s.ListFolder(filter)

	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.FolderResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file %s", fileLocation)
				continue
			}
		}
		var newFolder models.CreateFolderCommand
		if err = json.Unmarshal(rawFolder, &newFolder); err != nil {
			log.WithError(err).Warn("failed to unmarshall folder")
			continue
		}
		skipCreate := false
		for _, existingFolder := range folderItems {
			if existingFolder.UID == newFolder.UID {
				log.Warnf("Folder '%s' already exists, skipping", existingFolder.Title)
				skipCreate = true
			}

		}
		if skipCreate {
			continue
		}
		params := folders.NewCreateFolderParams()
		params.Body = &newFolder
		f, err := s.client.Folders.CreateFolder(params, s.getAuth())
		if err != nil {
			log.Errorf("failed to create folder %s", newFolder.Title)
			continue
		}
		result = append(result, f.Payload.Title)

	}
	return result
}

func (s *DashNGoImpl) DeleteAllFolder(filter filters.Filter) []string {
	var result []string
	folderListing := s.ListFolder(filter)
	for _, folder := range folderListing {
		params := folders.NewDeleteFolderParams()
		params.FolderUID = folder.UID
		_, err := s.client.Folders.DeleteFolder(params, s.getAuth())
		if err == nil {
			result = append(result, folder.Title)
		}
	}
	return result
}

// getFolderNameIDMap helper function to build a mapping for name to folderID
func getFolderNameIDMap(folders []*models.Hit) map[string]int64 {
	var folderMap = make(map[string]int64)
	for _, folder := range folders {
		folderMap[folder.Title] = folder.ID
	}
	return folderMap
}

func reverseLookUp[T comparable, Y comparable](m map[T]Y) map[Y]T {
	reverse := make(map[Y]T, 0)
	for key, val := range m {
		reverse[val] = key
	}

	return reverse
}
