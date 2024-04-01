package service

import (
	"encoding/json"
	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/access_control"
	"github.com/grafana/grafana-openapi-client-go/models"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
)

const (
	connectionResource = "datasources"
)

// ListConnectionPermissions lists all connection permission matching the given filter
func (s *DashNGoImpl) ListConnectionPermissions(filter filters.Filter) map[*models.DataSourceListItemDTO][]*models.ResourcePermissionDTO {
	if !s.grafanaConf.IsEnterprise() {
		log.Fatal("Requires Enterprise to be enabled.  Please check your GDG configuration and try again")
	}
	result := make(map[*models.DataSourceListItemDTO][]*models.ResourcePermissionDTO)
	connections := s.ListConnections(filter)
	for ndx, connection := range connections {
		permission, err := s.getConnectionPermission(connection.UID)
		if err != nil {
			slog.Error("unable to retrieve connection permissions for ID", "id", connection.ID)
			continue
		}
		result[&connections[ndx]] = permission

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
	_ = rawFolder
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
		newEntries := make([]models.ResourcePermissionDTO, 0)
		err = json.Unmarshal(rawFolder, &newEntries)
		if err != nil {
			slog.Warn("Failed to Decode payload for file", "filename", fileLocation)
			continue
		}
		slog.Info("Woot", slog.Any("Size", len(newEntries)))

		/*

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

		*/
	}

	slog.Info("Removing all previous permissions and re-applying")
	return dataFiles

}

// DeleteAllConnectionPermissions clear all non-default permissions from all connections
func (s *DashNGoImpl) DeleteAllConnectionPermissions(filter filters.Filter) []string {
	dataSources := make([]string, 0)
	connectionPermissions := s.ListConnectionPermissions(filter)
	for key, connection := range connectionPermissions {
		for _, p := range connection {
			if strings.Contains(p.RoleName, "managed:users") {
				err := s.extended.UpdateUserAccessPermission(key.UID, p.UserID, api.NoPermission)
				if err != nil {
					slog.Warn("unable to remove permission for user",
						slog.Any("permissionID", p.ID),
						slog.Any("userId", p.UserID))
				} else {
					slog.Info("Remove permission for user",
						slog.Any("permissionID", p.ID),
						slog.Any("userId", p.UserID))
				}
			}
		}
		dataSources = append(dataSources, key.Name)
	}

	return dataSources
}

func (s *DashNGoImpl) removeUserPermission(permissions []*models.SetResourcePermissionCommand, resourceId string) error {

	return nil
}

func (s *DashNGoImpl) updateConnectionPermissions(permissions []*models.SetResourcePermissionCommand, resourceId string) error {
	p := access_control.NewSetResourcePermissionsParams()
	p.Resource = connectionResource
	p.ResourceID = resourceId
	body := models.SetPermissionsCommand{
		Permissions: permissions,
	}
	p.SetBody(&body)
	_, err := s.GetClient().AccessControl.SetResourcePermissions(p)
	return err

}

// deleteConnectionPermission delete a given permission associated with a given datasourceId
func (s *DashNGoImpl) deleteConnectionPermission(permissionId int64, datasourceId int64) bool {
	if true {
		return true
	}

	//permissionIdStr := fmt.Sprintf("%d", permissionId)
	//connectionId := fmt.Sprintf("%d", datasourceId)
	p := access_control.NewSetResourcePermissionsParams()
	p.ResourceID = connectionResource
	body := models.SetPermissionsCommand{
		Permissions: make([]*models.SetResourcePermissionCommand, 0),
	}
	p.SetBody(&body)
	resp, err := s.GetClient().AccessControl.SetResourcePermissions(p)
	//resp, err := s.GetClient().DatasourcePermissions.DeletePermissions(permissionIdStr, connectionId)
	if err != nil {
		return false
	}
	slog.Debug("permission has been removed associated with datasource %d: %s", "permissionId", permissionId, "datasourceId", datasourceId, "response", resp.GetPayload().Message)
	return true
}

// getConnectionPermission Get all permissions for a given connection
func (s *DashNGoImpl) getConnectionPermission(uid string) ([]*models.ResourcePermissionDTO, error) {
	pay, err := s.GetClient().AccessControl.GetResourcePermissions(uid, connectionResource)
	if err != nil {
		return nil, err
	}

	return pay.GetPayload(), nil
}
