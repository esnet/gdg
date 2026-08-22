package api

import (
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/internal/ports/outbound"
)

// Compile-time assertion that DashNGoImpl satisfies the full GrafanaService contract.
var _ outbound.GrafanaService = (*DashNGoImpl)(nil)

// DashNGoImpl is the top-level GrafanaService implementation. It embeds baseService
// for shared infrastructure and composes typed service implementations behind their
// port interfaces for resources that have been decomposed (currently: dashboards).
// All other resource types (folders, connections, alerting, etc.) continue to be
// implemented directly on DashNGoImpl until they are similarly extracted.
type DashNGoImpl struct {
	baseService                            // promoted fields + methods
	dashboardSvc outbound.DashboardService // behind the port interface
}

func (s *DashNGoImpl) DashboardSvc() outbound.DashboardService {
	return s.dashboardSvc
}

func (s *DashNGoImpl) GetGdgConfig() *config_domain.GDGAppConfiguration {
	return s.gdgConfig
}

func (s *DashNGoImpl) SetStorage(v outbound.Storage) {
	s.storage = v
}

func NewDashNGo(
	cfg *config_domain.GDGAppConfiguration,
	encoder outbound.CipherEncoder,
	disk outbound.Storage,
	extended outbound.ExtendedApi,
	resource ports.Resources,
) outbound.GrafanaService {
	base := baseService{
		gdgConfig:   cfg,
		grafanaConf: cfg.GetDefaultGrafanaConfig(),
		extended:    extended,
		encoder:     encoder,
		storage:     disk,
		resources:   resource,
	}

	obj := &DashNGoImpl{
		baseService: base,
	}

	// DashboardServiceImpl gets its own baseService copy — same values, no
	// back-reference to DashNGoImpl. Stored behind the port interface.
	obj.dashboardSvc = NewDashboardService(base)

	return obj
}

// ---------------------------------------------------------------------------
// DashboardService — thin delegation behind the port interface
// ---------------------------------------------------------------------------

// ListDashboards delegates to the DashboardService implementation which routes
// between v1 and v2 based on server version and experimental config.
func (s *DashNGoImpl) ListDashboards(filter outbound.Filter) []*domain.DashboardV2Gdg {
	return s.dashboardSvc.ListDashboards(filter)
}

// DownloadDashboards delegates to the DashboardService implementation.
func (s *DashNGoImpl) DownloadDashboards(filter outbound.Filter) []string {
	return s.dashboardSvc.DownloadDashboards(filter)
}

// UploadDashboards delegates to the DashboardService implementation, which
// inspects each file's GdgApiVersion and routes to v1 or v2 with mismatch warnings.
func (s *DashNGoImpl) UploadDashboards(filterReq outbound.Filter) ([]string, error) {
	return s.dashboardSvc.UploadDashboards(filterReq)
}

// DeleteAllDashboards delegates to the DashboardService implementation.
func (s *DashNGoImpl) DeleteAllDashboards(filter outbound.Filter) []string {
	return s.dashboardSvc.DeleteAllDashboards(filter)
}

// TestCreatedFolders is a test entry point that exercises folder creation logic
// independently. It delegates to baseService.createdFoldersWithBaseUID which is
// promoted via embedding.
func (s *DashNGoImpl) TestCreatedFolders(folderName string) (map[string]string, error) {
	return s.createdFoldersWithBaseUID(folderName, "")
}

// ---------------------------------------------------------------------------
// DashboardPermissionsApi — delegated to DashboardService
// ---------------------------------------------------------------------------

func (s *DashNGoImpl) ListDashboardPermissions(filter outbound.Filter) ([]domain.DashboardAndPermissions, error) {
	return s.dashboardSvc.ListDashboardPermissions(filter)
}

func (s *DashNGoImpl) DownloadDashboardPermissions(filter outbound.Filter) ([]string, error) {
	return s.dashboardSvc.DownloadDashboardPermissions(filter)
}

func (s *DashNGoImpl) UploadDashboardPermissions(filter outbound.Filter) ([]string, error) {
	return s.dashboardSvc.UploadDashboardPermissions(filter)
}

func (s *DashNGoImpl) ClearDashboardPermissions(filter outbound.Filter) error {
	return s.dashboardSvc.ClearDashboardPermissions(filter)
}
