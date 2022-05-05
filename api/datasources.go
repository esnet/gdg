package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
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
	ds, err := s.client.GetAllDatasources(ctx)
	if err != nil {
		panic(err)
	}
	result := make([]sdk.Datasource, 0)
	for _, item := range ds {
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
		dsSettings := s.grafanaConf.GetDataSourceSettings()
		if dsSettings.FiltersEnabled() && (!dsSettings.Filters.ValidName(ds.Name) || !dsSettings.Filters.ValidDataType(ds.Type)) {
			log.Infof("Skipping data source: %s since it fails filter checks with dataType of: %s", ds.Name, ds.Type)
			continue
		}
		if err = ioutil.WriteFile(dsPath, dsPacked, os.FileMode(int(0666))); err != nil {
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
		msg, err := s.client.DeleteDatasource(ctx, item.ID)
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
	filesInDir, err := ioutil.ReadDir(getResourcePath(config.DataSourceResource))
	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for datasources")
	}
	datasources = s.ListDataSources(filter)

	var rawDS []byte

	for _, file := range filesInDir {
		fileLocation := filepath.Join(getResourcePath(config.DataSourceResource), file.Name())
		if strings.HasSuffix(file.Name(), ".json") {
			if rawDS, err = ioutil.ReadFile(fileLocation); err != nil {
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
					if status, err = s.client.DeleteDatasource(ctx, existingDS.ID); err != nil {
						log.Errorf("error on deleting datasource %s with %s", newDS.Name, err)
					}
					break
				}
			}
			if status, err = s.client.CreateDatasource(ctx, newDS); err != nil {
				log.Errorf("error on importing datasource %s with %s (%s)", newDS.Name, err, *status.Message)
			} else {
				exported = append(exported, fileLocation)
			}

		}
	}
	return exported
}
