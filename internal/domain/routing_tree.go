package domain

// App Platform (Kubernetes-style) coordinates for the Notification Policy
// "RoutingTree" API served under /apis. Unlike the legacy
// /api/v1/provisioning/policies endpoint (a single flat policy tree per org,
// see AlertPolicies), this API supports multiple named routing trees per
// namespace/org and exposes the full recursive route structure.
const (
	RoutingTreeGroup        = "notifications.alerting.grafana.app"
	RoutingTreeVersion      = "v1beta1" // discover served versions via GET /apis/notifications.alerting.grafana.app
	RoutingTreeResourceName = "routingtrees"
	RoutingTreeKind         = "RoutingTree"

	// GdgRoutingTreeApiVersion tags a stored routing tree as originating from
	// the App Platform /apis endpoint.
	GdgRoutingTreeApiVersion = RoutingTreeGroup + "/" + RoutingTreeVersion
)

// RoutingTreeGdg is the GDG-internal wrapper around a K8s RoutingTree resource.
// Resource holds the raw unstructured object exactly as returned by the App
// Platform API (metadata, spec with its full recursive routes, etc.), so no
// hand-written schema is required to round-trip it to disk.
// GdgApiVersion records the API version in use at download time.
type RoutingTreeGdg struct {
	Resource      map[string]any `json:"resource"`
	GdgApiVersion string         `json:"gdg_api_version"`
}

// Name returns the routing tree's metadata.name, or "" if unavailable.
func (r *RoutingTreeGdg) Name() string {
	return r.nestedString("metadata", "name")
}

// Namespace returns the routing tree's metadata.namespace, or "" if unavailable.
func (r *RoutingTreeGdg) Namespace() string {
	return r.nestedString("metadata", "namespace")
}

// DefaultReceiver returns the routing tree's spec.defaults.receiver, or "" if unavailable.
func (r *RoutingTreeGdg) DefaultReceiver() string {
	return r.nestedString("spec", "defaults", "receiver")
}

// RouteCount returns the number of top-level routes defined in spec.routes.
func (r *RoutingTreeGdg) RouteCount() int {
	if r.Resource == nil {
		return 0
	}
	spec, ok := r.Resource["spec"].(map[string]any)
	if !ok {
		return 0
	}
	routes, ok := spec["routes"].([]any)
	if !ok {
		return 0
	}
	return len(routes)
}

// nestedString walks the resource map following the given keys and returns
// the string found there, or "" if any step is missing or not the right type.
func (r *RoutingTreeGdg) nestedString(keys ...string) string {
	if r.Resource == nil {
		return ""
	}
	var current any = r.Resource
	for _, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = m[key]
		if !ok {
			return ""
		}
	}
	s, _ := current.(string)
	return s
}
