package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/config"
	"github.com/gosimple/slug"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
)

//ListDataSources: list all the currently configured datasources
func (s *DashNGoImpl) ListDataSources(filter Filter) []sdk.Datasource {

	ctx := context.Background()
	ds, err := s.legacyClient.GetAllDatasources(ctx)
	if err != nil {
		panic(err)
	}
	result := make([]sdk.Datasource, 0)
	dsSettings := s.grafanaConf.GetDataSourceSettings()
	for _, item := range ds {
		if dsSettings.FiltersEnabled() && (!dsSettings.Filters.ValidName(item.Name) || !dsSettings.Filters.ValidDataType(item.Type)) {
			log.Debugf("Skipping data source: %s since it fails filter checks with dataType of: %s", item.Name, item.Type)
			continue
		}
		if filter.Validate(map[string]string{Name: GetSlug(item.Name)}) {
			result = append(result, item)
		}
	}

	return result
}

//ImportDataSources: will read in all the configured datasources.
//NOTE: credentials cannot be retrieved and need to be set via configuration
func (s *DashNGoImpl) ImportDataSources(filter Filter) []string {
	var (
		datasources []sdk.Datasource
		dsPacked    []byte
		meta        sdk.BoardProperties
		err         error
		dataFiles   []string
	)
	datasources = s.ListDataSources(filter)
	for _, ds := range datasources {
		if dsPacked, err = json.MarshalIndent(ds, "", "	"); err != nil {
			log.Errorf("%s for %s\n", err, ds.Name)
			continue
		}

		dsPath := buildResourcePath(slug.Make(ds.Name), config.DataSourceResource)

		if err = s.storage.WriteFile(dsPath, dsPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err, meta.Slug)
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

//Removes all current datasources
func (s *DashNGoImpl) DeleteAllDataSources(filter Filter) []string {
	ctx := context.Background()
	var ds []string = make([]string, 0)
	items := s.ListDataSources(filter)
	for _, item := range items {
		msg, err := s.legacyClient.DeleteDatasource(ctx, item.ID)
		if err != nil {
			log.Warningf("Failed to delete datasource: %s, response: %s", item.Name, *msg.Message)
			continue
		}
		ds = append(ds, item.Name)
	}
	return ds
}

//ExportDataSources: exports all datasources to grafana using the credentials configured in config file.
func (s *DashNGoImpl) ExportDataSources(filter Filter) []string {
	var datasources []sdk.Datasource
	var status sdk.StatusMessage
	var exported []string = make([]string, 0)

	ctx := context.Background()
	log.Infof("Reading files from folder: %s", getResourcePath(config.DataSourceResource))
	fmt.Printf("Reading files from folder: %s", getResourcePath(config.DataSourceResource))
	filesInDir, err := s.storage.FindAllFiles(getResourcePath(config.DataSourceResource), false)
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for datasources")
	}
	//fmt.Printf("There are %d found from S3", len(filesInDir))
	datasources = s.ListDataSources(filter)

	var rawDS []byte

	dsSettings := s.grafanaConf.GetDataSourceSettings()
	for _, file := range filesInDir {
		fileLocation := filepath.Join(getResourcePath(config.DataSourceResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file: %s", fileLocation)
				continue
			}
			var newDS sdk.Datasource

			if err = json.Unmarshal(rawDS, &newDS); err != nil {
				log.WithError(err).Errorf("failed to unmarshall file: %s", fileLocation)
				continue
			}

			if !filter.Validate(map[string]string{Name: GetSlug(newDS.Name)}) {
				continue
			}
			dsConfig := s.grafanaConf
			var creds *config.GrafanaDataSource

			if *newDS.BasicAuth {
				creds, err = dsConfig.GetCredentials(newDS.Name)
				if err != nil { //Attempt to get Credentials by URL regex
					creds, _ = dsConfig.GetCredentialByUrl(newDS.URL)
				}
			} else {
				creds = nil
			}

			if dsSettings.FiltersEnabled() && (!dsSettings.Filters.ValidName(newDS.Name) || !dsSettings.Filters.ValidDataType(newDS.Type)) {
				log.Debugf("Skipping local JSON file since source: %s since it fails filter checks with dataType of: %s", newDS.Name, newDS.Type)
				continue
			}

			if creds != nil {
				user := creds.User
				var secureData map[string]string = make(map[string]string)
				newDS.BasicAuthUser = &user
				secureData["basicAuthPassword"] = creds.Password
				newDS.SecureJSONData = secureData
			} else {
				enabledAuth := false
				newDS.BasicAuth = &enabledAuth
			}

			for _, existingDS := range datasources {
				if existingDS.Name == newDS.Name {
					if status, err = s.legacyClient.DeleteDatasource(ctx, existingDS.ID); err != nil {
						log.Errorf("error on deleting datasource %s with %s", newDS.Name, err)
					}
					break
				}
			}
			if status, err = s.legacyClient.CreateDatasource(ctx, newDS); err != nil {
				log.Errorf("error on importing datasource %s with %s (%s)", newDS.Name, err, *status.Message)
			} else {
				exported = append(exported, fileLocation)
			}

		}
	}
	return exported
}
