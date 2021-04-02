package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/esnet/grafana-dashboard-manager/config"
)

//buildDashboardPath returns the dashboard path for a given folder
func buildDashboardPath(conf config.Provider, folderName string) string {
	v := fmt.Sprintf("%s/%s", getResourcePath(conf, "dashboard"), folderName)
	os.MkdirAll(v, 0755)
	return v
}

//buildDataSourcePath returns the expected file for a given datasource
func buildDataSourcePath(conf config.Provider, name string) string {
	dsPath := getResourcePath(conf, "ds")
	v := fmt.Sprintf("%s/%s.json", dsPath, name)
	os.MkdirAll(dsPath, 0755)
	return v
}

//getResourcePath for a gven resource type: ["dashboard", "ds"] it'll return the configured location
func getResourcePath(conf config.Provider, resourceType string) string {
	if resourceType == "dashboard" {
		return conf.GetString("env.output.dashboards")
	} else if resourceType == "ds" {
		return conf.GetString("env.output.datasources")
	}
	return ""
}

//findAllFiles recursively list all files for a given path
func findAllFiles(folder string) []string {
	fileList := []string{}
	err := filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() != true {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return fileList
}
