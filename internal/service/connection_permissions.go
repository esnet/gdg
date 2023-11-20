package service

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/tools"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/datasource_permissions"
	"github.com/grafana/grafana-openapi-client-go/models"
)

type ConnectionPermissions interface {
	// Permissions Enterprise only
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
			slog.Error("unable to retrieve connection permissions for ID", "id", connection.ID)
			continue
		}
		result[&connections[ndx]] = permission.GetPayload()

	}

	return result
}

// DownloadConnectionPermissions download permissions to local file system
func (s *DashNGoImpl) DownloadConnectionPermissions(filter filters.Filter) []string {
	slog.Info("Downloading connection permissions")
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	currentPermissions := s.ListConnectionPermissions(filter)
	for connection, permission := range currentPermissions {
		if dsPacked, err = json.MarshalIndent(permission, "", "	"); err != nil {
			slog.Error("unable to marshall json ", "err", err.Error(), "connectionName", connection.Name)
			continue
		}
		dsPath := buildResourcePath(slug.Make(connection.Name), config.ConnectionPermissionResource)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("unable to write file. ", "filename", slug.Make(connection.Name), "error", err.Error())
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
		log.Fatalf("Failed to read folders permission imports: %s", err.Error())
	}
	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.ConnectionPermissionResource), file)
		if !filter.ValidateAll(map[filters.FilterType]string{filters.Name: strings.ReplaceAll(file, ".json", "")}) {
			slog.Debug("File does not match pattern, skipping file", "filename", file)
			continue
		}
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file %s", "filename", fileLocation, "err", err)
				continue
			}
		}

		newEntries := new(models.DataSourcePermissionsDTO)
		err = json.Unmarshal(rawFolder, &newEntries)
		if err != nil {
			slog.Warn("Failed to Decode payload for file", "filename", fileLocation)
			continue
		}
		// Get current permissions
		permissions, err := s.getConnectionPermission(newEntries.DatasourceID)
		if err != nil {
			slog.Error("connection permission could not be retrieved, cannot update permissions")
			continue
		}

		success := true
		// Delete datasource Permissions
		for _, p := range permissions.GetPayload().Permissions {
			success = s.deleteConnectionPermission(p.ID, newEntries.DatasourceID)
		}

		if !success {
			slog.Error("Failed to delete previous data, cannot update permissions")
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
			_, err = s.GetClient().DatasourcePermissions.AddPermission(p)
			if err != nil {
				slog.Error("Failed to update folder permissions")
			} else {
				dataFiles = append(dataFiles, fileLocation)
			}
		}
	}

	slog.Info("Removing all previous permissions and re-applying")
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
	permissionIdStr := fmt.Sprintf("%d", permissionId)
	connectionId := fmt.Sprintf("%d", datasourceId)
	resp, err := s.GetClient().DatasourcePermissions.DeletePermissions(permissionIdStr, connectionId)
	if err != nil {
		return false
	}
	slog.Debug("permission has been removed associated with datasource %d: %s", "permissionId", permissionId, "datasourceId", datasourceId, "response", resp.GetPayload().Message)
	return true
}

// getConnectionPermission Get all permissions for a given connection
func (s *DashNGoImpl) getConnectionPermission(id int64) (*datasource_permissions.GetAllPermissionsOK, error) {
	return s.GetClient().DatasourcePermissions.GetAllPermissions(fmt.Sprintf("%d", id))
}
