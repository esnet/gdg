package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/esnet/gdg/internal/adapter/filters/v2"
	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	domain "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/tidwall/gjson"

	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/gosimple/slug"
)

func setupConnectionReaders(filterObj outbound.Filter) {
	err := filterObj.RegisterReader(reflect.TypeFor[models.DataSourceListItemDTO](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
		val, ok := a.(models.DataSourceListItemDTO)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.Name:
			return val.Name, nil

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Connection Filter, aborting.")
	}
	err = filterObj.RegisterReader(reflect.TypeFor[[]byte](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.Name:
			{
				r := gjson.GetBytes(val, "name")
				if !r.Exists() || r.IsArray() {
					return nil, fmt.Errorf("no valid connection name found")
				}
				return r.String(), nil

			}
		case domain.ConnectionName:
			{
				r := gjson.GetBytes(val, "Connection.name")
				if !r.Exists() || r.IsArray() {
					return nil, fmt.Errorf("no valid connection name found")
				}
				return r.String(), nil
			}

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Connection Filter, aborting.")
	}
}

func NewConnectionFilter(name string) outbound.Filter {
	resourceHelper := resources.NewHelpers()
	filterEntity := v2.NewBaseFilter()
	setupConnectionReaders(filterEntity)
	getValidateFunc := func(filterType domain.FilterType) func(ctx context.Context, value any, expected any) error {
		return func(ctx context.Context, value any, expected any) error {
			val, expression, convErr := v2.GetParams[string](value, expected, filterType)
			if convErr != nil {
				return convErr
			}
			if expression == "" {
				return nil
			}
			if name != resourceHelper.GetSlug(val) {
				return fmt.Errorf("invalid connection filter. Expected: %v", expression)
			}
			return nil
		}
	}

	filterEntity.AddValidation(domain.Name, getValidateFunc(domain.Name), name)
	// used to check filter for connection permissions
	filterEntity.AddValidation(domain.ConnectionName, getValidateFunc(domain.ConnectionName), name)

	return filterEntity
}

// ListConnections list all the currently configured connections
func (s *DashNGoImpl) ListConnections(filter outbound.Filter) []models.DataSourceListItemDTO {
	err := s.SwitchOrganizationByName(s.grafanaConf.GetOrganizationName())
	if err != nil {
		log.Fatalf("Failed to switch organization ID %s: ", s.grafanaConf.OrganizationName)
	}

	ds, err := s.GetClient().Datasources.GetDataSources()
	if err != nil {
		panic(err)
	}
	result := make([]models.DataSourceListItemDTO, 0)

	dsSettings := s.grafanaConf.GetConnectionSettings()
	for _, item := range ds.GetPayload() {
		if dsSettings.FiltersEnabled() && dsSettings.IsExcluded(item) {
			slog.Debug("Skipping data source, since it fails datatype filter checks", "datasource", item.Name, "datatype", item.Type)
			continue
		}
		if filter.Validate(context.Background(), domain.Name, *item) {
			result = append(result, *item)
		}
	}

	return result
}

// DownloadConnections  will read in all the configured datasources.
// NOTE: credentials cannot be retrieved and need to be set via configuration
func (s *DashNGoImpl) DownloadConnections(filter outbound.Filter) []string {
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

		dsPath := s.resources.BuildResourcePath(s.grafanaConf, slug.Make(ds.Name), domain.ConnectionResource, s.isLocal(), s.GetGlobals().ClearOutput)

		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "filename", slug.Make(ds.Name), "err", err)
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

// DeleteAllConnections Removes all current datasources
func (s *DashNGoImpl) DeleteAllConnections(filter outbound.Filter) []string {
	ds := make([]string, 0)
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
func (s *DashNGoImpl) UploadConnections(filter outbound.Filter) []string {
	var dsListing []models.DataSourceListItemDTO

	var exported []string

	orgName := s.grafanaConf.GetOrganizationName()
	slog.Info("Reading files from folder", "folder", s.grafanaConf.GetPath(domain.ConnectionResource, orgName))
	filesInDir, err := s.storage.FindAllFiles(s.grafanaConf.GetPath(domain.ConnectionResource, orgName), false)
	if err != nil {
		slog.Error("failed to list files in directory for datasources", "err", err)
	}
	dsListing = s.ListConnections(filter)

	var rawDS []byte

	dsSettings := s.grafanaConf.GetConnectionSettings()
	for _, file := range filesInDir {
		fileLocation := filepath.Join(s.grafanaConf.GetPath(domain.ConnectionResource, orgName), file)
		if !strings.HasSuffix(fileLocation, ".json") {
			slog.Debug("Ignoring file", "fileLocation", fileLocation)
			continue
		}
		if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
			slog.Error("failed to read file", "filename", fileLocation, "credentialsErr", err)
			continue
		}
		if !filter.Validate(context.Background(), domain.Name, rawDS) {
			continue
		}

		var newDS models.AddDataSourceCommand
		if err = json.Unmarshal(rawDS, &newDS); err != nil {
			slog.Error("failed to unmarshall file", "filename", fileLocation, "credentialsErr", err)
			continue
		}

		dsConfig := s.grafanaConf

		secureLocation := s.grafanaConf.SecureLocation()
		credentials, credentialsErr := dsConfig.GetCredentials(newDS, secureLocation, s.encoder)
		if credentialsErr != nil { // Attempt to get Credentials by URL regex
			slog.Warn("DataSource has no secureData configured.  Please check your configuration.")
		}

		if dsSettings.FiltersEnabled() && dsSettings.IsExcluded(newDS) {
			slog.Debug("Skipping local JSON file since source fails datatype filter checks", "datasource", newDS.Name, "datatype", newDS.Type)
			continue
		}

		if credentials != nil {
			// Sets basic auth if secureData contains it
			if credentials.User() != "" && credentials.Password() != "" {
				newDS.BasicAuthUser = credentials.User()
				newDS.BasicAuth = true
			}
			// Pass any secure data that GDG is configured to use
			newDS.SecureJSONData = *credentials
		} else {
			// if credentials are nil, then basicAuth has to be false
			newDS.BasicAuth = false
		}

		for _, existingDS := range dsListing {
			if existingDS.Name == newDS.Name {
				if _, err := s.GetClient().Datasources.DeleteDataSourceByID(fmt.Sprintf("%d", existingDS.ID)); err != nil {
					slog.Error("error on deleting datasource", "datasource", newDS.Name, "credentialsErr", err)
				}
				break
			}
		}

		if createStatus, err := s.GetClient().Datasources.AddDataSource(&newDS); err != nil {
			slog.Error("error on importing datasource", "datasource", newDS.Name, "credentialsErr", err, "createStatus", createStatus)
		} else {
			exported = append(exported, fileLocation)
		}

	}
	return exported
}
