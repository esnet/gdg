package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/tools"
	"github.com/gosimple/slug"
)

var (
	DefaultFolderName   = "General"
	searchTypeDashboard = "dash-db"
	SearchTypeFolder    = "dash-folder"
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

// getFolderFromResourcePath if a use encodes a path separator in path, we can't determine the folder name.  This strips away
// all the known components of a resource type leaving only the folder name.
func getFolderFromResourcePath(storageEngine string, filePath string) (string, error) {
	basePath := fmt.Sprintf("%s/", config.Config().GetDefaultGrafanaConfig().GetPath(config.DashboardResource))
	// Take into account cloud prefix is enabled
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
		return folderName, nil
	}
	return "", errors.New("unable to parse resource to retrieve folder name")
}

func BuildResourceFolder(folderName string, resourceType config.ResourceType) string {
	if resourceType == config.DashboardResource && folderName == "" {
		folderName = DefaultFolderName
	}

	v := fmt.Sprintf("%s/%s", config.Config().GetDefaultGrafanaConfig().GetPath(resourceType), folderName)
	tools.CreateDestinationPath(v)
	return v
}

func buildResourcePath(folderName string, resourceType config.ResourceType) string {
	v := fmt.Sprintf("%s/%s.json", config.Config().GetDefaultGrafanaConfig().GetPath(resourceType), folderName)
	tools.CreateDestinationPath(filepath.Dir(v))
	return v
}
