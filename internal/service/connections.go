package service

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/datasources"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
)

// ConnectionsApi Contract definition
type ConnectionsApi interface {
	ListConnections(filter filters.Filter) []models.DataSourceListItemDTO
	DownloadConnections(filter filters.Filter) []string
	UploadConnections(filter filters.Filter) []string
	DeleteAllConnections(filter filters.Filter) []string
	ConnectionPermissions
}

// NewConnectionFilter
func NewConnectionFilter(name string) filters.Filter {
	filterEntity := filters.NewBaseFilter()
	filterEntity.AddFilter(filters.Name, name)
	filterEntity.AddValidation(filters.DefaultFilter, func(i interface{}) bool {
		val, ok := i.(map[filters.FilterType]string)
		if !ok {
			return ok
		}
		if filterEntity.GetFilter(filters.Name) == "" {
			return true
		}
		return val[filters.Name] == filterEntity.GetFilter(filters.Name)
	})

	return filterEntity
}

// ListConnections list all the currently configured datasources
func (s *DashNGoImpl) ListConnections(filter filters.Filter) []models.DataSourceListItemDTO {
	ds, err := s.client.Datasources.GetDataSources(datasources.NewGetDataSourcesParams(), s.getAuth())
	if err != nil {
		panic(err)
	}
	result := make([]models.DataSourceListItemDTO, 0)

	dsSettings := s.grafanaConf.GetDataSourceSettings()
	for _, item := range ds.GetPayload() {
		if dsSettings.FiltersEnabled() && dsSettings.IsExcluded(item) {
			log.Debugf("Skipping data source: %s since it fails filter checks with dataType of: %s", item.Name, item.Type)
			continue
		}
		if filter.ValidateAll(map[filters.FilterType]string{filters.Name: GetSlug(item.Name)}) {
			result = append(result, *item)
		}
	}

	return result
}

// DownloadConnections  will read in all the configured datasources.
// NOTE: credentials cannot be retrieved and need to be set via configuration
func (s *DashNGoImpl) DownloadConnections(filter filters.Filter) []string {
	var (
		dsListing []models.DataSourceListItemDTO
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	dsListing = s.ListConnections(filter)
	for _, ds := range dsListing {
		if dsPacked, err = json.MarshalIndent(ds, "", "	"); err != nil {
			log.Errorf("%s for %s\n", err, ds.Name)
			continue
		}

		dsPath := buildResourcePath(slug.Make(ds.Name), config.ConnectionResource)

		if err = s.storage.WriteFile(dsPath, dsPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err, slug.Make(ds.Name))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

// Removes all current datasources
func (s *DashNGoImpl) DeleteAllConnections(filter filters.Filter) []string {
	var ds []string = make([]string, 0)
	items := s.ListConnections(filter)
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
func (s *DashNGoImpl) UploadConnections(filter filters.Filter) []string {
	var dsListing []models.DataSourceListItemDTO

	var exported = make([]string, 0)

	log.Infof("Reading files from folder: %s", config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionResource))
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionResource), false)

	if err != nil {
		log.WithError(err).Errorf("failed to list files in directory for datasources")
	}
	dsListing = s.ListConnections(filter)

	var rawDS []byte

	dsSettings := s.grafanaConf.GetDataSourceSettings()
	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionResource), file)
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

			if !filter.ValidateAll(map[filters.FilterType]string{filters.Name: GetSlug(newDS.Name)}) {
				continue
			}
			dsConfig := s.grafanaConf
			var creds *config.GrafanaConnection

			if newDS.BasicAuth {
				creds, err = dsConfig.GetCredentials(newDS)
				if err != nil { //Attempt to get Credentials by URL regex
					log.Warn("DataSource has Auth enabled but has no valid Credentials that could be retrieved.  Please check your configuration and try again.")
				}
			} else {
				creds = nil
			}

			if dsSettings.FiltersEnabled() && dsSettings.IsExcluded(newDS) {
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
