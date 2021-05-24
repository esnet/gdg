package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/netsage-project/grafana-dashboard-manager/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var DefaultFolderName = "General"

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
		return config.GetDefaultGrafanaConfig().OutputDashboard
	} else if resourceType == "ds" {
		return config.GetDefaultGrafanaConfig().OutputDataSource
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
