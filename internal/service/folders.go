package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	resourceTypes "github.com/esnet/gdg/pkg/config/domain"

	"github.com/esnet/gdg/internal/service/domain"

	"github.com/esnet/gdg/internal/service/filters/v2"

	"github.com/esnet/gdg/internal/tools/encode"

	configDomain "github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/folders"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
)

const (
	folderPathSeparator = string(os.PathSeparator)
)

func NewFolderFilter(cfg *configDomain.GDGAppConfiguration) filters.V2Filter {
	filterObj := v2.NewBaseFilter()
	err := filterObj.RegisterReader(reflect.TypeOf(&domain.NestedHit{}), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.(*domain.NestedHit)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.FolderFilter:
			return val.NestedPath, nil
		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("unable to register a valid reader for folder filter")
	}

	folderArr := cfg.GetDefaultGrafanaConfig().GetMonitoredFolders(false)
	filterObj.AddValidation(filters.FolderFilter, func(value any, expected any) error {
		val, expressions, convErr := v2.GetMismatchParams[string, []string](value, expected, filters.FolderFilter)
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
	return filterObj
}

// DownloadFolderPermissions downloads all the current folder permissions based on filter.
func (s *DashNGoImpl) DownloadFolderPermissions(filter filters.V2Filter) []string {
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
		fileName := folder.NestedPath
		if fileName == "" {
			fileName = folder.Title
		}
		dsPath := buildResourcePath(s.grafanaConf, slug.Make(fileName), resourceTypes.FolderPermissionResource, s.isLocal(), s.GetGlobals().ClearOutput)
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
func (s *DashNGoImpl) UploadFolderPermissions(filter filters.V2Filter) []string {
	var (
		rawFolder []byte
		dataFiles []string
	)
	orgName := s.grafanaConf.GetOrganizationName()
	filesInDir, err := s.storage.FindAllFiles(s.grafanaConf.GetPath(resourceTypes.FolderPermissionResource, orgName), false)
	if err != nil {
		log.Fatalf("Failed to read folders permission imports, %v", err)
	}
	for _, file := range filesInDir {
		fileLocation := filepath.Join(s.grafanaConf.GetPath(resourceTypes.FolderPermissionResource, orgName), file)
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

		_, err := s.GetClient().Folders.UpdateFolderPermissions(uid.String(), payload)
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
func (s *DashNGoImpl) ListFolderPermissions(filter filters.V2Filter) map[*domain.NestedHit][]*models.DashboardACLInfoDTO {
	// get list of folders
	var foldersList []*domain.NestedHit
	if filter == nil {
		foldersList = s.ListFolders(nil)
	} else {
		foldersList = s.ListFolders(NewFolderFilter(s.gdgConfig))
	}

	r := make(map[*domain.NestedHit][]*models.DashboardACLInfoDTO)

	for ndx, foldersEntry := range foldersList {
		results, err := s.GetClient().Folders.GetFolderPermissionList(foldersEntry.UID)
		if err != nil {
			msg := fmt.Sprintf("Unable to get folder permissions for folderUID: %s", foldersEntry.UID)

			var getFolderPermissionListInternalServerError *folders.GetFolderPermissionListInternalServerError
			switch {
			case errors.As(err, &getFolderPermissionListInternalServerError):
				var castError *folders.GetFolderPermissionListInternalServerError
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

// ListFolders list the current existing folders that match the given filter.
func (s *DashNGoImpl) ListFolders(filter filters.V2Filter) []*domain.NestedHit {
	result := make([]*domain.NestedHit, 0)
	if s.grafanaConf.GetDashboardSettings().IgnoreFilters {
		filter = nil
	}

	p := search.NewSearchParams()
	p.Type = &SearchTypeFolder
	folderRawListing, err := s.GetClient().Search.Search(p)
	if err != nil {
		log.Fatal("unable to retrieve folder list.")
	}

	folderListing := make([]*domain.NestedHit, 0)

	lo.ForEach(folderRawListing.GetPayload(), func(item *models.Hit, index int) {
		newItem := &domain.NestedHit{Hit: item}
		folderListing = append(folderListing, newItem)
	})
	folderUid := getFolderUIDEntityMapByList(folderListing)

	addFolder := func(ndx int, nestedVal string) {
		item := folderListing[ndx]
		item.NestedPath = nestedVal
		result = append(result, item)
	}
	for ndx, val := range folderListing {
		nestedVal := getNestedFolder(val.Title, val.UID, folderUid)
		val.NestedPath = nestedVal
		if filter == nil || filter.Validate(filters.FolderFilter, val) { // filter.ValidateAll(map[filters.FilterType]string{filters.FolderFilter: nestedVal}) {
			addFolder(ndx, nestedVal)
		}
	}

	return result
}

// DownloadFolders Download all the given folders matching filter
func (s *DashNGoImpl) DownloadFolders(filter filters.V2Filter) []string {
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	folderListing := s.ListFolders(filter)
	for _, folder := range folderListing {
		if dsPacked, err = json.MarshalIndent(folder, "", "	"); err != nil {
			slog.Error("Unable to serialize data to JSON", "err", err, "folderName", folder.Title)
			continue
		}
		dsPath := buildResourcePath(s.grafanaConf, folder.NestedPath, resourceTypes.FolderResource, s.isLocal(), s.GetGlobals().ClearOutput)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("Unable to write file.", "err", err.Error(), "folderName", slug.Make(folder.Title))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles
}

// getPathFolderList constructs
func getPathFolderList(folder string) []string {
	elements := strings.Split(folder, folderPathSeparator)
	elements = lo.Filter(elements, func(item string, index int) bool {
		return item != "" && item != folderPathSeparator
	})
	if len(elements) == 1 {
		return nil
	}

	return elements[0 : len(elements)-1]
}

func getNestedFolderFromFile(file string, resourceDir string) string {
	folderNestPath := filepath.Dir(file)
	folderName := filepath.Base(file)
	folderNestPath = strings.Replace(folderNestPath, resourceDir, "", 1)
	folderNestPath = strings.TrimPrefix(folderNestPath, "/")
	return path.Join(folderNestPath, strings.Replace(folderName, ".json", "", 1))
}

// UploadFolders upload all the given folders to grafana
// TODO: handle setting parent
func (s *DashNGoImpl) UploadFolders(filter filters.V2Filter) []string {
	var (
		result    []string
		rawFolder []byte
	)
	// addFolder
	addFolder := func(getCreateCmd func() (*models.CreateFolderCommand, error), existingFolders map[string]*domain.NestedHit) (string, error) {
		const empty = ""

		newFolder, err := getCreateCmd()
		if err != nil {
			return empty, err
		}

		if existingFolder, ok := existingFolders[newFolder.UID]; ok {
			slog.Debug("Folder already exists, skipping", "folderName", existingFolder.Title)
			return empty, nil
		}

		params := folders.NewCreateFolderParams()
		params.Body = newFolder
		f, err := s.GetClient().Folders.CreateFolder(newFolder)
		if err != nil {
			return empty, err
		}
		return f.GetPayload().UID, err
	}

	resourceDir := s.grafanaConf.GetPath(resourceTypes.FolderResource, s.grafanaConf.GetOrganizationName())
	filesInDir, err := s.storage.FindAllFiles(resourceDir, true)
	if err != nil {
		log.Fatalf("Failed to read folders imports, %v", err)
	}
	folderItems := s.ListFolders(filter)
	folderUidMap := getFolderUIDEntityMapByList(folderItems)
	// build a mapping of the nested path to UID for all existing folders
	nestedPathToUidExisting := getFolderMapping(folderItems,
		func(fld *domain.NestedHit) string {
			return getNestedFolder(fld.Title, fld.UID, folderUidMap)
		},
		func(fld *domain.NestedHit) *domain.NestedHit { return fld },
	)

	// build nested path of local file for all files being processed
	nestedPathMap := s.buildNestedFilePath(filesInDir)
	processed := make(map[string]bool)

	for _, fileLocation := range filesInDir {
		if processed[fileLocation] {
			slog.Debug("Skipping file, already processed", slog.Any("file", fileLocation))
			continue
		}
		slog.Debug("processing file", slog.Any("file", fileLocation))
		if strings.HasSuffix(fileLocation, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
		}

		parentUid := ""
		nestedFolder := getNestedFolderFromFile(fileLocation, resourceDir)
		requiredFolders := getPathFolderList(nestedFolder)
		// check if nested folder exists.
		sb := new(strings.Builder)
		errorOut := false
		parentFolder := ""

		for ndx, fld := range requiredFolders {
			parentFolder = sb.String()
			if ndx == 0 {
				sb.WriteString(fld)
			} else {
				sb.WriteString(folderPathSeparator + fld)
			}
			// parentFolder folder missing, create entity
			if entity, ok := nestedPathToUidExisting[sb.String()]; !ok {
				// subfolder does not exist and needs to be created
				slog.Info("Parent Folder is missing, attempting to create parentFolder folder", slog.Any("parentFolder", sb.String()), slog.Any("folder", nestedFolder))
				// check if folder definition exists.
				var parentAddErr error
				if parentFile, parentOk := nestedPathMap[sb.String()]; parentOk {
					getNewFolder := func() (*models.CreateFolderCommand, error) {
						if strings.HasSuffix(parentFile, ".json") {
							if rawFolder, err = s.storage.ReadFile(parentFile); err != nil {
								slog.Error("failed to read fileOrName", "filename", parentFile, "err", err)
							}
						}
						newFolderCmd := &models.CreateFolderCommand{}
						if err := json.Unmarshal(rawFolder, &newFolderCmd); err != nil {
							slog.Warn("failed to unmarshall folder", "err", err)
							return newFolderCmd, err
						}
						r := gjson.Get(string(rawFolder), "folderUid")
						if r.String() != "" {
							newFolderCmd.ParentUID = r.String()
						}
						return newFolderCmd, nil
					}
					parentUid, parentAddErr = addFolder(getNewFolder, folderUidMap)
					if parentAddErr != nil {
						slog.Error("Unable to created parentFolder folder", slog.Any("parentFolder", parentFile))
						errorOut = true
					}
				} else {
					getNewFolder := func() (*models.CreateFolderCommand, error) {
						newFolderCmd := new(models.CreateFolderCommand)
						newFolderCmd.Title = encode.Decode(fld)
						if val, ok := nestedPathToUidExisting[parentFolder]; ok {
							newFolderCmd.ParentUID = val.UID
						}
						return newFolderCmd, nil
					}
					// no matching file, use title
					parentUid, parentAddErr = addFolder(getNewFolder, folderUidMap)
					if parentAddErr != nil {
						slog.Error("Unable to created parentFolder folder", slog.Any("parentFolder", parentFile))
						errorOut = true
					}
				}
				if errorOut {
					break
				}
				processed[filepath.Join(resourceDir, fmt.Sprintf("%s.json", sb.String()))] = true
				newParentFolder, err := s.getFolderByUid(parentUid)
				if err != nil {
					slog.Error("unable to get newly created parentFolder folder", slog.Any("parentFolder", sb.String()))
					break
				}
				folderUidMap[parentUid] = newParentFolder
				nestedPathToUidExisting[sb.String()] = newParentFolder

			} else {
				parentUid = entity.UID
				// folder exists, continue
				slog.Debug("Parent already exists, continuing", slog.Any("ParentFolder", sb.String()))
				parentResource := filepath.Join(resourceDir, fmt.Sprintf("%s.json", sb.String()))
				if val, entryExists := processed[parentResource]; !entryExists || !val {
					processed[parentResource] = true
				}
			}

		}
		if errorOut {
			slog.Error("unable to add folder", slog.Any("folder", nestedFolder))
			continue
		}
		var newFolder models.CreateFolderCommand
		if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
			slog.Error("failed to read file", "filename", fileLocation, "err", err)
			continue
		}
		if err = json.Unmarshal(rawFolder, &newFolder); err != nil {
			slog.Warn("failed to unmarshall folder", "err", err)
			continue
		}

		skipCreate := false
		if existingFolder, ok := folderUidMap[newFolder.UID]; ok {
			slog.Debug("Folder already exists with given UID, skipping", "folderName", existingFolder.Title)
			skipCreate = true
		}
		if existingFolder, ok := nestedPathToUidExisting[nestedFolder]; ok {
			slog.Debug("Folder with name path already exists", slog.String("nestedPath", nestedFolder), slog.String("folder", existingFolder.Title))
			skipCreate = true
		}

		if skipCreate {
			slog.Warn("folder already exists, skipping creation", slog.Any("folder", fileLocation))
			continue
		}
		params := folders.NewCreateFolderParams()
		// patch parentFolder here if nested
		if newFolder.ParentUID == "" {
			newFolder.ParentUID = parentUid
		}
		params.Body = &newFolder
		f, createErr := s.GetClient().Folders.CreateFolder(&newFolder)
		if err != nil {
			slog.Error("failed to create folder.", "folderName", newFolder.Title, "err", createErr)
			continue
		}
		processed[fileLocation] = true

		folderUidMap[f.GetPayload().UID] = s.folderToHit(f.GetPayload())
		nestedPathToUidExisting[nestedFolder] = s.folderToHit(f.GetPayload())
		result = append(result, nestedFolder)

	}
	return result
}

// buildNestedFilePath returns a dictionary of nestedPaths to a matching file if one exists.
func (s *DashNGoImpl) buildNestedFilePath(files []string) map[string]string {
	orgName := s.grafanaConf.GetOrganizationName()
	resourceBaseDir := s.grafanaConf.GetPath(resourceTypes.FolderResource, orgName)
	m := make(map[string]string)

	for _, file := range files {
		folderNestPath := getNestedFolderFromFile(file, resourceBaseDir)
		m[folderNestPath] = file
	}

	return m
}

// DeleteAllFolders deletes all the matching folders from grafana
func (s *DashNGoImpl) DeleteAllFolders(filter filters.V2Filter) []string {
	var (
		result        []string
		folderListing []*domain.NestedHit
	)
	if filter == nil {
		folderListing = s.ListFolders(nil)
	} else {
		folderListing = s.ListFolders(filter)
	}
	sort.Slice(folderListing, func(i, j int) bool {
		return strings.Compare(folderListing[i].NestedPath, folderListing[j].NestedPath) > 0
	})
	for _, folder := range folderListing {
		params := folders.NewDeleteFolderParams()
		params.FolderUID = folder.UID

		_, err := s.GetClient().Folders.DeleteFolder(params)
		if err == nil {
			result = append(result, folder.NestedPath)
		} else {
			slog.Error("failed to delete folder", "err", err, "folder", folder.Title)
		}
	}
	return result
}

// getFolderMapping returns a mapping of any comparable T to any value based on the folder entity.
// key is a function that takes the folder type as a parameter and returns the key to use for the resulting map.
// val is a function that takes the folder type as a parameter and returns the value to set the map value to.
func getFolderMapping[T comparable, V any](folders []*domain.NestedHit, key func(fld *domain.NestedHit) T, val func(fld *domain.NestedHit) V) map[T]V {
	m := make(map[T]V)
	for _, f := range folders {
		m[key(f)] = val(f)
	}
	return m
}

// getFolderUIDEntityMap builds a map from folder UID to NestedHit using ListFolders.
func (s *DashNGoImpl) getFolderUIDEntityMap(filter filters.V2Filter) map[string]*domain.NestedHit {
	return getFolderUIDEntityMapByList(s.ListFolders(filter))
}

// getFolderUIDEntityMapByList helper function to build a mapping for name to folderID
func getFolderUIDEntityMapByList(folders []*domain.NestedHit) map[string]*domain.NestedHit {
	return getFolderMapping(folders, func(fld *domain.NestedHit) string {
		return fld.UID
	},
		func(fld *domain.NestedHit) *domain.NestedHit {
			return fld
		},
	)
}

// getFolderNameUIDMap helper function to build a mapping for name to folderID
func (s *DashNGoImpl) getFolderNameUIDMap(folders []*domain.NestedHit) map[string]string {
	return getFolderMapping(folders, func(fld *domain.NestedHit) string {
		return fld.NestedPath
	},
		func(fld *domain.NestedHit) string {
			return fld.UID
		},
	)
}

// reverseLookUp Creates a reverse look up map, where the values are the keys and the keys are the values.
func reverseLookUp[T comparable, Y comparable](m map[T]Y) map[Y]T {
	reverse := make(map[Y]T)
	for key, val := range m {
		reverse[val] = key
	}

	return reverse
}

// getFolderByUid gets a given folder given a valid Uid
func (s *DashNGoImpl) getFolderByUid(uid string) (*domain.NestedHit, error) {
	res, err := s.GetClient().Folders.GetFolderByUID(uid)
	if err != nil {
		return nil, err
	}
	return s.folderToHit(res.GetPayload()), nil
}

// folderToHit converts a models.Folder struct to a models.Hit struct
func (s *DashNGoImpl) folderToHit(fld *models.Folder) *domain.NestedHit {
	res := new(domain.NestedHit)
	res.Hit = new(models.Hit)
	res.Title = fld.Title
	res.UID = fld.UID
	res.FolderUID = fld.ParentUID
	res.Type = models.HitType(SearchTypeFolder)
	res.URL = fld.URL
	paths := lo.Map(fld.Parents, func(item *models.Folder, index int) string {
		return encode.Encode(item.Title)
	})
	if val := path.Join(paths...); val == "" {
		res.NestedPath = encode.Encode(res.Title)
	} else {
		res.NestedPath = path.Join(val, encode.Encode(res.Title))
	}

	return res
}
