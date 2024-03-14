package service

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/grafana/grafana-openapi-client-go/models"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	"log"
)

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
	err := s.SwitchOrganizationByName(s.grafanaConf.GetOrganizationName())
	if err != nil {
		log.Fatalf("Failed to switch organization ID %s: ", s.grafanaConf.OrganizationName)
	}

	ds, err := s.GetClient().Datasources.GetDataSources()
	if err != nil {
		panic(err)
	}
	result := make([]models.DataSourceListItemDTO, 0)

	dsSettings := s.grafanaConf.GetDataSourceSettings()
	for _, item := range ds.GetPayload() {
		if dsSettings.FiltersEnabled() && dsSettings.IsExcluded(item) {
			slog.Debug("Skipping data source, since it fails datatype filter checks", "datasource", item.Name, "datatype", item.Type)
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
			slog.Error("unable to marshall file", "datasource", ds.Name, "err", err)
			continue
		}

		dsPath := buildResourcePath(slug.Make(ds.Name), config.ConnectionResource)

		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "filename", slug.Make(ds.Name), "err", err)
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

// DeleteAllConnections Removes all current datasources
func (s *DashNGoImpl) DeleteAllConnections(filter filters.Filter) []string {
	var ds []string = make([]string, 0)
	items := s.ListConnections(filter)
	for _, item := range items {
		dsItem, err := s.GetClient().Datasources.DeleteDataSourceByID(fmt.Sprintf("%d", item.ID))
		if err != nil {
			slog.Warn("Failed to delete datasource", "datasource", item.Name, "err", dsItem.Error())
			continue
		}
		ds = append(ds, item.Name)
	}
	return ds
}

// UploadConnections exports all connections to grafana using the credentials configured in config file.
func (s *DashNGoImpl) UploadConnections(filter filters.Filter) []string {
	var dsListing []models.DataSourceListItemDTO

	var exported []string

	slog.Info("Reading files from folder", "folder", config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionResource))
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionResource), false)

	if err != nil {
		slog.Error("failed to list files in directory for datasources", "err", err)
	}
	dsListing = s.ListConnections(filter)

	var rawDS []byte

	dsSettings := s.grafanaConf.GetDataSourceSettings()
	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
			var newDS models.AddDataSourceCommand

			if err = json.Unmarshal(rawDS, &newDS); err != nil {
				slog.Error("failed to unmarshall file", "filename", fileLocation, "err", err)
				continue
			}

			if !filter.ValidateAll(map[filters.FilterType]string{filters.Name: GetSlug(newDS.Name)}) {
				continue
			}
			dsConfig := s.grafanaConf

			secureLocation := config.Config().GetDefaultGrafanaConfig().GetPath(config.SecureSecretsResource)
			credentials, err := dsConfig.GetCredentials(newDS, secureLocation)
			if err != nil { //Attempt to get Credentials by URL regex
				slog.Warn("DataSource has no secureData configured.  Please check your configuration.")
			}

			if dsSettings.FiltersEnabled() && dsSettings.IsExcluded(newDS) {
				slog.Debug("Skipping local JSON file since source fails datatype filter checks", "datasource", newDS.Name, "datatype", newDS.Type)
				continue
			}

			if credentials != nil {
				//Sets basic auth if secureData contains it
				if credentials.User() != "" && (*credentials)["basicAuthPassword"] != "" {
					newDS.BasicAuthUser = credentials.User()
					newDS.BasicAuth = true
				}
				//Pass any secure data that GDG is configured to use
				newDS.SecureJSONData = *credentials
			} else {
				//if credentials are nil, then basicAuth has to be false
				newDS.BasicAuth = false
			}

			for _, existingDS := range dsListing {
				if existingDS.Name == newDS.Name {
					if _, err := s.GetClient().Datasources.DeleteDataSourceByID(fmt.Sprintf("%d", existingDS.ID)); err != nil {
						slog.Error("error on deleting datasource", "datasource", newDS.Name, "err", err)
					}
					break
				}
			}

			if createStatus, err := s.GetClient().Datasources.AddDataSource(&newDS); err != nil {
				slog.Error("error on importing datasource", "datasource", newDS.Name, "err", err, "createError", createStatus.Error())
			} else {
				exported = append(exported, fileLocation)
			}

		}
	}
	return exported
}
