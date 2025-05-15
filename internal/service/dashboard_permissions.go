package service

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/esnet/gdg/internal/types"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
)

func (s *DashNGoImpl) ListDashboardPermissions(filterReq filters.Filter) ([]types.DashboardAndPermissions, error) {
	validateDashboardEnterpriseSupport(s)
	dashboards := s.ListDashboards(filterReq)
	var result []types.DashboardAndPermissions
	for _, dashboard := range dashboards {
		item := types.DashboardAndPermissions{Dashboard: dashboard.Hit}
		perms, err := s.GetClient().DashboardPermissions.GetDashboardPermissionsListByUID(dashboard.UID)
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

func (s *DashNGoImpl) DownloadDashboardPermissions(filterReq filters.Filter) ([]string, error) {
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

		dsPath := fmt.Sprintf("%s/%s.json", BuildResourceFolder(link.Dashboard.FolderTitle, config.DashboardPermissionsResource, s.isLocal(), s.globalConf.ClearOutput), slug.Make(link.Dashboard.Title))
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

func (s *DashNGoImpl) UploadDashboardPermissions(filterReq filters.Filter) ([]string, error) {
	if !s.IsEnterprise() {
		log.Fatalf("Enterprise support is required for Dashboard Permissions")
	}
	validateDashboardEnterpriseSupport(s)
	var (
		rawFolder  []byte
		dataFiles  []string
		folderName string
	)
	// Fallback on defaults
	if filterReq == nil {
		filterReq = NewDashboardFilter("", "", "")
	}
	validFolders := filterReq.GetEntity(filters.FolderFilter)

	path := config.Config().GetDefaultGrafanaConfig().GetPath(config.DashboardPermissionsResource)
	filesInDir, err := s.storage.FindAllFiles(path, true)
	if err != nil {
		log.Fatalf("Failed to read folders permission imports: %s", err.Error())
	}
	for _, file := range filesInDir {
		// TODO: add validation of dashboard
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")
		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json files are supported, skipping", "filename", file)
			continue
		}
		// Extract Folder Name based on path
		folderName, err = getFolderFromResourcePath(file, config.DashboardPermissionsResource, s.storage.GetPrefix())
		if err != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default")
		}
		if folderName == "" {
			folderName = DefaultFolderName
		}
		if !slices.Contains(validFolders, folderName) && !config.Config().GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters {
			slog.Debug("Skipping file since it doesn't match any valid folders", "filename", file)
			continue
		}
		validateMap := map[filters.FilterType]string{filters.FolderFilter: folderName, filters.DashFilter: baseFile}
		// If folder OR slug is filtered, then skip if it doesn't match
		if !filterReq.ValidateAll(validateMap) {
			continue
		}

		if err != nil {
			slog.Warn("unable to determine dashboard folder name, falling back on default")
		}
		if rawFolder, err = s.storage.ReadFile(file); err != nil {
			slog.Warn("Unable to read file", "filename", file, "err", err)
			continue
		}

		var permissions []*models.DashboardACLInfoDTO
		err = json.Unmarshal(rawFolder, &permissions)
		if err != nil || len(permissions) == 0 {
			slog.Error("failed to unmarshall permissions for file.", slog.String("filename", file), "err", err)
			continue
		}
		dashboardId := permissions[0].UID
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
		_, err = s.GetClient().DashboardPermissions.UpdateDashboardPermissionsByUID(dashboardId, request)
		if err != nil {
			slog.Error("Failed to process file", slog.String("filename", file))
		} else {
			dataFiles = append(dataFiles, file)
		}
	}
	return dataFiles, nil
}

func (s *DashNGoImpl) ClearDashboardPermissions(filterReq filters.Filter) error {
	validateDashboardEnterpriseSupport(s)
	boardLinks, err := s.ListDashboardPermissions(filterReq)
	if err != nil {
		slog.Error("unable to retrieve dashboards", slog.Any("err", err))
		return err
	}
	for _, link := range boardLinks {
		request := &models.UpdateDashboardACLCommand{}
		request.Items = make([]*models.DashboardACLUpdateItem, 0)
		_, err := s.GetClient().DashboardPermissions.UpdateDashboardPermissionsByUID(link.Dashboard.UID, request)
		if err != nil {
			slog.Error("Failed clear permissions fir dashboard",
				slog.String("dashboard", fmt.Sprintf("%s%s", link.Dashboard.FolderTitle, link.Dashboard.Title)))
		}
	}
	return nil
}
