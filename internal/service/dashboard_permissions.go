package service

import (
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/grafana/grafana-openapi-client-go/models"
	"log/slog"
)

type DashboardAndPermissions struct {
	Dashboard   *models.Hit
	Permissions []*models.DashboardACLInfoDTO
}

func (s *DashNGoImpl) ListDashboardPermissions(filterReq filters.Filter) ([]DashboardAndPermissions, error) {
	dashboards := s.ListDashboards(filterReq)
	var result []DashboardAndPermissions
	for _, dashboard := range dashboards {
		item := DashboardAndPermissions{Dashboard: dashboard}
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

func (s *DashNGoImpl) GetDashboardPermissions(filterReq filters.Filter) {

}

func (s *DashNGoImpl) UploadDashboardPermissions(filterReq filters.Filter) {

}
