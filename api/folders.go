package api

import (
	"context"
	"encoding/json"
	"github.com/esnet/gdg/config"
	"github.com/gosimple/slug"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

func (s *DashNGoImpl) ListFolder(filter Filter) []sdk.Folder {
	ctx := context.Background()
	folders, err := s.client.GetAllFolders(ctx)
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
	ctx := context.Background()
	folders, err := s.client.GetAllFolders(ctx)
	if err != nil {
		log.WithError(err).Fatal("failed to create folders")
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
	ctx := context.Background()
	filesInDir, err := s.storage.ReadDir(getResourcePath(config.FolderResource))
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
		var newFolder sdk.Folder
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
		f, err := s.client.CreateFolder(ctx, newFolder)
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
	ctx := context.Background()
	folders := s.ListFolder(filter)
	for _, folder := range folders {
		success, err := s.client.DeleteFolderByUID(ctx, folder.UID)
		if err == nil && success {
			result = append(result, folder.Title)
		}
	}
	return result
}
