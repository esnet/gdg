package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/access_control"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

const (
	dashboardResourceType = "dashboards"
)

// ---------------------------------------------------------------------------
// Public CRUD — route to v1 (legacy ACL API) or v2 (RBAC access-control API)
// ---------------------------------------------------------------------------

// ListDashboardPermissions returns all dashboard permissions for the matching
// dashboards. On v13+ it uses the RBAC access-control API; on older servers it
// uses the legacy ACL API and maps the result to the unified ResourcePermissionDTO.
func (d *DashboardServiceImpl) ListDashboardPermissions(filterReq outbound.Filter) ([]domain.DashboardAndPermissions, error) {
	if d.isV2Enabled() {
		return d.listDashboardPermissionsV2(filterReq)
	}
	return d.listDashboardPermissionsV1(filterReq)
}

// DownloadDashboardPermissions saves dashboard permissions to storage, writing
// a DashboardPermissionsGdg wrapper that includes gdg_api_version so uploads
// can detect the origin format and route to the correct API.
func (d *DashboardServiceImpl) DownloadDashboardPermissions(filterReq outbound.Filter) ([]string, error) {
	if d.isV2Enabled() {
		return d.downloadDashboardPermissionsV2(filterReq)
	}
	return d.downloadDashboardPermissionsV1(filterReq)
}

// UploadDashboardPermissions reads each permissions file, inspects its
// gdg_api_version, and routes to the correct upload path:
//
//   - serverIsV2=true  + file is v2 → RBAC access-control API
//   - serverIsV2=true  + file is v1 → legacy ACL API with warning
//   - serverIsV2=false + file is v1 → legacy ACL API
//   - serverIsV2=false + file is v2 → skipped with warning
func (d *DashboardServiceImpl) UploadDashboardPermissions(filterReq outbound.Filter) ([]string, error) {
	if filterReq == nil {
		filterReq = NewDashboardFilter(d.gdgConfig, "", "", "")
	}

	serverIsV2 := d.isV2Enabled()
	orgName := d.grafanaConf.GetOrganizationName()
	folderUidMap := d.getFolderNameUIDMap(d.listFolders(NewFolderFilter(d.gdgConfig)))
	path := d.grafanaConf.GetPath(domain.DashboardPermissionsResource, orgName)

	filesInDir, err := d.storage.FindAllFiles(path, true)
	if err != nil {
		return nil, fmt.Errorf("failed to read dashboard permissions directory: %w", err)
	}

	var uploaded []string

	for _, file := range filesInDir {
		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json files are supported, skipping", "filename", file)
			continue
		}

		rawFile, readErr := d.storage.ReadFile(file)
		if readErr != nil {
			slog.Warn("Unable to read file", "filename", file, "err", readErr)
			continue
		}

		// Detect format by inspecting gdg_api_version.
		// Files with no gdg_api_version are treated as legacy v1 (backward-compat).
		wrapper := &domain.DashboardPermissionsGdg{}
		isV2File := false
		if umErr := json.Unmarshal(rawFile, wrapper); umErr == nil && wrapper.GdgApiVersion == domain.GdgPermApiVersionV2 {
			isV2File = true
		}

		// Validate folder before upload.
		folderName, foldErr := d.resources.GetFolderFromResourcePath(d.grafanaConf, file, domain.DashboardPermissionsResource, d.storage.GetPrefix(), orgName)
		if foldErr != nil || folderName == "" {
			folderName = domain.ApiConsts.DefaultFolderName
		}
		_, folderUidMap, err = d.baseFolderValidation(filterReq, folderName, folderUidMap)
		if err != nil {
			slog.Warn("folder validation failed, skipping", "file", file, "err", err)
			continue
		}

		if serverIsV2 {
			if isV2File {
				if d.uploadDashboardPermissionsV2(file, wrapper) {
					uploaded = append(uploaded, file)
				}
			} else {
				slog.Warn("v1-format dashboard permissions file detected on v13+ server; uploading via legacy ACL API. "+
					"Re-download permissions to migrate to v2 format.", "file", file)
				if d.uploadDashboardPermissionsV1(file, rawFile, folderUidMap) {
					uploaded = append(uploaded, file)
				}
			}
		} else {
			// On pre-v13 servers both v1 and v2 files are uploaded via the legacy ACL API.
			// v2 files carry ResourcePermissionDTO which is mapped back to DashboardACLUpdateItem,
			// so the same upload helper handles both formats seamlessly.
			if isV2File {
				slog.Warn("v2-format dashboard permissions file on a pre-v13 server; uploading via legacy ACL API. "+
					"Upgrade to Grafana v13+ to use the RBAC access-control API.", "file", file)
			}
			if d.uploadDashboardPermissionsV1(file, rawFile, folderUidMap) {
				uploaded = append(uploaded, file)
			}
		}
	}

	return uploaded, nil
}

// ClearDashboardPermissions removes all non-default permissions from every matching
// dashboard, routing to the appropriate API based on server version.
func (d *DashboardServiceImpl) ClearDashboardPermissions(filterReq outbound.Filter) error {
	if d.isV2Enabled() {
		return d.clearDashboardPermissionsV2(filterReq)
	}
	return d.clearDashboardPermissionsV1(filterReq)
}

// ---------------------------------------------------------------------------
// v1 — legacy ACL API (pre-v13)
// ---------------------------------------------------------------------------

// listDashboardPermissionsV1 uses the legacy search + ACL endpoints.
// DashboardACLInfoDTO results are mapped to ResourcePermissionDTO for a unified
// return type, so callers always work with the same domain struct.
func (d *DashboardServiceImpl) listDashboardPermissionsV1(filterReq outbound.Filter) ([]domain.DashboardAndPermissions, error) {
	dashboards := d.listDashboardsV1(filterReq)
	var result []domain.DashboardAndPermissions
	for _, dashboard := range dashboards {
		perms, err := d.GetClient().Dashboards.GetDashboardPermissionsListByUID(dashboard.UID)
		if err != nil {
			slog.Warn("Unable to retrieve permissions for dashboard",
				slog.String("uid", dashboard.UID),
				slog.String("name", dashboard.Title))
			continue
		}
		mapped := lo.Map(perms.GetPayload(), aclDTOToResourcePermission)
		result = append(result, domain.DashboardAndPermissions{
			Dashboard:   dashboard,
			Permissions: mapped,
		})
	}
	return result, nil
}

// downloadDashboardPermissionsV1 serializes permissions as DashboardPermissionsGdg
// with GdgApiVersion=v1 so uploads can detect the legacy format.
func (d *DashboardServiceImpl) downloadDashboardPermissionsV1(filterReq outbound.Filter) ([]string, error) {
	boardLinks, err := d.listDashboardPermissionsV1(filterReq)
	if err != nil {
		return nil, err
	}
	return d.writeDashboardPermissions(boardLinks, domain.GdgPermApiVersionV1)
}

// uploadDashboardPermissionsV1 reads a legacy or v1-tagged file and posts via
// UpdateDashboardPermissionsByUID. Returns true on success.
func (d *DashboardServiceImpl) uploadDashboardPermissionsV1(file string, rawFile []byte, _ map[string]string) bool {
	// Support both the new DashboardPermissionsGdg wrapper and the old flat array format.
	wrapper := &domain.DashboardPermissionsGdg{}
	var permissions []*models.ResourcePermissionDTO
	var dashboardUID string

	if umErr := json.Unmarshal(rawFile, wrapper); umErr == nil && wrapper.DashboardUID != "" {
		// New wrapper format.
		dashboardUID = wrapper.DashboardUID
		permissions = wrapper.Permissions
	} else {
		// Legacy flat DashboardACLInfoDTO array — extract UID via gjson.
		r := gjson.GetBytes(rawFile, "#.uid")
		if !r.Exists() || !r.IsArray() {
			slog.Error("No valid dashboard UID found in file, cannot apply permissions", "file", file)
			return false
		}
		uids := lo.Uniq(lo.Map(r.Array(), func(item gjson.Result, _ int) string { return item.String() }))
		if len(uids) != 1 {
			slog.Error("Expected exactly one dashboard UID in file", "file", file, "uids", uids)
			return false
		}
		dashboardUID = uids[0]

		var aclPerms []*models.DashboardACLInfoDTO
		if umErr2 := json.Unmarshal(rawFile, &aclPerms); umErr2 != nil || len(aclPerms) == 0 {
			slog.Error("Failed to unmarshal legacy permissions file", "file", file, "err", umErr2)
			return false
		}
		permissions = lo.Map(aclPerms, aclDTOToResourcePermission)
	}

	request := &models.UpdateDashboardACLCommand{Items: make([]*models.DashboardACLUpdateItem, 0)}
	for _, perm := range permissions {
		request.Items = append(request.Items, &models.DashboardACLUpdateItem{
			Permission: permissionStringToType(perm.Permission),
			Role:       perm.BuiltInRole,
			TeamID:     perm.TeamID,
			UserID:     perm.UserID,
		})
	}

	if _, err := d.GetClient().Dashboards.UpdateDashboardPermissionsByUID(dashboardUID, request); err != nil {
		slog.Error("Failed to upload permissions via legacy ACL API", "file", file, "err", err)
		return false
	}
	return true
}

// clearDashboardPermissionsV1 sends an empty ACL update for each dashboard,
// resetting permissions to the Grafana defaults.
func (d *DashboardServiceImpl) clearDashboardPermissionsV1(filterReq outbound.Filter) error {
	boardLinks, err := d.listDashboardPermissionsV1(filterReq)
	if err != nil {
		return err
	}
	for _, link := range boardLinks {
		request := &models.UpdateDashboardACLCommand{Items: make([]*models.DashboardACLUpdateItem, 0)}
		if _, clearErr := d.GetClient().Dashboards.UpdateDashboardPermissionsByUID(link.Dashboard.UID, request); clearErr != nil {
			slog.Error("Failed to clear permissions for dashboard",
				slog.String("dashboard", link.Dashboard.Title), slog.Any("err", clearErr))
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// v2 — RBAC access-control API (v13+)
// ---------------------------------------------------------------------------

// listDashboardPermissionsV2 uses the v2 App Platform dashboard listing and the
// RBAC access-control endpoint GET /api/access-control/dashboards/{uid}.
func (d *DashboardServiceImpl) listDashboardPermissionsV2(filterReq outbound.Filter) ([]domain.DashboardAndPermissions, error) {
	dashboards := d.listDashboardsV2(filterReq)
	var result []domain.DashboardAndPermissions
	for _, gdg := range dashboards {
		uid := gdg.Resource.Name
		perms, err := d.GetClient().AccessControl.GetResourcePermissions(uid, dashboardResourceType)
		if err != nil {
			slog.Warn("Unable to retrieve permissions for dashboard",
				slog.String("uid", uid),
				slog.String("name", gdg.Resource.Spec.Title))
			continue
		}
		// Map the v2 dashboard back to a NestedHit so the return type stays unified.
		nestedHit := v2GdgToNestedHit(gdg)
		result = append(result, domain.DashboardAndPermissions{
			Dashboard:   nestedHit,
			Permissions: perms.GetPayload(),
		})
	}
	return result, nil
}

// downloadDashboardPermissionsV2 serializes permissions as DashboardPermissionsGdg
// with GdgApiVersion=v2.
func (d *DashboardServiceImpl) downloadDashboardPermissionsV2(filterReq outbound.Filter) ([]string, error) {
	boardLinks, err := d.listDashboardPermissionsV2(filterReq)
	if err != nil {
		return nil, err
	}
	return d.writeDashboardPermissions(boardLinks, domain.GdgPermApiVersionV2)
}

// uploadDashboardPermissionsV2 sets permissions via the RBAC access-control API,
// mirroring the pattern in updatedConnectionPermission. Returns true on success.
func (d *DashboardServiceImpl) uploadDashboardPermissionsV2(file string, wrapper *domain.DashboardPermissionsGdg) bool {
	uid := wrapper.DashboardUID
	success := true

	// First clear existing managed permissions so we start from a clean state.
	existing, err := d.GetClient().AccessControl.GetResourcePermissions(uid, dashboardResourceType)
	if err != nil {
		slog.Error("Failed to retrieve existing permissions for dashboard", "uid", uid, "err", err)
		return false
	}
	for _, perm := range existing.GetPayload() {
		if err := d.updatedDashboardPermission(uid, perm, ""); err != nil {
			slog.Warn("Failed to remove existing permission", "uid", uid, "err", err)
			success = false
		}
	}
	if !success {
		slog.Error("Failed to clear existing permissions before upload, skipping", "file", file)
		return false
	}

	// Apply new permissions from the file.
	for _, perm := range wrapper.Permissions {
		if err := d.updatedDashboardPermission(uid, perm, perm.Permission); err != nil {
			slog.Error("Failed to set dashboard permission", "uid", uid, "err", err)
			success = false
		}
	}
	return success
}

// clearDashboardPermissionsV2 removes all managed permissions from each dashboard
// using the RBAC access-control API.
func (d *DashboardServiceImpl) clearDashboardPermissionsV2(filterReq outbound.Filter) error {
	boardLinks, err := d.listDashboardPermissionsV2(filterReq)
	if err != nil {
		return err
	}
	for _, link := range boardLinks {
		uid := link.Dashboard.UID
		for _, perm := range link.Permissions {
			if err := d.updatedDashboardPermission(uid, perm, ""); err != nil {
				slog.Error("Failed to clear permission for dashboard",
					slog.String("dashboard", link.Dashboard.Title), slog.Any("err", err))
			}
		}
	}
	return nil
}

// updatedDashboardPermission sets or removes a single permission on a dashboard
// via the RBAC access-control API. Mirrors updatedConnectionPermission in
// connection_permissions.go. An empty permission string removes the permission.
func (d *DashboardServiceImpl) updatedDashboardPermission(uid string, perm *models.ResourcePermissionDTO, permission string) error {
	action := "Added"
	if permission == "" {
		action = "Removed"
	}

	switch permType := getPermissionType(*perm); permType {
	case ConnectionRolePermission:
		if perm.Permission == "Admin" && permission == "" {
			// Do not remove the built-in Admin role grant — Grafana manages it.
			return nil
		}
		p := access_control.NewSetResourcePermissionsForBuiltInRoleParams()
		p.BuiltInRole = perm.BuiltInRole
		p.Resource = dashboardResourceType
		p.ResourceID = uid
		p.Body = &models.SetPermissionCommand{Permission: permission}
		r, err := d.GetClient().AccessControl.SetResourcePermissionsForBuiltInRole(p)
		if r != nil {
			slog.Debug(action+" dashboard access for builtInRole",
				slog.String("role", perm.BuiltInRole), slog.String("message", r.GetPayload().Message))
		}
		return err

	case ConnectionUserPermission:
		if perm.UserLogin == "admin" && perm.UserID == 1 {
			// Never modify the Grafana admin user's permission.
			return nil
		}
		p := access_control.NewSetResourcePermissionsForUserParams()
		p.UserID = perm.UserID
		p.Resource = dashboardResourceType
		p.ResourceID = uid
		p.Body = &models.SetPermissionCommand{Permission: permission}
		r, err := d.GetClient().AccessControl.SetResourcePermissionsForUser(p)
		if r != nil {
			slog.Debug(action+" dashboard access for user",
				slog.String("user", perm.UserLogin), slog.String("message", r.GetPayload().Message))
		}
		return err

	case ConnectionTeamPermission:
		p := access_control.NewSetResourcePermissionsForTeamParams()
		p.TeamID = perm.TeamID
		p.Resource = dashboardResourceType
		p.ResourceID = uid
		p.Body = &models.SetPermissionCommand{Permission: permission}
		r, err := d.GetClient().AccessControl.SetResourcePermissionsForTeam(p)
		if r != nil {
			slog.Debug(action+" dashboard access for team",
				slog.String("team", perm.Team), slog.String("message", r.GetPayload().Message))
		}
		return err

	default:
		return fmt.Errorf("unsupported permission type %s for dashboard %s", permType, uid)
	}
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// writeDashboardPermissions serializes each DashboardAndPermissions entry as a
// DashboardPermissionsGdg JSON file tagged with gdgApiVersion and writes it to
// the configured storage location.
func (d *DashboardServiceImpl) writeDashboardPermissions(boardLinks []domain.DashboardAndPermissions, gdgApiVersion string) ([]string, error) {
	var dataFiles []string
	for _, link := range boardLinks {
		if len(link.Permissions) == 0 {
			continue
		}
		wrapper := &domain.DashboardPermissionsGdg{
			DashboardUID:  link.Dashboard.UID,
			DashboardName: link.Dashboard.Title,
			GdgApiVersion: gdgApiVersion,
			Permissions:   link.Permissions,
		}
		dsPacked, err := json.Marshal(wrapper)
		if err != nil {
			slog.Error("unable to marshal dashboard permissions", "dashboard", link.Dashboard.Title, "err", err)
			continue
		}
		dsPath := fmt.Sprintf("%s/%s.json",
			d.resources.BuildResourceFolder(d.grafanaConf, link.Dashboard.NestedPath, domain.DashboardPermissionsResource, d.isLocal(), d.GetGlobals().ClearOutput),
			slug.Make(link.Dashboard.Title))
		if err = d.storage.WriteFile(dsPath, pretty.Pretty(dsPacked)); err != nil {
			slog.Error("unable to write permissions file", "filename", slug.Make(link.Dashboard.Title), "err", err)
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles, nil
}

// aclDTOToResourcePermission maps a legacy DashboardACLInfoDTO to the unified
// ResourcePermissionDTO type used by both v1 and v2 code paths.
func aclDTOToResourcePermission(acl *models.DashboardACLInfoDTO, _ int) *models.ResourcePermissionDTO {
	return &models.ResourcePermissionDTO{
		UserID:      acl.UserID,
		UserLogin:   acl.UserLogin,
		TeamID:      acl.TeamID,
		Team:        acl.Team,
		BuiltInRole: acl.Role,          // "Viewer"/"Editor"/"Admin" — same enum values
		Permission:  acl.PermissionName, // "View"/"Edit"/"Admin" — string, matches ResourcePermissionDTO.Permission
	}
}

// permissionStringToType converts a ResourcePermissionDTO permission string
// back to the legacy PermissionType int required by UpdateDashboardACLCommand.
// "View"→1, "Edit"→2, "Admin"→4. Unknown values default to 0.
func permissionStringToType(s string) models.PermissionType {
	switch strings.ToLower(s) {
	case "view":
		return models.PermissionType(1)
	case "edit":
		return models.PermissionType(2)
	case "admin":
		return models.PermissionType(4)
	default:
		slog.Warn("unknown permission string, defaulting to 0", "permission", s)
		return models.PermissionType(0)
	}
}

// v2GdgToNestedHit constructs a minimal NestedHit from a DashboardV2Gdg so the
// unified DashboardAndPermissions struct can be populated from v2 listing results.
func v2GdgToNestedHit(gdg *domain.DashboardV2Gdg) *domain.NestedHit {
	hit := &domain.NestedHit{
		NestedPath: gdg.NestedPath,
	}
	// Populate the embedded models.Hit fields that callers depend on.
	if gdg.Resource != nil {
		hit.Hit = &models.Hit{}
		hit.UID = gdg.Resource.Name
		hit.Title = gdg.Resource.Spec.Title
	}
	return hit
}
