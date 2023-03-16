package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/config"

	"github.com/esnet/gdg/apphelpers"
	"github.com/gosimple/slug"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
)

var DefaultFolderName = "General"

func GetSlug(title string) string {
	return strings.ToLower(slug.Make(title))
}

// Update the slug in the board returned
func updateSlug(board sdk.FoundBoard) string {
	elements := strings.Split(board.URI, "/")
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

	v := fmt.Sprintf("%s/%s", getResourcePath(resourceType), folderName)
	CreateDestinationPath(v)
	return v
}

func buildResourcePath(folderName string, resourceType config.ResourceType) string {
	v := fmt.Sprintf("%s/%s.json", getResourcePath(resourceType), folderName)
	CreateDestinationPath(filepath.Dir(v))
	return v

}

// getResourcePath for a given resource type: config.ResourceType it'll return the configured location
func getResourcePath(resourceType config.ResourceType) string {
	switch resourceType {
	case config.DashboardResource:
		return apphelpers.GetCtxDefaultGrafanaConfig().GetDashboardOutput()
	case config.DataSourceResource:
		return apphelpers.GetCtxDefaultGrafanaConfig().GetDataSourceOutput()
	case config.AlertNotificationResource:
		return apphelpers.GetCtxDefaultGrafanaConfig().GetAlertNotificationOutput()
	case config.FolderResource:
		return apphelpers.GetCtxDefaultGrafanaConfig().GetFolderOutput()
	case config.UserResource:
		return apphelpers.GetCtxDefaultGrafanaConfig().GetUserOutput()
	case config.TeamResource:
		return apphelpers.GetCtxDefaultGrafanaConfig().GetTeamOutput()
	default:
		return ""
	}

}

// getFolderNameIDMap helper function to build a mapping for name to folderID
func getFolderNameIDMap(client *sdk.Client, ctx context.Context) map[string]int {

	folders, _ := client.GetAllFolders(ctx)
	var folderMap map[string]int = make(map[string]int)
	for _, folder := range folders {
		folderMap[folder.Title] = folder.ID
	}
	return folderMap
}
