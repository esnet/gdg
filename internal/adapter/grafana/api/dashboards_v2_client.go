package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/esnet/gdg/internal/domain"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// dashboardV2Client returns a namespaced dynamic client for dashboard resources,
// the resolved namespace, and the resolved API version string. The API group
// version is auto-discovered from the Grafana App Platform;
// DashboardV2Version is used when discovery fails.
func (d *DashboardServiceImpl) dashboardV2Client() (dynamic.ResourceInterface, string, string, error) {
	cfg, err := d.k8sRestConfig()
	if err != nil {
		return nil, "", "", err
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, "", "", fmt.Errorf("unable to build dynamic client: %w", err)
	}
	version := d.discoverAPIGroupVersion(domain.DashboardV2Group, domain.DashboardV2Version)
	gvr := schema.GroupVersionResource{
		Group:    domain.DashboardV2Group,
		Version:  version,
		Resource: domain.DashboardV2Resource,
	}
	ns := d.k8sNamespace()
	apiVersion := domain.DashboardV2Group + "/" + version
	return dyn.Resource(gvr).Namespace(ns), ns, apiVersion, nil
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
// toUnstructured builds the unstructured object sent to the dynamic client.
// apiVersion must match the GVR used by the dynamic client (discovered at runtime).
func toUnstructured(dr domain.DashboardResourceV2, apiVersion string) (*unstructured.Unstructured, error) {
	// Ensure TypeMeta is always set correctly before sending.
	dr.APIVersion = apiVersion
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
