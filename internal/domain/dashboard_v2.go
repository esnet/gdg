package domain

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// App Platform (Kubernetes-style) coordinates for the v2 Dashboard API served
// under /apis. These replace the deprecated /api/dashboards endpoints.
const (
	DashboardV2Group    = "dashboard.grafana.app"
	DashboardV2Version  = "v2" // GA in Grafana 13; discover served versions via GET /apis/dashboard.grafana.app
	DashboardV2Resource = "dashboards"
	DashboardV2Kind     = "Dashboard"

	// AnnotationFolder holds the parent folder UID on a dashboard resource.
	AnnotationFolder = "grafana.app/folder"

	// GdgApiVersionV1 tags a stored dashboard as originating from the legacy /api/dashboards endpoint.
	GdgApiVersionV1 = "v1"
	// GdgApiVersionV2 tags a stored dashboard as originating from the App Platform /apis endpoint.
	GdgApiVersionV2 = DashboardV2Group + "/" + DashboardV2Version
)

// DashboardResourceV2 is the pure App Platform wire envelope for a dashboard.
// It contains only the fields that Grafana sends and receives — no GDG-internal
// bookkeeping. This makes marshalling/unmarshalling clean and removes the need
// for the intermediate dashboardV2Payload type in the adapter layer.
type DashboardResourceV2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              dashboardv2.Dashboard `json:"spec"`
	Status            map[string]any        `json:"status,omitempty"`
}

// FolderUID returns the parent folder UID from the resource annotations, if any.
// An empty string means the dashboard lives in the default (General) folder.
func (d *DashboardResourceV2) FolderUID() string {
	if d.Annotations == nil {
		return ""
	}
	return d.Annotations[AnnotationFolder]
}

// SetFolderUID sets the parent folder UID annotation on the resource. An empty
// value clears it, placing the dashboard in the default (General) folder.
func (d *DashboardResourceV2) SetFolderUID(uid string) {
	if uid == "" {
		delete(d.Annotations, AnnotationFolder)
		return
	}
	if d.Annotations == nil {
		d.Annotations = make(map[string]string)
	}
	d.Annotations[AnnotationFolder] = uid
}

// DashboardV2Gdg is the GDG-internal wrapper around a DashboardResourceV2.
// It carries bookkeeping fields that are written to local storage alongside the
// resource but are never sent to the Grafana API.
//
//   - Resource      holds the pure API envelope (what Grafana sends/receives).
//   - NestedPath    is the resolved folder path used for the on-disk storage layout.
//   - GdgApiVersion records the API version in use at download time
//     (GdgApiVersionV1 or GdgApiVersionV2).  At upload time this is compared
//     against the target server to detect format mismatches early.
type DashboardV2Gdg struct {
	Resource      *DashboardResourceV2 `json:"resource"`
	NestedPath    string               `json:"nested_path"`
	GdgApiVersion string               `json:"gdg_api_version"`
}
