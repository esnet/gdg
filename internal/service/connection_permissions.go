package service

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/internal/tools"
	"github.com/esnet/gdg/internal/types"
	"github.com/grafana/grafana-openapi-client-go/client/access_control"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
)

type PermissionType string

const (
	ConnectionUserPermission PermissionType = "UserPermission"
	ConnectionTeamPermission PermissionType = "TeamPermission"
	ConnectionRolePermission PermissionType = "RolePermission"
	connectionResourceType   string         = "datasources"
)

const connectionPermissionMinVersion = "v10.2.3"

// ListConnectionPermissions lists all connection permission matching the given filter
func (s *DashNGoImpl) ListConnectionPermissions(filter filters.Filter) []types.ConnectionPermissionItem {
	if !s.IsEnterprise() {
		log.Fatal("Requires Enterprise to be enabled.  Please check your GDG configuration and try again")
	} else if !tools.ValidateMinimumVersion(connectionPermissionMinVersion, s) {
		slog.Warn("Permission with connection is broken prior to 10.2.3.  GDG won't support a prior version.  Listing is allowed, but all other operations won't work.",
			slog.Any("Your Grafana Version", "v"+s.GetServerInfo()["Version"].(string)))
	}
	result := make([]types.ConnectionPermissionItem, 0)
	connections := s.ListConnections(filter)
	for ndx, connection := range connections {

		permission, err := s.getConnectionPermission(connection.UID)
		if err != nil {
			slog.Error("unable to retrieve connection permissions for ID.",
				slog.Any("uid", connection.UID),
				slog.Any("connection_name", connection.Name),
				slog.Any("err", err),
			)
			continue
		}
		entry := types.ConnectionPermissionItem{
			Connection:  &connections[ndx],
			Permissions: permission.GetPayload(),
		}
		result = append(result, entry)
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
	if !tools.ValidateMinimumVersion(connectionPermissionMinVersion, s) {
		log.Fatalf("Permission with connection is broken prior to 10.2.3.  GDG won't support a prior version.  Listing is allowed, but all other operations won't work.  Your Grafana version is: v%s", s.GetServerInfo()["Version"].(string))
	}
	currentPermissions := s.ListConnectionPermissions(filter)
	for _, connection := range currentPermissions {
		if dsPacked, err = json.MarshalIndent(connection, "", "	"); err != nil {
			slog.Error("unable to marshall json ", "err", err.Error(), "connectionName", connection.Connection.Name)
			continue
		}
		dsPath := buildResourcePath(slug.Make(connection.Connection.Name), config.ConnectionPermissionResource, s.isLocal(), s.globalConf.ClearOutput)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("unable to write file. ", "filename", slug.Make(connection.Connection.Name), "error", err.Error())
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

// UploadConnectionPermissions upload connection permissions
func (s *DashNGoImpl) UploadConnectionPermissions(filter filters.Filter) []string {
	if !tools.ValidateMinimumVersion(connectionPermissionMinVersion, s) {
		log.Fatalf("Permission with connection is broken prior to 10.2.3.  GDG won't support a prior version.  Listing is allowed, but all other operations won't work.  Your Grafana version is: v%s", s.GetServerInfo()["Version"].(string))
	}
	if !s.IsEnterprise() {
		log.Fatal("Requires Enterprise to be enabled.  Please check your GDG configuration and try again")
	}
	//if !tools.ValidateMinimumVersion("11.0.0", s) {
	//	log.Fatal("Behavior prior to version 11.0.0 is broken. ")
	//
	//}
	var (
		rawFolder []byte
		dataFiles []string
	)

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

		newEntries := new(types.ConnectionPermissionItem)
		err = json.Unmarshal(rawFolder, &newEntries)
		if err != nil {
			slog.Warn("Failed to Decode payload for file", "filename", fileLocation)
			continue
		}
		// Get current permissions
		permissions, err := s.getConnectionPermission(newEntries.Connection.UID)
		if err != nil {
			slog.Error("connection permission could not be retrieved, cannot update permissions")
			continue
		}

		var removePermissionError error
		// Delete datasource Permissions
		for _, p := range permissions.GetPayload() {
			err := s.updatedConnectionPermission(newEntries.Connection, p, "")
			if err != nil {
				removePermissionError = err
			}

		}

		if removePermissionError != nil {
			slog.Error("Failed to delete previous data, cannot update permissions")
			continue
		}

		success := true
		for _, permission := range newEntries.Permissions {
			err = s.updatedConnectionPermission(newEntries.Connection, permission, permission.Permission)
			if err != nil {
				slog.Error("Failed to update connection permissions", slog.Any("userId", permission.UserLogin), slog.Any("team", permission.Team), slog.Any("role", permission.BuiltInRole), slog.Any("permission", permission.Permission))
				success = false
			}

		}
		if success {
			dataFiles = append(dataFiles, fileLocation)
		}
	}

	slog.Info("Removing all previous permissions and re-applying")
	return dataFiles
}

// DeleteAllConnectionPermissions clear all non-default permissions from all connections
func (s *DashNGoImpl) DeleteAllConnectionPermissions(filter filters.Filter) []string {
	if !tools.ValidateMinimumVersion(connectionPermissionMinVersion, s) {
		log.Fatalf("Permission with connection is broken prior to 10.2.3.  GDG won't support a prior version.  Listing is allowed, but all other operations won't work.  Your Grafana version is: v%s", s.GetServerInfo()["Version"].(string))
	}
	dataSources := make([]string, 0)
	connectionPermissions := s.ListConnectionPermissions(filter)
	for _, conn := range connectionPermissions {
		success := true
		for _, p := range conn.Permissions {
			deleteConnectionErr := s.updatedConnectionPermission(conn.Connection, p, "")
			if deleteConnectionErr != nil {
				success = false
			}
		}
		if success {
			dataSources = append(dataSources, conn.Connection.Name)
		}
	}

	return dataSources
}

func getPermissionType(perm models.ResourcePermissionDTO) PermissionType {
	if perm.Team != "" {
		return ConnectionTeamPermission
	} else if perm.UserLogin != "" {
		return ConnectionUserPermission
	}

	return ConnectionRolePermission
}

// updatedConnectionPermission a given permission associated with a given resource.  If permission is empty string, it will be removed, otherwise it will be added.
func (s *DashNGoImpl) updatedConnectionPermission(key *models.DataSourceListItemDTO, perm *models.ResourcePermissionDTO, permission string) error {
	action := "Added"
	if permission == "" {
		action = "Removed"
	}
	permissionIdStr := fmt.Sprintf("%d", perm.ID)
	connectionId := key.UID
	switch permType := getPermissionType(*perm); permType {
	case ConnectionRolePermission:
		if perm.Permission == "Admin" {
			slog.Info("Skipping modifications to admin role permission")
			return nil
		}
		// update User Role
		// POST /api/access-control/datasources/:uid/builtInRoles/:builtinRoleName
		p := access_control.NewSetResourcePermissionsForBuiltInRoleParams()
		p.BuiltInRole = perm.BuiltInRole
		p.Resource = connectionResourceType
		p.ResourceID = key.UID
		p.Body = &models.SetPermissionCommand{Permission: permission}
		r, err := s.GetClient().AccessControl.SetResourcePermissionsForBuiltInRole(p)
		if r != nil {
			slog.Debug(action+" access for builtInRole", slog.String("role", perm.BuiltInRole), slog.String("permissionID", permissionIdStr), slog.String("message", r.GetPayload().Message))
		}
		if err != nil {
			return err
		}
	case ConnectionUserPermission:
		if perm.UserLogin == "admin" && perm.UserID == 1 {
			slog.Info("Skipping modifications to admin user permission")
			return nil
		}
		// POST /api/access-control/datasources/:uid/users/:id
		p := access_control.NewSetResourcePermissionsForUserParams()
		p.UserID = perm.UserID
		p.Body = &models.SetPermissionCommand{Permission: permission}
		p.Resource = connectionResourceType
		p.ResourceID = connectionId
		r, err := s.GetClient().AccessControl.SetResourcePermissionsForUser(p)
		if r != nil {
			slog.Debug(action+" access for user", slog.String("user", perm.UserLogin), slog.String("permissionID", permissionIdStr), slog.String("message", r.GetPayload().Message))
		}
		if err != nil {
			return err
		}
	case ConnectionTeamPermission:
		// delete Team
		// POST /api/access-control/datasources/:uid/builtInRoles/:builtinRoleName
		p := access_control.NewSetResourcePermissionsForTeamParams()
		p.TeamID = perm.TeamID
		p.Resource = connectionResourceType
		p.ResourceID = connectionId
		p.Body = &models.SetPermissionCommand{Permission: permission}
		r, err := s.GetClient().AccessControl.SetResourcePermissionsForTeam(p)
		if r != nil {
			slog.Debug(action+" access for team", slog.String("team", perm.Team), slog.String("permissionID", permissionIdStr), slog.String("message", r.GetPayload().Message))
		}
		if err != nil {
			return err
		}
	default:
		slog.Warn("permission type is not supported", slog.Any("permissionType", permType))
		return fmt.Errorf("permission type %s is not supported", permType)
	}
	return nil
}

// getConnectionPermission Get all permissions for a given connection
func (s *DashNGoImpl) getConnectionPermission(uid string) (*access_control.GetResourcePermissionsOK, error) {
	return s.GetClient().AccessControl.GetResourcePermissions(uid, connectionResourceType)
}
