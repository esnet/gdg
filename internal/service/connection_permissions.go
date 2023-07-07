package service

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/tools"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/datasource_permissions"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

type ConnectionPermissions interface {
	//Permissions Enterprise only
	ListConnectionPermissions(filter filters.Filter) map[*models.DataSourceListItemDTO]*models.DataSourcePermissionsDTO
	DownloadConnectionPermissions(filter filters.Filter) []string
	UploadConnectionPermissions(filter filters.Filter) []string
	DeleteAllConnectionPermissions(filter filters.Filter) []string
}

// ListConnectionPermissions lists all connection permission matching the given filter
func (s *DashNGoImpl) ListConnectionPermissions(filter filters.Filter) map[*models.DataSourceListItemDTO]*models.DataSourcePermissionsDTO {
	if !s.grafanaConf.IsEnterprise() {
		log.Fatal("Requires Enterprise to be enabled.  Please check your GDG configuration and try again")
	}
	result := make(map[*models.DataSourceListItemDTO]*models.DataSourcePermissionsDTO)
	connections := s.ListConnections(filter)
	for ndx, connection := range connections {
		permission, err := s.getConnectionPermission(connection.ID)
		if err != nil {
			log.Errorf("unable to retrieve connection permissions for ID: %d", connection.ID)
			continue
		}
		result[&connections[ndx]] = permission.GetPayload()

	}

	return result
}

// DownloadConnectionPermissions download permissions to local file system
func (s *DashNGoImpl) DownloadConnectionPermissions(filter filters.Filter) []string {
	log.Infof("Downloading connection permissions")
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	currentPermissions := s.ListConnectionPermissions(filter)
	for connection, permission := range currentPermissions {
		if dsPacked, err = json.MarshalIndent(permission, "", "	"); err != nil {
			log.Errorf("unable to marshall json %s for %s Permissions\n", err, connection.Name)
			continue
		}
		dsPath := buildResourcePath(slug.Make(connection.Name), config.ConnectionPermissionResource)
		if err = s.storage.WriteFile(dsPath, dsPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err.Error(), slug.Make(connection.Name))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

// UploadConnectionPermissions upload connection permissions
func (s *DashNGoImpl) UploadConnectionPermissions(filter filters.Filter) []string {
	var (
		rawFolder []byte
		dataFiles []string
	)
	if !s.grafanaConf.IsEnterprise() {
		log.Fatal("Requires Enterprise to be enabled.  Please check your GDG configuration and try again")
	}

	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionPermissionResource), false)
	if err != nil {
		log.WithError(err).Fatal("Failed to read folders permission imports")
	}
	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionPermissionResource), file)
		if !filter.ValidateAll(map[filters.FilterType]string{filters.Name: strings.ReplaceAll(file, ".json", "")}) {
			log.Debugf("File does not match pattern, skipping %s", file)
			continue
		}
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file %s", fileLocation)
				continue
			}
		}

		newEntries := new(models.DataSourcePermissionsDTO)
		err = json.Unmarshal(rawFolder, &newEntries)
		if err != nil {
			log.Warnf("Failed to Decode payload for %s", fileLocation)
			continue
		}
		//Get current permissions
		permissions, err := s.getConnectionPermission(newEntries.DatasourceID)
		if err != nil {
			log.Errorf("connection permission could not be retrieved, cannot update permissions")
			continue
		}

		success := true
		//Delete datasource Permissions
		for _, p := range permissions.GetPayload().Permissions {
			success = s.deleteConnectionPermission(p.ID, newEntries.DatasourceID)
		}

		if !success {
			log.Errorf("Failed to delete previous data, cannot update permissions")
			continue
		}

		for _, entry := range newEntries.Permissions {
			p := datasource_permissions.NewAddPermissionParams()
			p.SetUserID(tools.PtrOf(entry.UserID))
			p.SetDatasourceID(fmt.Sprintf("%d", entry.DatasourceID))
			p.SetTeamID(tools.PtrOf(entry.TeamID))
			p.SetPermission(tools.PtrOf(int64(entry.Permission)))
			if entry.BuiltInRole != "" {
				p.SetBuiltinRole(tools.PtrOf(entry.BuiltInRole))
			}
			err = s.extended.AddConnectionPermission(p)
			if err != nil {
				log.Errorf("Failed to update folder permissions")
			} else {
				dataFiles = append(dataFiles, fileLocation)
			}
		}
	}

	log.Infof("Removing all previous permissions and re-applying")
	return dataFiles
}

// DeleteAllConnectionPermissions clear all non-default permissions from all connections
func (s *DashNGoImpl) DeleteAllConnectionPermissions(filter filters.Filter) []string {
	dataSources := make([]string, 0)
	connectionPermissions := s.ListConnectionPermissions(filter)
	for key, connection := range connectionPermissions {
		success := true
		for _, p := range connection.Permissions {
			res := s.deleteConnectionPermission(p.ID, connection.DatasourceID)
			if !res {
				success = false
			}
		}
		if success {
			dataSources = append(dataSources, key.Name)
		}
	}

	return dataSources
}

// deleteConnectionPermission delete a given permission associated with a given datasourceId
func (s *DashNGoImpl) deleteConnectionPermission(permissionId int64, datasourceId int64) bool {
	deleteMe := datasource_permissions.NewDeletePermissionsParams()
	deleteMe.PermissionID = fmt.Sprintf("%d", permissionId)
	deleteMe.DatasourceID = fmt.Sprintf("%d", datasourceId)
	resp, err := s.client.DatasourcePermissions.DeletePermissions(deleteMe, s.getAuth())
	if err != nil {
		return false
	}
	log.Debugf("%d permission has been removed associated with datasource %d: %s", permissionId, datasourceId, resp.GetPayload().Message)

	return true
}

// getConnectionPermission Get all permissions for a given connection
func (s *DashNGoImpl) getConnectionPermission(id int64) (*datasource_permissions.GetAllPermissionsOK, error) {
	p := datasource_permissions.NewGetAllPermissionsParams()
	//p.DatasourceID = fmt.Sprintf("%d", connection.ID)
	p.DatasourceID = fmt.Sprintf("%d", id)
	return s.client.DatasourcePermissions.GetAllPermissions(p, s.getAuth())
}
