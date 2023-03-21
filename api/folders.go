package api

import (
	"encoding/json"
	"github.com/esnet/gdg/config"
	"github.com/gosimple/slug"
	gclient "github.com/grafana/grafana-api-golang-client"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

func (s *DashNGoImpl) ListFolder(filter Filter) []gclient.Folder {
	folders, err := s.client.Folders()
	if err != nil {
		log.WithError(err).Fatal("Failed to retrieve folders")
	}
	return folders

}
func (s *DashNGoImpl) ImportFolder(filter Filter) []string {
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	folders, err := s.client.Folders()
	if err != nil {
		log.WithError(err).Fatal("failed to list current folders")
	}
	for _, folder := range folders {
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

func (s *DashNGoImpl) ExportFolder(filter Filter) []string {
	var (
		result    []string
		rawFolder []byte
	)
	filesInDir, err := s.storage.FindAllFiles(getResourcePath(config.FolderResource), false)
	if err != nil {
		log.WithError(err).Fatal("Failed to read folders imports")
	}
	folders := s.ListFolder(filter)

	for _, file := range filesInDir {
		fileLocation := filepath.Join(getResourcePath(config.FolderResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file %s", fileLocation)
				continue
			}
		}
		var newFolder gclient.Folder
		if err = json.Unmarshal(rawFolder, &newFolder); err != nil {
			log.WithError(err).Warn("failed to unmarshall folder")
			continue
		}
		skipCreate := false
		for _, existingFolder := range folders {
			if existingFolder.UID == newFolder.UID {
				log.Warnf("Folder '%s' already exists, skipping", existingFolder.Title)
				skipCreate = true
			}

		}
		if skipCreate {
			continue
		}
		f, err := s.client.NewFolder(newFolder.Title, newFolder.UID)
		if err != nil {
			log.Errorf("failed to create folder %s", newFolder.Title)
			continue
		}

		result = append(result, f.Title)

	}
	return result
}

func (s *DashNGoImpl) DeleteAllFolder(filter Filter) []string {
	var result []string
	folders := s.ListFolder(filter)
	for _, folder := range folders {
		err := s.client.DeleteFolder(folder.UID)
		if err == nil {
			result = append(result, folder.Title)
		}
	}
	return result
}
