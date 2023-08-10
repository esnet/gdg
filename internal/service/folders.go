package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/folder_permissions"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/folders"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/search"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slices"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	log.Infof("Downloading folder permissions")
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	currentPermissions := s.ListFolderPermissions(filter)
	for folder, permission := range currentPermissions {
		if dsPacked, err = json.MarshalIndent(permission, "", "	"); err != nil {
			log.Errorf("%s for %s Permissions\n", err, folder.Title)
			continue
		}
		dsPath := buildResourcePath(slug.Make(folder.Title), config.FolderPermissionResource)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			log.Errorf("%s for %s\n", err.Error(), slug.Make(folder.Title))
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
		log.WithError(err).Fatal("Failed to read folders permission imports")
	}
	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.FolderPermissionResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file %s", fileLocation)
				continue
			}
		}
		uid := gjson.GetBytes(rawFolder, "0.uid")

		newEntries := make([]*models.DashboardACLUpdateItem, 0)
		err = json.Unmarshal(rawFolder, &newEntries)
		if err != nil {
			log.Warnf("Failed to Decode payload for %s", fileLocation)
			continue
		}
		payload := &models.UpdateDashboardACLCommand{
			Items: newEntries,
		}

		p := folder_permissions.NewUpdateFolderPermissionsParams()
		p.FolderUID = uid.String()
		p.Body = payload
		_, err := s.client.FolderPermissions.UpdateFolderPermissions(p, s.getAuth())
		if err != nil {
			log.Errorf("Failed to update folder permissions")
		} else {
			dataFiles = append(dataFiles, fileLocation)

		}
	}
	log.Infof("Patching server with local folder permissions")
	return dataFiles
}

// ListFolderPermissions retrieves all current folder permissions
// TODO: add concurrency to folder permissions calls
func (s *DashNGoImpl) ListFolderPermissions(filter filters.Filter) map[*models.Hit][]*models.DashboardACLInfoDTO {
	//get list of folders
	foldersList := s.ListFolder(filter)
	cpuCount := runtime.NumCPU()
	queueList := make(chan *models.Hit, cpuCount)
	type permissionResult struct {
		key   *models.Hit
		value []*models.DashboardACLInfoDTO
	}
	outputList := make(chan permissionResult, len(foldersList))
	//Queue up tasks
	go func() {
		for ndx, _ := range foldersList {
			queueList <- foldersList[ndx]
		}
		//indicate end of data
		close(queueList)
	}()

	var lock sync.RWMutex
	_ = lock

	var wg = new(sync.WaitGroup)
	for foldersEntry := range queueList {
		wg.Add(1)
		go func(folder *models.Hit, wg *sync.WaitGroup) {
			defer wg.Done()
			//log.Infof("Starting a new go routine for folder %s", folder.FolderTitle)
			p := folder_permissions.NewGetFolderPermissionListParams()
			p.FolderUID = folder.UID
			_, client := s.GetNewClient()
			p.SetHTTPClient(client)

			//lock.Lock()
			//log.Infof("Getting folder permissions for folder %s", folder.Title)
			results, err := s.client.FolderPermissions.GetFolderPermissionList(p, s.getAuth())
			//results, err := s.client.FolderPermissions.GetFolderPermissionList(p, s.getAuth())
			//lock.Unlock()
			//log.Infof("Releasing client lock, finished retrieving data for folder: %s", folder.Title)

			if err != nil {
				msg := fmt.Sprintf("Unable to get folder permissions for folderUID: %s", p.FolderUID)
				var getFolderPermissionListInternalServerError *folder_permissions.GetFolderPermissionListInternalServerError
				switch {
				case errors.As(err, &getFolderPermissionListInternalServerError):
					castError := err.(*folder_permissions.GetFolderPermissionListInternalServerError)
					log.WithField("message", *castError.GetPayload().Message).
						WithError(err).Error(msg)
				default:
					log.WithError(err).Error(msg)
				}

			} else {

				outputList <- permissionResult{
					key:   folder,
					value: results.GetPayload(),
				}
			}
		}(foldersEntry, wg)

	}
	wg.Wait()
	close(outputList)
	r := make(map[*models.Hit][]*models.DashboardACLInfoDTO, 0)
	for entry := range outputList {
		r[entry.key] = entry.value

	}

	/*

		for ndx, foldersEntry := range foldersList {
			func(j *models.Hit) {
				p := folder_permissions.NewGetFolderPermissionListParams()
				p.FolderUID = j.UID
				results, err := s.client.FolderPermissions.GetFolderPermissionList(p, s.getAuth())
				if err != nil {
					msg := fmt.Sprintf("Unable to get folder permissions for folderUID: %s", p.FolderUID)
					switch err.(type) {
					case *folder_permissions.GetFolderPermissionListInternalServerError:
						castError := err.(*folder_permissions.GetFolderPermissionListInternalServerError)
						log.WithField("message", *castError.GetPayload().Message).
							WithError(err).Error(msg)
					default:
						log.WithError(err).Error(msg)
					}

				} else {
					r[foldersList[ndx]] = results.GetPayload()
				}

			}(foldersEntry)
		}

	*/

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
	folderListing, err := s.client.Search.Search(p, s.getAuth())
	folderListing.GetPayload()
	if err != nil {
		log.Fatal("unable to retrieve folder list.")
	}

	for ndx, val := range folderListing.GetPayload() {
		valid := s.checkFolderName(val.Title)
		if filter == nil {
			if !valid {
				log.Warningf("Folder '%s' has an invalid character and is not supported. Path seperators are not allowed", val.Title)
				continue
			}
			result = append(result, folderListing.GetPayload()[ndx])
		} else if filter.ValidateAll(map[filters.FilterType]string{filters.FolderFilter: val.Title}) {
			if !valid {
				log.Warningf("Folder '%s' has an invalid character and is not supported. Path seperators are not allowed", val.Title)
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
			log.Errorf("%s for %s\n", err, folder.Title)
			continue
		}
		dsPath := buildResourcePath(slug.Make(folder.Title), config.FolderResource)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			log.Errorf("%s for %s\n", err.Error(), slug.Make(folder.Title))
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
		if !s.checkFolderName(newFolder.Title) {
			log.Warningf("Folder '%s' has an invalid character and is not supported, skipping folder", newFolder.Title)
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

// DeleteAllFolder deletes all the matching folders from grafana
func (s *DashNGoImpl) DeleteAllFolders(filter filters.Filter) []string {
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

// Creates a reverse look up map, where the values are the keys and the keys are the values.
func reverseLookUp[T comparable, Y comparable](m map[T]Y) map[Y]T {
	reverse := make(map[Y]T, 0)
	for key, val := range m {
		reverse[val] = key
	}

	return reverse
}
