package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	"github.com/grafana-tools/sdk"
	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var DefaultFolderName = "General"

func GetSlug(title string) string {
	return strings.ToLower(slug.Make(title))
}

//Update the slug in the board returned
func updateSlug(board *sdk.FoundBoard) {
	elements := strings.Split(board.URI, "/")
	if len(elements) > 1 {
		board.Slug = elements[len(elements)-1]
	}
}

//buildDashboardPath returns the dashboard path for a given folder
func buildDashboardPath(conf *viper.Viper, folderName string) string {
	if folderName == "" {
		folderName = DefaultFolderName
	}
	v := fmt.Sprintf("%s/%s", getResourcePath(conf, "dashboard"), folderName)
	os.MkdirAll(v, 0755)
	return v
}

//buildDataSourcePath returns the expected file for a given datasource
func buildDataSourcePath(conf *viper.Viper, name string) string {
	dsPath := getResourcePath(conf, "ds")
	v := fmt.Sprintf("%s/%s.json", dsPath, name)
	os.MkdirAll(dsPath, 0755)
	return v
}

//getResourcePath for a gven resource type: ["dashboard", "ds"] it'll return the configured location
func getResourcePath(conf *viper.Viper, resourceType string) string {
	if resourceType == "dashboard" {
		return apphelpers.GetCtxDefaultGrafanaConfig().OutputDashboard
	} else if resourceType == "ds" {
		return apphelpers.GetCtxDefaultGrafanaConfig().OutputDataSource
	}
	return ""
}

//findAllFiles recursively list all files for a given path
func findAllFiles(folder string) []string {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		log.Warn("Output folder was not found")
		return []string{}
	}
	fileList := []string{}
	err := filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return fileList
}

//getFolderNameIDMap helper function to build a mapping for name to folderID
func getFolderNameIDMap(client *sdk.Client, ctx context.Context) map[string]int {

	folders, _ := client.GetAllFolders(ctx)
	var folderMap map[string]int = make(map[string]int)
	for _, folder := range folders {
		folderMap[folder.Title] = folder.ID
	}
	return folderMap
}
