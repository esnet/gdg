package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/grafana/grafana-openapi-client-go/client/access_control"

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

// ListConnectionPermissions lists all connection permission matching the given filter
func (s *DashNGoImpl) ListConnectionPermissions(filter outbound.Filter) []domain.ConnectionPermissionItem {
	if !s.IsEnterprise() {
		log.Fatal("Requires Enterprise to be enabled.  Please check your GDG configuration and try again")
	}
	result := make([]domain.ConnectionPermissionItem, 0)
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
		entry := domain.ConnectionPermissionItem{
			Connection:  &connections[ndx],
			Permissions: permission.GetPayload(),
		}
		result = append(result, entry)
	}

	return result
}

// DownloadConnectionPermissions download permissions to local file system
func (s *DashNGoImpl) DownloadConnectionPermissions(filter outbound.Filter) []string {
	slog.Info("Downloading connection permissions")
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	currentPermissions := s.ListConnectionPermissions(filter)
	for _, connection := range currentPermissions {
		if dsPacked, err = json.MarshalIndent(connection, "", "	"); err != nil {
			slog.Error("unable to marshall json ", "err", err.Error(), "connectionName", connection.Connection.Name)
			continue
		}
		dsPath := s.resources.BuildResourcePath(s.grafanaConf, slug.Make(connection.Connection.Name), domain.ConnectionPermissionResource, s.isLocal(), s.GetGlobals().ClearOutput)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("unable to write file. ", "filename", slug.Make(connection.Connection.Name), "error", err.Error())
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

// UploadConnectionPermissions upload connection permissions
func (s *DashNGoImpl) UploadConnectionPermissions(filter outbound.Filter) []string {
	if !s.IsEnterprise() {
		log.Fatal("Requires Enterprise to be enabled.  Please check your GDG configuration and try again")
	}
	var (
		rawFolder []byte
		dataFiles []string
	)

	// Build a name→live-connection map so that fixture UIDs (which are captured
	// from a previous Grafana instance and differ on every fresh container) are
	// replaced with the current server-assigned UIDs before any API calls.
	liveConnections := s.ListConnections(filter)
	liveByName := make(map[string]*models.DataSourceListItemDTO, len(liveConnections))
	for i := range liveConnections {
		liveByName[liveConnections[i].Name] = &liveConnections[i]
	}

	// Build login→live-ID and team-name→live-ID maps so that user/team IDs
	// stored in fixture files (which are assigned by Grafana and can differ
	// across containers) are resolved to the current server-assigned IDs before
	// any permission API calls.
	liveUsers := s.ListUsers(NewUserFilter(""))
	userLoginToID := make(map[string]int64, len(liveUsers))
	for _, u := range liveUsers {
		userLoginToID[u.Login] = u.ID
	}
	liveTeams := s.ListTeams(NewTeamFilter(""))
	teamNameToID := make(map[string]int64)
	for team := range liveTeams {
		if team.Name != nil && team.ID != nil {
			teamNameToID[*team.Name] = *team.ID
		}
	}

	orgName := s.grafanaConf.GetOrganizationName()
	filesInDir, err := s.storage.FindAllFiles(s.grafanaConf.GetPath(domain.ConnectionPermissionResource, orgName), false)
	if err != nil {
		log.Fatalf("Failed to read folders permission imports: %s", err.Error())
	}
	for _, file := range filesInDir {
		fileLocation := filepath.Join(s.grafanaConf.GetPath(domain.ConnectionPermissionResource, orgName), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file %s", "filename", fileLocation, "err", err)
				continue
			}
		}
		if !filter.Validate(context.Background(), domain.ConnectionName, rawFolder) {
			slog.Debug("File does not match pattern, skipping file", "filename", file)
			continue
		}
		newEntries := new(domain.ConnectionPermissionItem)
		err = json.Unmarshal(rawFolder, &newEntries)
		if err != nil {
			slog.Warn("Failed to Decode payload for file", "filename", fileLocation)
			continue
		}

		// Replace the fixture connection (stale UID) with the live connection (fresh UID).
		if live, ok := liveByName[newEntries.Connection.Name]; ok {
			newEntries.Connection = live
		} else {
			slog.Warn("connection from fixture not found on server, skipping permissions upload",
				slog.String("name", newEntries.Connection.Name))
			continue
		}

		// Get current permissions
		permissions, err := s.getConnectionPermission(newEntries.Connection.UID)
		if err != nil {
			slog.Error("connection permission could not be retrieved, cannot update permissions")
			continue
		}

		var removePermissionError error
		// Delete datasource Permissions.
		// Only attempt to remove managed, non-inherited, non-built-in-role permissions.
		// Inherited entries, non-managed entries, and built-in role grants (Viewer/Editor/Admin)
		// are system-level and cannot be removed via the access-control API in Grafana v13+;
		// attempting to do so returns an error that would abort the whole upload for this connection.
		// Only explicit user and team permission grants are safe to clear here.
		for _, p := range permissions.GetPayload() {
			if p.IsInherited || !p.IsManaged || p.BuiltInRole != "" {
				slog.Debug("skipping non-removable permission during pre-clear",
					slog.String("userLogin", p.UserLogin),
					slog.String("team", p.Team),
					slog.String("builtInRole", p.BuiltInRole),
					slog.Bool("isManaged", p.IsManaged),
					slog.Bool("isInherited", p.IsInherited),
				)
				continue
			}
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
			// Skip built-in role entries (Viewer, Editor, Admin, etc.) — Grafana
			// manages these grants itself and rejects attempts to set them via the
			// access-control API on v13+. Only user and team grants are applied.
			if permission.BuiltInRole != "" {
				slog.Debug("skipping built-in role permission during upload",
					slog.String("builtInRole", permission.BuiltInRole),
					slog.String("permission", permission.Permission))
				continue
			}

			// Resolve live user/team IDs — fixture IDs are captured from a previous
			// Grafana instance and may not match the current container's assignments.
			resolved := *permission
			if resolved.UserLogin != "" {
				if liveID, ok := userLoginToID[resolved.UserLogin]; ok {
					resolved.UserID = liveID
				} else {
					slog.Warn("user from fixture not found on server, skipping permission",
						slog.String("userLogin", resolved.UserLogin))
					continue
				}
			}
			if resolved.Team != "" {
				if liveID, ok := teamNameToID[resolved.Team]; ok {
					resolved.TeamID = liveID
				} else {
					slog.Warn("team from fixture not found on server, skipping permission",
						slog.String("team", resolved.Team))
					continue
				}
			}
			err = s.updatedConnectionPermission(newEntries.Connection, &resolved, resolved.Permission)
			if err != nil {
				slog.Error("Failed to update connection permissions",
					slog.Any("userId", resolved.UserLogin),
					slog.Any("resolvedUserID", resolved.UserID),
					slog.Any("team", resolved.Team),
					slog.Any("resolvedTeamID", resolved.TeamID),
					slog.Any("role", resolved.BuiltInRole),
					slog.Any("permission", resolved.Permission),
					slog.Any("connectionUID", newEntries.Connection.UID),
					slog.Any("err", err))
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

// DeleteAllConnectionPermissions clear all non-default permissions from all connections.
// Inherited and non-managed permissions (e.g. fixed admin grants in Grafana v13+) are
// skipped because they cannot be removed via the access-control API.
func (s *DashNGoImpl) DeleteAllConnectionPermissions(filter outbound.Filter) []string {
	dataSources := make([]string, 0)
	connectionPermissions := s.ListConnectionPermissions(filter)
	for _, conn := range connectionPermissions {
		success := true
		for _, p := range conn.Permissions {
			if p.IsInherited || !p.IsManaged || p.BuiltInRole != "" {
				slog.Debug("skipping non-removable permission during delete-all",
					slog.String("connection", conn.Connection.Name),
					slog.String("userLogin", p.UserLogin),
					slog.String("team", p.Team),
					slog.String("builtInRole", p.BuiltInRole),
				)
				continue
			}
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

// IsDataSourcePermissionsEnabled probes whether fine-grained per-user/per-team
// datasource permissions are available on this Grafana instance.
//
// Grafana Enterprise is required, and beyond that the license must include the
// Fine-Grained Access Control (FGAC) tier that covers datasource permissions.
// Instances without FGAC return 403 {"message":"Unlicensed"} on write calls even
// when IsEnterprise() is true.
//
// The probe works by attempting to set an empty permission on a well-known
// sentinel UID ("__probe__").  Grafana evaluates the license before looking up
// the resource, so a 403 Unlicensed response means FGAC is unavailable, while
// any other response (404 not found, 200 OK, etc.) means FGAC is active.
func (s *DashNGoImpl) IsDataSourcePermissionsEnabled() bool {
	if !s.IsEnterprise() {
		return false
	}
	p := access_control.NewSetResourcePermissionsForUserParams()
	p.UserID = 0
	p.Resource = connectionResourceType
	p.ResourceID = "__probe__"
	p.Body = &models.SetPermissionCommand{Permission: ""}
	_, err := s.GetClient().AccessControl.SetResourcePermissionsForUser(p)
	if err == nil {
		return true
	}
	errStr := err.Error()
	// 403 Unlicensed means FGAC is not available on this license tier.
	if strings.Contains(errStr, "403") && strings.Contains(errStr, "Unlicensed") {
		slog.Info("datasource permissions (FGAC) are not available on this Grafana license",
			slog.String("err", errStr))
		return false
	}
	// Any other error (404 resource not found, 400 bad request, etc.) means
	// the license check passed and FGAC is enabled — the probe resource just
	// doesn't exist, which is expected.
	return true
}
