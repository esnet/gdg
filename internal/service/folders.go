package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/folder_permissions"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slices"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
)

// FoldersApi Contract definition
type FoldersApi interface {
	ListFolder(filter filters.Filter) []*models.Hit
	DownloadFolders(filter filters.Filter) []string
	UploadFolders(filter filters.Filter) []string
	DeleteAllFolders(filter filters.Filter) []string
	//Permissions
	ListFolderPermissions(filter filters.Filter) map[*models.Hit][]*models.DashboardACLInfoDTO
	DownloadFolderPermissions(filter filters.Filter) []string
	UploadFolderPermissions(filter filters.Filter) []string
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

// checkFolderName returns true if folder is valid, otherwise false if special chars are found
// in folder name.
func (s *DashNGoImpl) checkFolderName(folderName string) bool {
	if strings.Contains(folderName, "/") || strings.Contains(folderName, "\\") {
		return false
	}
	return true
}

// DownloadFolderPermissions downloads all the current folder permissions based on filter.
func (s *DashNGoImpl) DownloadFolderPermissions(filter filters.Filter) []string {
	slog.Info("Downloading folder permissions")
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	currentPermissions := s.ListFolderPermissions(filter)
	for folder, permission := range currentPermissions {
		if dsPacked, err = json.MarshalIndent(permission, "", "	"); err != nil {
			slog.Error("Unable to marshall file", "err", err, "folderName", folder.Title)
			continue
		}
		dsPath := buildResourcePath(slug.Make(folder.Title), config.FolderPermissionResource)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "err", err.Error(), "filename", slug.Make(folder.Title))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles

}

// UploadFolderPermissions update current folder permissions to match local file system.
// Note: This expects all the current users and teams to already exist.
func (s *DashNGoImpl) UploadFolderPermissions(filter filters.Filter) []string {
	var (
		rawFolder []byte
		dataFiles []string
	)
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.FolderPermissionResource), false)
	if err != nil {
		log.Fatalf("Failed to read folders permission imports, %v", err)
	}
	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.FolderPermissionResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
		}
		uid := gjson.GetBytes(rawFolder, "0.uid")

		newEntries := make([]*models.DashboardACLUpdateItem, 0)
		err = json.Unmarshal(rawFolder, &newEntries)
		if err != nil {
			slog.Warn("Failed to Decode payload file", "filename", fileLocation)
			continue
		}
		payload := &models.UpdateDashboardACLCommand{
			Items: newEntries,
		}

		_, err := s.GetClient().FolderPermissions.UpdateFolderPermissions(uid.String(), payload)
		if err != nil {
			slog.Error("Failed to update folder permissions")
		} else {
			dataFiles = append(dataFiles, fileLocation)

		}
	}
	slog.Info("Patching server with local folder permissions")
	return dataFiles
}

// ListFolderPermissions retrieves all current folder permissions
// TODO: add concurrency to folder permissions calls
func (s *DashNGoImpl) ListFolderPermissions(filter filters.Filter) map[*models.Hit][]*models.DashboardACLInfoDTO {
	//get list of folders
	foldersList := s.ListFolder(filter)

	r := make(map[*models.Hit][]*models.DashboardACLInfoDTO, 0)

	for ndx, foldersEntry := range foldersList {
		results, err := s.GetClient().FolderPermissions.GetFolderPermissionList(foldersEntry.UID)
		if err != nil {
			msg := fmt.Sprintf("Unable to get folder permissions for folderUID: %s", foldersEntry.UID)

			var getFolderPermissionListInternalServerError *folder_permissions.GetFolderPermissionListInternalServerError
			switch {
			case errors.As(err, &getFolderPermissionListInternalServerError):
				var castError *folder_permissions.GetFolderPermissionListInternalServerError
				errors.As(err, &castError)
				slog.Error(msg, "message", *castError.GetPayload().Message, "err", err)
			default:
				slog.Error(msg, "err", err)
			}
		} else {
			r[foldersList[ndx]] = results.GetPayload()
		}
	}

	return r
}

// ListFolder list the current existing folders that match the given filter.
func (s *DashNGoImpl) ListFolder(filter filters.Filter) []*models.Hit {
	var result = make([]*models.Hit, 0)
	if config.Config().GetDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
		filter = nil
	}
	p := search.NewSearchParams()
	p.Type = &searchTypeFolder
	folderListing, err := s.GetClient().Search.Search(p)
	folderListing.GetPayload()
	if err != nil {
		log.Fatal("unable to retrieve folder list.")
	}

	for ndx, val := range folderListing.GetPayload() {
		valid := s.checkFolderName(val.Title)
		if filter == nil {
			if !valid {
				slog.Warn("Folder has an invalid character and is not supported. Path separators are not allowed", "folderName", val.Title)
				continue
			}
			result = append(result, folderListing.GetPayload()[ndx])
		} else if filter.ValidateAll(map[filters.FilterType]string{filters.FolderFilter: val.Title}) {
			if !valid {
				slog.Warn("Folder has an invalid character and is not supported. Path separators are not allowed", "folderName", val.Title)
				continue
			}
			result = append(result, folderListing.GetPayload()[ndx])
		}
	}

	return result

}

// DownloadFolders Download all the given folders matching filter
func (s *DashNGoImpl) DownloadFolders(filter filters.Filter) []string {
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	folderListing := s.ListFolder(filter)
	for _, folder := range folderListing {
		if dsPacked, err = json.MarshalIndent(folder, "", "	"); err != nil {
			slog.Error("Unable to serialize data to JSON", "err", err, "folderName", folder.Title)
			continue
		}
		dsPath := buildResourcePath(slug.Make(folder.Title), config.FolderResource)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("Unable to write file.", "err", err.Error(), "folderName", slug.Make(folder.Title))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles
}

// UploadFolders upload all the given folders to grafana
func (s *DashNGoImpl) UploadFolders(filter filters.Filter) []string {
	var (
		result    []string
		rawFolder []byte
	)
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.FolderResource), false)
	if err != nil {
		log.Fatalf("Failed to read folders imports, %v", err)
	}
	folderItems := s.ListFolder(filter)

	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.FolderResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
		}

		var newFolder models.CreateFolderCommand
		//var newFolder models.CreateFolderCommand
		if err = json.Unmarshal(rawFolder, &newFolder); err != nil {
			slog.Warn("failed to unmarshall folder", "err", err)
			continue
		}
		if !s.checkFolderName(newFolder.Title) {
			slog.Warn("Folder has an invalid character and is not supported, skipping folder", "folderName", newFolder.Title)
			continue
		}
		skipCreate := false
		for _, existingFolder := range folderItems {
			if existingFolder.UID == newFolder.UID {
				slog.Warn("Folder already exists, skipping", "folderName", existingFolder.Title)
				skipCreate = true
			}

		}
		if skipCreate {
			continue
		}
		params := folders.NewCreateFolderParams()
		params.Body = &newFolder
		f, err := s.GetClient().Folders.CreateFolder(&newFolder)
		if err != nil {
			slog.Error("failed to create folder.", "folderName", newFolder.Title, "err", err)
			continue
		}
		result = append(result, f.Payload.Title)

	}
	return result
}

// DeleteAllFolders deletes all the matching folders from grafana
func (s *DashNGoImpl) DeleteAllFolders(filter filters.Filter) []string {
	var result []string
	folderListing := s.ListFolder(filter)
	for _, folder := range folderListing {
		params := folders.NewDeleteFolderParams()
		params.FolderUID = folder.UID
		_, err := s.GetClient().Folders.DeleteFolder(params)
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

// Creates a reverse look up map, where the values are the keys and the keys are the values.
func reverseLookUp[T comparable, Y comparable](m map[T]Y) map[Y]T {
	reverse := make(map[Y]T, 0)
	for key, val := range m {
		reverse[val] = key
	}

	return reverse
}
