package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"sort"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"

	"github.com/gosimple/slug"
	"github.com/tidwall/pretty"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// listDashboardsV2 lists dashboards via the App Platform (/apis) dashboard.grafana.app/v2
// endpoint. It is an unexported adapter-layer helper called by DashboardServiceImpl.
func (d *DashboardServiceImpl) listDashboardsV2(filterReq outbound.Filter) []*domain.DashboardV2Gdg {
	if filterReq == nil {
		filterReq = NewDashboardFilter(d.gdgConfig, "", "", "")
	}

	client, _, err := d.dashboardV2Client()
	if err != nil {
		log.Fatal("unable to create dashboard v2 client: ", err)
	}

	// The App Platform list endpoint paginates server-side and returns a metadata.continue
	// token whenever more items remain.  Loop until the server stops returning a
	// continue token to collect every dashboard.
	ctx := context.Background()
	var items []unstructured.Unstructured
	opts := metav1.ListOptions{}
	for {
		list, listErr := client.List(ctx, opts)
		if listErr != nil {
			log.Fatal("failed to list dashboards via app platform api: ", listErr)
		}
		items = append(items, list.Items...)
		cont := list.GetContinue()
		if cont == "" {
			break
		}
		opts.Continue = cont
	}

	folderUIDMap := d.getFolderUIDEntityMap(nil)
	results := make([]*domain.DashboardV2Gdg, 0, len(items))
	for i := range items {
		dr, convErr := fromUnstructured(&items[i])
		if convErr != nil {
			slog.Error("skipping dashboard, unable to decode resource", "err", convErr)
			continue
		}
		nestedPath := d.dashboardV2FolderPath(dr, folderUIDMap)
		if !d.dashboardV2PassesFilters(filterReq, dr, nestedPath) {
			slog.Debug("skipping dashboard, failed filter check", "title", dr.Spec.Title, "folder", nestedPath)
			continue
		}
		results = append(results, &domain.DashboardV2Gdg{
			Resource:      dr,
			NestedPath:    nestedPath,
			GdgApiVersion: domain.GdgApiVersionV2,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Resource.Spec.Title < results[j].Resource.Spec.Title
	})
	return results
}

// downloadDashboardsV2 saves all dashboards via the App Platform API to the configured
// storage location. It is an unexported adapter-layer helper called by DashboardServiceImpl.
func (d *DashboardServiceImpl) downloadDashboardsV2(filterReq outbound.Filter) []string {
	listed := d.listDashboardsV2(filterReq)

	var boards []string
	for _, gdg := range listed {
		rawBoard, err := json.Marshal(gdg)
		if err != nil {
			slog.Error("unable to serialize dashboard", "dashboard", gdg.Resource.Name, "err", err)
			continue
		}

		name := slug.Make(gdg.Resource.Spec.Title)

		fileName := fmt.Sprintf("%s/%s.json",
			d.resources.BuildResourceFolder(d.grafanaConf, gdg.NestedPath, domain.DashboardResource, d.isLocal(), d.GetGlobals().ClearOutput),
			name)
		if err = d.storage.WriteFile(fileName, pretty.Pretty(rawBoard)); err != nil {
			slog.Error("Unable to save dashboard to file", "err", err, "dashboard", name)
			continue
		}
		boards = append(boards, fileName)
	}
	return boards
}

// uploadDashboardsV2 uploads dashboards via the App Platform API.
// It is an unexported adapter-layer helper called by uploadDashboardsMixed, which
// pre-reads all files and passes only v2-format entries via the files map
// (path → raw bytes), eliminating redundant storage calls in this function.
// Pruning is handled by the caller via pruneDashboards.
func (d *DashboardServiceImpl) uploadDashboardsV2(filterReq outbound.Filter, files map[string][]byte) ([]string, error) {
	if filterReq == nil {
		filterReq = NewDashboardFilter(d.gdgConfig, "", "", "")
	}

	client, ns, err := d.dashboardV2Client()
	if err != nil {
		return nil, err
	}

	folderNameUIDMap := d.getFolderNameUIDMap(d.listFolders(NewFolderFilter(d.gdgConfig)))
	seen := make(map[string]bool)

	ctx := context.Background()
	var uploaded []string
	for file, rawBoard := range files {
		gdg := &domain.DashboardV2Gdg{}
		if umErr := json.Unmarshal(rawBoard, gdg); umErr != nil {
			slog.Warn("Failed to unmarshal dashboard resource", "filename", file, "err", umErr)
			continue
		}
		dr := gdg.Resource
		if dr == nil {
			slog.Warn("Skipping file: not a v2 dashboard format (missing resource field), use v1 upload path instead", "file", file)
			continue
		}

		// Determine folder from the on-disk layout.
		folderName, folderErr := d.resources.GetFolderFromResourcePath(d.grafanaConf, file, domain.DashboardResource, d.storage.GetPrefix(), d.grafanaConf.GetOrganizationName())
		if folderErr != nil || folderName == "" {
			folderName = domain.ApiConsts.DefaultFolderName
		}

		if !d.dashboardV2PassesFilters(filterReq, dr, folderName) {
			slog.Debug("dashboard failed filter check, skipping", "file", file)
			continue
		}

		folderUID, resolveErr := d.resolveDashboardV2Folder(folderName, folderNameUIDMap)
		if resolveErr != nil {
			slog.Warn("unable to resolve/create folder, skipping", "folder", folderName, "err", resolveErr)
			continue
		}
		dr.SetFolderUID(folderUID)

		name := dr.Name
		if seen[name] {
			slog.Warn("dashboard with same name already processed, skipping", "name", name, "file", file)
			continue
		}
		seen[name] = true

		sanitizeForApply(dr, ns)

		obj, convErr := toUnstructured(*dr)
		if convErr != nil {
			slog.Warn("unable to encode dashboard for upload", "file", file, "err", convErr)
			continue
		}

		if upErr := upsertDashboardV2(ctx, client, name, obj); upErr != nil {
			slog.Info("error uploading dashboard", "dashboard-filename", file, "err", upErr)
			continue
		}
		uploaded = append(uploaded, file)
	}

	return uploaded, nil
}

// deleteAllDashboardsV2 removes all dashboards matching the filter via the App Platform API.
// It is an unexported adapter-layer helper called by DashboardServiceImpl.
func (d *DashboardServiceImpl) deleteAllDashboardsV2(filterReq outbound.Filter) []string {
	client, _, err := d.dashboardV2Client()
	if err != nil {
		log.Fatal("unable to create dashboard v2 client: ", err)
	}

	ctx := context.Background()
	deleted := make([]string, 0)
	for _, gdg := range d.listDashboardsV2(filterReq) {
		if delErr := client.Delete(ctx, gdg.Resource.Name, metav1.DeleteOptions{}); delErr != nil {
			slog.Warn("Unable to remove dashboard", slog.String("title", gdg.Resource.Spec.Title), slog.String("name", gdg.Resource.Name))
			continue
		}
		deleted = append(deleted, gdg.Resource.Spec.Title)
	}
	return deleted
}

// dashboardV2FolderPath resolves the nested folder path for a dashboard resource
// from its folder-UID annotation, mirroring the v1 nesting behavior.
func (d *DashboardServiceImpl) dashboardV2FolderPath(dr *domain.DashboardResourceV2, folderUIDMap map[string]*domain.NestedHit) string {
	folderUID := dr.FolderUID()
	if folderUID == "" {
		return domain.ApiConsts.DefaultFolderName
	}
	entity, ok := folderUIDMap[folderUID]
	if !ok || entity == nil || entity.Hit == nil {
		return domain.ApiConsts.DefaultFolderName
	}
	return getNestedFolder(entity.Title, entity.UID, folderUIDMap)
}

// dashboardV2PassesFilters applies the folder, tag and title filters to a resource.
// Tag/title filters reuse the existing JSON readers by presenting the fields they
// expect at the top level.
func (d *DashboardServiceImpl) dashboardV2PassesFilters(filterReq outbound.Filter, dr *domain.DashboardResourceV2, nestedPath string) bool {
	if dr == nil {
		return false
	}
	ctx := context.Background()

	folderValid := false
	if d.grafanaConf.GetDashboardSettings().IgnoreFilters && !d.grafanaConf.IsFilterSet() {
		folderValid = true
	} else if filterReq.Validate(ctx, domain.FolderFilter, map[string]any{NestedDashFolderName: nestedPath}) {
		folderValid = true
	}
	if !folderValid {
		return false
	}

	filterJSON, err := json.Marshal(map[string]any{
		"uid":  dr.Name,
		"tags": dr.Spec.Tags,
	})
	if err != nil {
		return false
	}
	if !filterReq.Validate(ctx, domain.TagsFilter, filterJSON) {
		return false
	}
	return filterReq.Validate(ctx, domain.DashFilter, filterJSON)
}

// resolveDashboardV2Folder resolves a nested folder path to a folder UID, creating
// any missing folders. The default (General) folder maps to an empty UID (root).
func (d *DashboardServiceImpl) resolveDashboardV2Folder(folderName string, folderNameUIDMap map[string]string) (string, error) {
	if folderName == "" || folderName == domain.ApiConsts.DefaultFolderName {
		return "", nil
	}
	if uid, ok := folderNameUIDMap[folderName]; ok {
		return uid, nil
	}
	newFolders, err := d.createdFolders(folderName)
	if err != nil {
		return "", err
	}
	maps.Copy(folderNameUIDMap, newFolders)
	return folderNameUIDMap[folderName], nil
}

// sanitizeForApply strips server-managed metadata so a stored resource can be
// re-applied via the App Platform API without conflicting on immutable fields.
func sanitizeForApply(dr *domain.DashboardResourceV2, namespace string) {
	dr.Namespace = namespace
	dr.ResourceVersion = ""
	dr.UID = ""
	dr.Generation = 0
	dr.CreationTimestamp = metav1.Time{}
	dr.ManagedFields = nil
	dr.Status = nil
}
