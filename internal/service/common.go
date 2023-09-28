package service

import (
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultFolderName   = "General"
	DefaultFolderId     = int64(0)
	searchTypeDashboard = "dash-db"
	searchTypeFolder    = "dash-folder"
)

func GetSlug(title string) string {
	return strings.ToLower(slug.Make(title))
}

// Update the slug in the board returned
func updateSlug(board string) string {
	elements := strings.Split(board, "/")
	if len(elements) > 1 {
		return elements[len(elements)-1]
	}

	return ""
}

// CreateDestinationPath Handle osMkdir Errors
func CreateDestinationPath(v string) {
	err := os.MkdirAll(v, 0750)
	if err != nil {
		log.WithError(err).Panicf("unable to create path %s", v)
	}
}

// getFolderFromResourcePath if a use encodes a path seperator in path, we can't determine the folder name.  This strips away
// all the known components of a resource type leaving only the folder name.
func getFolderFromResourcePath(storageEngine string, filePath string, resourceType config.ResourceType) (string, error) {
	basePath := fmt.Sprintf("%s/", config.Config().GetDefaultGrafanaConfig().GetPath(config.DashboardResource))
	//Take into account cloud prefix is enabled
	if storageEngine != "" {
		cloudType, data := config.Config().GetCloudConfiguration(storageEngine)
		if cloudType != "local" && data["prefix"] != "" {
			basePath = fmt.Sprintf("%s/%s", data["prefix"], basePath)
		}
	}

	folderName := strings.Replace(filePath, basePath, "", 1)
	ndx := strings.LastIndex(folderName, string(os.PathSeparator))
	if ndx != -1 {
		folderName = folderName[0:ndx]
		log.Debugf("Folder name is: %s", folderName)
		return folderName, nil
	}
	return "", errors.New("unable to parse resource to retrieve folder name")
}

func buildResourceFolder(folderName string, resourceType config.ResourceType) string {
	if resourceType == config.DashboardResource && folderName == "" {
		folderName = DefaultFolderName
	}
	strSeperator := string(os.PathSeparator)

	if strings.Contains(folderName, strSeperator) {
		folderName = strings.ReplaceAll(folderName, strSeperator, fmt.Sprintf("//%s", strSeperator))
	}
	v := fmt.Sprintf("%s/%s", config.Config().GetDefaultGrafanaConfig().GetPath(resourceType), folderName)
	CreateDestinationPath(v)
	return v
}

func buildResourcePath(folderName string, resourceType config.ResourceType) string {
	v := fmt.Sprintf("%s/%s.json", config.Config().GetDefaultGrafanaConfig().GetPath(resourceType), folderName)
	CreateDestinationPath(filepath.Dir(v))
	return v

}
