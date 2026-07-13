package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/tools"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// DashboardServiceImpl is the standalone implementation of outbound.DashboardService.
// It embeds baseService directly for all shared infrastructure — there is no
// back-reference to DashNGoImpl. All dashboard-specific logic (v1 API, v2 App
// Platform API, permissions) lives here.
type DashboardServiceImpl struct {
	baseService
}

// NewDashboardService constructs a DashboardServiceImpl from a copy of baseService.
// The caller owns the baseService value; DashboardServiceImpl keeps its own copy
// so there is no shared mutable state with DashNGoImpl.
func NewDashboardService(base baseService) outbound.DashboardService {
	return &DashboardServiceImpl{baseService: base}
}

// isV2Enabled returns true when the v2 App Platform dashboard API should be used.
// The v2 API is available on Grafana v13.0.0 or newer; older servers use the legacy v1 API.
func (d *DashboardServiceImpl) isV2Enabled() bool {
	return tools.ValidateMinimumVersion("v13.0.0", d)
}

// ---------------------------------------------------------------------------
// DashboardService — public CRUD methods
// ---------------------------------------------------------------------------

// ListDashboards routes to v1 or v2 and always returns []*domain.DashboardV2Gdg.
func (d *DashboardServiceImpl) ListDashboards(filter outbound.Filter) []*domain.DashboardV2Gdg {
	if d.isV2Enabled() {
		return d.listDashboardsV2(filter)
	}
	return mapV1ToDashboardV2Gdg(d.listDashboardsV1(filter))
}

// DownloadDashboards saves all matching dashboards to storage, routing to v1 or v2.
func (d *DashboardServiceImpl) DownloadDashboards(filter outbound.Filter) []string {
	if d.isV2Enabled() {
		return d.downloadDashboardsV2(filter)
	}
	return d.downloadDashboardsV1(filter)
}

// ---------------------------------------------------------------------------
// Upload routing
// ---------------------------------------------------------------------------

// UploadDashboards scans each file's GdgApiVersion and routes accordingly.
//
//   - serverIsV2=true:  v2 files → v2 upload; v1 files → v1 fallback with warning
//   - serverIsV2=false: v1 files → v1 upload; v2 files → skipped with warning
//
// Files are read exactly once here and partitioned into v1Files/v2Files maps
// (path → raw bytes) that are passed directly to the private upload helpers,
// eliminating redundant FindAllFiles and ReadFile calls in those functions.
//
// A single processed map (dashboard UID / resource name → true) is built from
// all local files during partitioning and passed to pruneDashboards, which
// deletes any remote dashboards not represented locally using the appropriate API.
func (d *DashboardServiceImpl) UploadDashboards(filter outbound.Filter) ([]string, error) {
	serverIsV2 := d.isV2Enabled()
	if filter == nil {
		filter = NewDashboardFilter(d.gdgConfig, "", "", "")
	}

	dashboardPath := d.grafanaConf.GetPath(domain.DashboardResource, d.grafanaConf.GetOrganizationName())
	filesInDir, err := d.storage.FindAllFiles(dashboardPath, true)
	if err != nil {
		return nil, err
	}

	v1Files := make(map[string][]byte)
	v2Files := make(map[string][]byte)
	// processed collects every local dashboard identifier (UID / resource name)
	// so pruneDashboards knows what must not be deleted on the server.
	// v1 files key by board["uid"]; v2 files key by resource.metadata.name —
	// both map to the same value Grafana exposes as the dashboard UID in its APIs.
	processed := make(map[string]bool)

	for _, file := range filesInDir {
		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json dashboards are supported, skipping", "filename", file)
			continue
		}
		raw, readErr := d.storage.ReadFile(file)
		if readErr != nil {
			slog.Warn("unable to read dashboard file, skipping", "file", file, "err", readErr)
			continue
		}
		gdg := &domain.DashboardV2Gdg{}
		if err := json.Unmarshal(raw, gdg); err != nil || gdg.GdgApiVersion == "" {
			v1Files[file] = raw
			board := make(map[string]any)
			if err := json.Unmarshal(raw, &board); err == nil {
				if uid, ok := board["uid"].(string); ok && uid != "" {
					processed[uid] = true
				}
			}
		} else if gdg.GdgApiVersion == domain.GdgApiVersionV2 {
			v2Files[file] = raw
			if gdg.Resource != nil && gdg.Resource.Name != "" {
				processed[gdg.Resource.Name] = true
			}
		} else {
			v1Files[file] = raw
		}
	}

	var uploaded []string

	if serverIsV2 {
		if len(v2Files) > 0 {
			results, upErr := d.uploadDashboardsV2(filter, v2Files)
			if upErr != nil {
				slog.Warn("v2 upload encountered errors", "err", upErr)
			}
			uploaded = append(uploaded, results...)
		}
		if len(v1Files) > 0 {
			slog.Warn("v1-format dashboard files detected on a v13+ server; uploading via legacy API. "+
				"Re-download dashboards to migrate them to v2 format.",
				"count", len(v1Files))
			results, upErr := d.uploadDashboardsV1(filter, v1Files)
			if upErr != nil {
				slog.Warn("v1 fallback upload encountered errors", "err", upErr)
			}
			uploaded = append(uploaded, results...)
		}
	} else {
		if len(v1Files) > 0 {
			results, upErr := d.uploadDashboardsV1(filter, v1Files)
			if upErr != nil {
				return uploaded, upErr
			}
			uploaded = append(uploaded, results...)
		}
		if len(v2Files) > 0 {
			slog.Warn("v2-format dashboard files found on a pre-v13 server; skipping. "+
				"Upgrade to Grafana v13+ or re-download dashboards to get v1-format files.",
				"count", len(v2Files))
		}
	}

	d.pruneDashboards(filter, processed, serverIsV2)

	return uploaded, nil
}

// DeleteAllDashboards removes all monitored dashboards, routing to v1 or v2.
func (d *DashboardServiceImpl) DeleteAllDashboards(filter outbound.Filter) []string {
	if d.isV2Enabled() {
		return d.deleteAllDashboardsV2(filter)
	}
	return d.deleteAllDashboardsV1(filter)
}

// pruneDashboards deletes remote dashboards not present in the processed map.
// When serverIsV2 is true it uses the App Platform API; otherwise it uses the
// legacy search API. Both cases key on the same identifier — the dashboard
// UID / resource metadata.name — so a single map covers all local file formats.
func (d *DashboardServiceImpl) pruneDashboards(filter outbound.Filter, processed map[string]bool, serverIsV2 bool) {
	if serverIsV2 {
		client, _, err := d.dashboardV2Client()
		if err != nil {
			slog.Error("unable to create v2 client for pruning, skipping", "err", err)
			return
		}
		ctx := context.Background()
		for _, gdg := range d.listDashboardsV2(filter) {
			if processed[gdg.Resource.Name] {
				continue
			}
			slog.Info("Deleting dashboard not found in backup", "dashboard", gdg.Resource.Spec.Title)
			if delErr := client.Delete(ctx, gdg.Resource.Name, metav1.DeleteOptions{}); delErr != nil {
				slog.Error("unable to delete dashboard", "dashboard", gdg.Resource.Spec.Title, "err", delErr)
			}
		}
	} else {
		for _, item := range d.listDashboardsV1(filter) {
			if processed[item.UID] {
				continue
			}
			slog.Info("Deleting dashboard not found in backup", "folder", item.FolderTitle, "dashboard", item.Title)
			if err := d.deleteDashboard(item.Hit); err != nil {
				slog.Error("unable to delete dashboard", "folder", item.FolderTitle, "dashboard", item.Title, "err", err)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// v1 → DashboardV2Gdg mapping
// ---------------------------------------------------------------------------

// mapV1ToDashboardV2Gdg converts v1 NestedHit results into the unified
// DashboardV2Gdg wrapper so the CLI always works with one type.
func mapV1ToDashboardV2Gdg(hits []*domain.NestedHit) []*domain.DashboardV2Gdg {
	result := make([]*domain.DashboardV2Gdg, 0, len(hits))
	for _, hit := range hits {
		if hit == nil || hit.Hit == nil {
			continue
		}
		res := &domain.DashboardResourceV2{}
		res.Name = hit.UID
		res.UID = k8stypes.UID(hit.UID)
		res.Spec.Title = hit.Title
		res.Spec.Tags = hit.Tags
		result = append(result, &domain.DashboardV2Gdg{
			Resource:      res,
			NestedPath:    hit.NestedPath,
			GdgApiVersion: domain.GdgApiVersionV1,
		})
	}
	return result
}
