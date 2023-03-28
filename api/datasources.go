package api

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/datasources"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/config"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
)

// DataSourcesApi Contract definition
type DataSourcesApi interface {
	ListDataSources(filter Filter) []models.DataSourceListItemDTO
	ImportDataSources(filter Filter) []string
	ExportDataSources(filter Filter) []string
	DeleteAllDataSources(filter Filter) []string
}

// ListDataSources: list all the currently configured datasources

func (s *DashNGoImpl) ListDataSources(filter Filter) []models.DataSourceListItemDTO {
	ds, err := s.client.Datasources.GetDataSources(datasources.NewGetDataSourcesParams(), s.getAuth())
	if err != nil {
		panic(err)
	}
	result := make([]models.DataSourceListItemDTO, 0)

	dsSettings := s.grafanaConf.GetDataSourceSettings()
	for _, item := range ds.GetPayload() {
		if dsSettings.FiltersEnabled() && (!dsSettings.Filters.ValidName(item.Name) || !dsSettings.Filters.ValidDataType(item.Type)) {
			log.Debugf("Skipping data source: %s since it fails filter checks with dataType of: %s", item.Name, item.Type)
			continue
		}
		if filter.Validate(map[string]string{Name: GetSlug(item.Name)}) {
			result = append(result, *item)
		}
	}

	return result
}

// ImportDataSources: will read in all the configured datasources.
// NOTE: credentials cannot be retrieved and need to be set via configuration
func (s *DashNGoImpl) ImportDataSources(filter Filter) []string {
	var (
		dsListing []models.DataSourceListItemDTO
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	dsListing = s.ListDataSources(filter)
	for _, ds := range dsListing {
		if dsPacked, err = json.MarshalIndent(ds, "", "	"); err != nil {
			log.Errorf("%s for %s\n", err, ds.Name)
			continue
		}

		dsPath := buildResourcePath(slug.Make(ds.Name), config.DataSourceResource)

		if err = s.storage.WriteFile(dsPath, dsPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err, slug.Make(ds.Name))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

// Removes all current datasources
func (s *DashNGoImpl) DeleteAllDataSources(filter Filter) []string {
	var ds []string = make([]string, 0)
	items := s.ListDataSources(filter)
	for _, item := range items {
		p := datasources.NewDeleteDataSourceByIDParams()
		p.ID = fmt.Sprintf("%d", item.ID)

		dsItem, err := s.client.Datasources.DeleteDataSourceByID(p, s.getAuth())
		if err != nil {
			log.Warningf("Failed to delete datasource: %s, response: %s", item.Name, dsItem.Error())
			continue
		}
		ds = append(ds, item.Name)
	}
	return ds
}

// ExportDataSources: exports all datasources to grafana using the credentials configured in config file.
func (s *DashNGoImpl) ExportDataSources(filter Filter) []string {
	var dsListing []models.DataSourceListItemDTO

	var exported []string = make([]string, 0)

	log.Infof("Reading files from folder: %s", apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.DataSourceResource))
	filesInDir, err := s.storage.FindAllFiles(apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.DataSourceResource), false)

	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for datasources")
	}
	dsListing = s.ListDataSources(filter)

	var rawDS []byte

	dsSettings := s.grafanaConf.GetDataSourceSettings()
	for _, file := range filesInDir {
		fileLocation := filepath.Join(apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.DataSourceResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file: %s", fileLocation)
				continue
			}
			var newDS models.AddDataSourceCommand

			if err = json.Unmarshal(rawDS, &newDS); err != nil {
				log.WithError(err).Errorf("failed to unmarshall file: %s", fileLocation)
				continue
			}

			if !filter.Validate(map[string]string{Name: GetSlug(newDS.Name)}) {
				continue
			}
			dsConfig := s.grafanaConf
			var creds *config.GrafanaDataSource

			if newDS.BasicAuth {
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
				newDS.BasicAuthUser = user
				secureData["basicAuthPassword"] = creds.Password
				newDS.SecureJSONData = secureData
			} else {
				newDS.BasicAuth = false
			}

			for _, existingDS := range dsListing {
				if existingDS.Name == newDS.Name {
					deleteParam := datasources.NewDeleteDataSourceByIDParams()
					deleteParam.ID = fmt.Sprintf("%d", existingDS.ID)
					if _, err := s.client.Datasources.DeleteDataSourceByID(deleteParam, s.getAdminAuth()); err != nil {
						log.Errorf("error on deleting datasource %s with %s", newDS.Name, err)
					}
					break
				}
			}
			p := datasources.NewAddDataSourceParams().WithBody(&newDS)
			if createStatus, err := s.client.Datasources.AddDataSource(p, s.getAuth()); err != nil {
				log.Errorf("error on importing datasource %s with %s (%s)", newDS.Name, err, createStatus.Error())
			} else {
				exported = append(exported, fileLocation)
			}

		}
	}
	return exported
}
