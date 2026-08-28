package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/esnet/gdg/internal/adapter/grafana/extended"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// dashboardV2GVR is the group/version/resource for the App Platform dashboard API.
var dashboardV2GVR = schema.GroupVersionResource{
	Group:    domain.DashboardV2Group,
	Version:  domain.DashboardV2Version,
	Resource: domain.DashboardV2Resource,
}

// k8sRestConfig maps GDG's existing Grafana connection settings onto a
// Kubernetes rest.Config pointed at the Grafana App Platform (/apis) endpoint.
// The same service-account token / basic-auth credentials used by the legacy
// client authenticate here as well.
func (d *DashboardServiceImpl) k8sRestConfig() (*rest.Config, error) {
	u, err := url.Parse(d.grafanaConf.GetURL())
	if err != nil {
		return nil, fmt.Errorf("invalid Grafana URL %q: %w", d.grafanaConf.GetURL(), err)
	}

	// Host carries the scheme, host and any sub-path prefix (e.g. /grafana);
	// APIPath appends the App Platform root so requests hit <base>/apis/...
	host := strings.TrimSuffix(fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path), "/")

	cfg := &rest.Config{
		Host:    host,
		APIPath: "apis",
	}

	if token := d.grafanaConf.GetAPIToken(); token != "" {
		cfg.BearerToken = token
	} else {
		cfg.Username = d.grafanaConf.UserName
		cfg.Password = d.grafanaConf.GetPassword()
	}

	if d.gdgConfig.IgnoreSSL() {
		cfg.TLSClientConfig = rest.TLSClientConfig{Insecure: true}
	}

	return cfg, nil
}

// k8sNamespace maps the active org to an App Platform namespace (org 1 ->
// "default", org N -> "org-N"), replacing the legacy X-Grafana-Org-Id header.
// Token/anonymous auth is pinned to one org that /api/user/orgs can't see
// (401/403), so those resolve it via GET /api/org instead — getting this
// wrong isn't cosmetic; it makes the App Platform reject the namespace
// outright. Basic/admin auth can switch orgs, so it still resolves
// organization_name to an ID via /api/user/orgs.
func (d *DashboardServiceImpl) k8sNamespace() string {
	orgID := int64(config_domain.DefaultOrganizationId)

	switch {
	case !d.grafanaConf.IsBasicAuth():
		// fatal=false: same /api/org call and fail-soft policy as
		// resolvePinnedOrgIdentity, via the shared helper.
		if org, err := d.getCurrentOrg(d.GetClient(), false); err == nil && org != nil {
			orgID = org.ID
		} else {
			slog.Warn("unable to determine the current organization via /api/org; defaulting to org 1/\"default\"", "err", err)
		}
	case d.grafanaConf.OrganizationName != "":
		if id, err := extended.NewExtendedApi(d.gdgConfig).GetConfiguredOrgId(d.grafanaConf.OrganizationName); err == nil {
			orgID = id
		}
	}

	if orgID <= config_domain.DefaultOrganizationId {
		return "default"
	}
	return fmt.Sprintf("org-%d", orgID)
}

// dashboardV2Client returns a namespaced dynamic client for dashboard resources
// along with the resolved namespace.
func (d *DashboardServiceImpl) dashboardV2Client() (dynamic.ResourceInterface, string, error) {
	cfg, err := d.k8sRestConfig()
	if err != nil {
		return nil, "", err
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, "", fmt.Errorf("unable to build dynamic client: %w", err)
	}
	ns := d.k8sNamespace()
	return dyn.Resource(dashboardV2GVR).Namespace(ns), ns, nil
}

// upsertDashboardV2 creates the dashboard resource, falling back to an update if
// it already exists. We deliberately avoid server-side Apply here: the dashboard
// spec is a preserve-unknown-fields schema (it legitimately carries legacy keys
// such as $$hashKey / legacyOptions), and Apply's typed-patch step fails to build
// against it. Create/Update submit the JSON directly and preserve the full spec.
func upsertDashboardV2(ctx context.Context, client dynamic.ResourceInterface, name string, obj *unstructured.Unstructured) error {
	if _, err := client.Create(ctx, obj, metav1.CreateOptions{}); err == nil {
		return nil
	} else if !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Resource exists: carry over the current resourceVersion for optimistic
	// concurrency, then replace it.
	existing, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to fetch existing dashboard %q: %w", name, err)
	}
	obj.SetResourceVersion(existing.GetResourceVersion())
	if _, err := client.Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

// toUnstructured builds the unstructured object sent to the dynamic client directly
// from DashboardResourceV2. Since DashboardResourceV2 is now the pure API envelope
// (no GDG-internal fields), it can be marshalled directly without an intermediate
// payload type. Routing through encoding/json ensures the foundation-sdk union types
// (VariableKind, Element, Layout, scalar-or-X wrappers, ...) are serialized via
// their custom MarshalJSON.
func toUnstructured(dr domain.DashboardResourceV2) (*unstructured.Unstructured, error) {
	// Ensure TypeMeta is always set correctly before sending.
	dr.APIVersion = domain.DashboardV2Group + "/" + domain.DashboardV2Version
	dr.Kind = domain.DashboardV2Kind

	data, err := json.Marshal(dr)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal dashboard: %w", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("unable to build unstructured object: %w", err)
	}
	return &unstructured.Unstructured{Object: obj}, nil
}

// fromUnstructured converts a dynamic-client object into a typed DashboardResourceV2.
// It routes through encoding/json so the foundation-sdk union types are decoded
// via their custom UnmarshalJSON, which the reflection-based runtime converter skips.
func fromUnstructured(u *unstructured.Unstructured) (*domain.DashboardResourceV2, error) {
	data, err := json.Marshal(u.Object)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal unstructured object: %w", err)
	}
	var dr domain.DashboardResourceV2
	if err := json.Unmarshal(data, &dr); err != nil {
		return nil, fmt.Errorf("unable to decode dashboard: %w", err)
	}
	return &dr, nil
}
