package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoutingTreeConstants(t *testing.T) {
	assert.Equal(t, "notifications.alerting.grafana.app", RoutingTreeGroup)
	assert.Equal(t, "v1beta1", RoutingTreeVersion)
	assert.Equal(t, "routingtrees", RoutingTreeResourceName)
	assert.Equal(t, "RoutingTree", RoutingTreeKind)
}

func newRoutingTreeResource(name, namespace string) map[string]any {
	return map[string]any{
		"kind":       "RoutingTree",
		"apiVersion": "notifications.alerting.grafana.app/v1beta1",
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]any{
			"defaults": map[string]any{
				"receiver": "devnull",
			},
			"routes": []any{
				map[string]any{"receiver": "a"},
				map[string]any{"receiver": "b"},
			},
		},
	}
}

func TestRoutingTreeGdg_Name(t *testing.T) {
	gdg := &RoutingTreeGdg{Resource: newRoutingTreeResource("user-defined", "default")}
	assert.Equal(t, "user-defined", gdg.Name())
}

func TestRoutingTreeGdg_Name_NilResource(t *testing.T) {
	gdg := &RoutingTreeGdg{}
	assert.Equal(t, "", gdg.Name())
}

func TestRoutingTreeGdg_Name_MissingMetadata(t *testing.T) {
	gdg := &RoutingTreeGdg{Resource: map[string]any{}}
	assert.Equal(t, "", gdg.Name())
}

func TestRoutingTreeGdg_Namespace(t *testing.T) {
	gdg := &RoutingTreeGdg{Resource: newRoutingTreeResource("moo", "default")}
	assert.Equal(t, "default", gdg.Namespace())
}

func TestRoutingTreeGdg_Namespace_NilResource(t *testing.T) {
	gdg := &RoutingTreeGdg{}
	assert.Equal(t, "", gdg.Namespace())
}

func TestRoutingTreeGdg_DefaultReceiver(t *testing.T) {
	gdg := &RoutingTreeGdg{Resource: newRoutingTreeResource("moo", "default")}
	assert.Equal(t, "devnull", gdg.DefaultReceiver())
}

func TestRoutingTreeGdg_DefaultReceiver_MissingSpec(t *testing.T) {
	gdg := &RoutingTreeGdg{Resource: map[string]any{}}
	assert.Equal(t, "", gdg.DefaultReceiver())
}

func TestRoutingTreeGdg_RouteCount(t *testing.T) {
	gdg := &RoutingTreeGdg{Resource: newRoutingTreeResource("moo", "default")}
	assert.Equal(t, 2, gdg.RouteCount())
}

func TestRoutingTreeGdg_RouteCount_MissingRoutes(t *testing.T) {
	gdg := &RoutingTreeGdg{Resource: map[string]any{}}
	assert.Equal(t, 0, gdg.RouteCount())
}

func TestRoutingTreeGdg_Fields(t *testing.T) {
	resource := newRoutingTreeResource("moo", "default")
	gdg := &RoutingTreeGdg{
		Resource:      resource,
		GdgApiVersion: GdgRoutingTreeApiVersion,
	}
	assert.Equal(t, "moo", gdg.Name())
	assert.Equal(t, GdgRoutingTreeApiVersion, gdg.GdgApiVersion)
	assert.Equal(t, "notifications.alerting.grafana.app/v1beta1", GdgRoutingTreeApiVersion)
}
