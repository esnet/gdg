package api

import (
	"fmt"
	"github.com/esnet/gdg/config"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/apphelpers"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
)

var (
	DefaultFolderName = "General"
	DefaultFolderId   = int64(0)
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

func buildResourceFolder(folderName string, resourceType config.ResourceType) string {
	if resourceType == config.DashboardResource && folderName == "" {
		folderName = DefaultFolderName
	}

	v := fmt.Sprintf("%s/%s", apphelpers.GetCtxDefaultGrafanaConfig().GetPath(resourceType), folderName)
	CreateDestinationPath(v)
	return v
}

func buildResourcePath(folderName string, resourceType config.ResourceType) string {
	v := fmt.Sprintf("%s/%s.json", apphelpers.GetCtxDefaultGrafanaConfig().GetPath(resourceType), folderName)
	CreateDestinationPath(filepath.Dir(v))
	return v

}

// getFolderNameIDMap helper function to build a mapping for name to folderID
func getFolderNameIDMap(folders []*models.FolderSearchHit) map[string]int64 {
	var folderMap = make(map[string]int64)
	for _, folder := range folders {
		folderMap[folder.Title] = folder.ID
	}
	return folderMap
}
