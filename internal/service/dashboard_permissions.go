package service

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strings"

	configDomain "github.com/esnet/gdg/pkg/config/domain"

	"github.com/esnet/gdg/internal/service/domain"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/samber/lo"
	"github.com/tidwall/gjson"

	"github.com/esnet/gdg/internal/service/filters"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
)

func (s *DashNGoImpl) ListDashboardPermissions(filterReq filters.V2Filter) ([]domain.DashboardAndPermissions, error) {
	validateDashboardEnterpriseSupport(s)
	dashboards := s.ListDashboards(filterReq)
	var result []domain.DashboardAndPermissions
	for _, dashboard := range dashboards {
		item := domain.DashboardAndPermissions{Dashboard: dashboard}
		perms, err := s.GetClient().Dashboards.GetDashboardPermissionsListByUID(dashboard.UID)
		if err != nil {
			slog.Warn("Unable to retrieve permissions for dashboard",
				slog.String("uid", dashboard.UID),
				slog.String("Name", dashboard.Title))
			continue
		} else {
			item.Permissions = perms.GetPayload()
		}
		result = append(result, item)
	}

	return result, nil
}

func (s *DashNGoImpl) DownloadDashboardPermissions(filterReq filters.V2Filter) ([]string, error) {
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)
	validateDashboardEnterpriseSupport(s)
	boardLinks, err := s.ListDashboardPermissions(filterReq)
	if err != nil {
		return nil, err
	}

	for _, link := range boardLinks {
		if len(link.Permissions) == 0 {
			continue
		}
		if dsPacked, err = json.MarshalIndent(link.Permissions, "", "	"); err != nil {
			slog.Error("unable to marshall json ", "err", err.Error(), "dashboard", link.Dashboard.Title)
			continue
		}

		dsPath := fmt.Sprintf("%s/%s.json", BuildResourceFolder(s.grafanaConf, link.Dashboard.NestedPath, configDomain.DashboardPermissionsResource, s.isLocal(), s.GetGlobals().ClearOutput), slug.Make(link.Dashboard.Title))
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("unable to write file. ", "filename", slug.Make(link.Dashboard.Title), "error", err.Error())
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles, nil
}

func validateDashboardEnterpriseSupport(s *DashNGoImpl) {
	if !s.IsEnterprise() {
		log.Fatalf("Enterprise support is required for Dashboard Permissions")
	}
}

func (s *DashNGoImpl) UploadDashboardPermissions(filterReq filters.V2Filter) ([]string, error) {
	validateDashboardEnterpriseSupport(s)
	var (
		rawFile   []byte
		dataFiles []string
		err       error
	)
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter(s.gdgConfig, "", "", "")
	}

	orgName := s.grafanaConf.GetOrganizationName()
	folderUidMap := s.getFolderNameUIDMap(s.ListFolders(NewFolderFilter(s.gdgConfig)))
	path := s.grafanaConf.GetPath(configDomain.DashboardPermissionsResource, orgName)
	filesInDir, err := s.storage.FindAllFiles(path, true)
	if err != nil {
		log.Fatalf("Failed to read folders permission imports: %s", err.Error())
	}
	for _, file := range filesInDir {

		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json files are supported, skipping", "filename", file)
			continue
		}
		if rawFile, err = s.storage.ReadFile(file); err != nil {
			slog.Warn("Unable to read file", "filename", file, "err", err)
			continue
		}

		r := gjson.GetBytes(rawFile, "#.uid")
		if !r.Exists() || !r.IsArray() {
			slog.Error("No valid dashboard UID references were found, cannot apply permission", "file", file)
			continue
		}
		uids := lo.Uniq(lo.Map(r.Array(), func(item gjson.Result, index int) string {
			return item.String()
		}))
		if len(uids) > 1 {
			slog.Error("too many UID references found in file. Cannot set permissions on dashboard", "file", file, "uids", uids)
			continue
		}

		// Extract Folder Name based on path
		folderName, foldErr := getFolderFromResourcePath(s.grafanaConf, file, configDomain.DashboardPermissionsResource, s.storage.GetPrefix(), orgName)
		if foldErr != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default", "err", foldErr)
			folderName = DefaultFolderName
		} else if folderName == "" {
			folderName = DefaultFolderName
		}
		folderUidMap, err = s.baseFolderValidation(filterReq, folderName, ptr.Of(""), folderUidMap, rawFile)
		if err != nil {
			slog.Warn("validation failed, skipping", "file", file, "err", err)
			continue
		}

		var permissions []*models.DashboardACLInfoDTO
		err = json.Unmarshal(rawFile, &permissions)
		if err != nil || len(permissions) == 0 {
			slog.Error("failed to unmarshall permissions for file.", slog.String("filename", file), "err", err)
			continue
		}
		dashboardId := uids[0]
		request := &models.UpdateDashboardACLCommand{Items: make([]*models.DashboardACLUpdateItem, 0)}
		for _, permission := range permissions {
			item := &models.DashboardACLUpdateItem{
				Permission: permission.Permission,
				Role:       permission.Role,
				TeamID:     permission.TeamID,
				UserID:     permission.UserID,
			}
			request.Items = append(request.Items, item)
		}
		_, err = s.GetClient().Dashboards.UpdateDashboardPermissionsByUID(dashboardId, request)
		if err != nil {
			slog.Error("Failed to process file", slog.String("filename", file))
		} else {
			dataFiles = append(dataFiles, file)
		}
	}
	return dataFiles, nil
}

func (s *DashNGoImpl) ClearDashboardPermissions(filterReq filters.V2Filter) error {
	validateDashboardEnterpriseSupport(s)
	boardLinks, err := s.ListDashboardPermissions(filterReq)
	if err != nil {
		slog.Error("unable to retrieve dashboards", slog.Any("err", err))
		return err
	}
	for _, link := range boardLinks {
		request := &models.UpdateDashboardACLCommand{}
		request.Items = make([]*models.DashboardACLUpdateItem, 0)
		_, err := s.GetClient().Dashboards.UpdateDashboardPermissionsByUID(link.Dashboard.UID, request)
		if err != nil {
			slog.Error("Failed clear permissions fir dashboard",
				slog.String("dashboard", fmt.Sprintf("%s%s", link.Dashboard.FolderTitle, link.Dashboard.Title)))
		}
	}
	return nil
}
