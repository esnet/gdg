package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestDashboardV2Constants(t *testing.T) {
	assert.Equal(t, "dashboard.grafana.app", DashboardV2Group)
	assert.Equal(t, "v2", DashboardV2Version)
	assert.Equal(t, "dashboards", DashboardV2Resource)
	assert.Equal(t, "Dashboard", DashboardV2Kind)
	assert.Equal(t, "grafana.app/folder", AnnotationFolder)
	assert.Equal(t, "v1", GdgApiVersionV1)
	assert.Equal(t, "dashboard.grafana.app/v2", GdgApiVersionV2)
}

// ---------------------------------------------------------------------------
// DashboardResourceV2.FolderUID
// ---------------------------------------------------------------------------

func TestFolderUID_NilAnnotations(t *testing.T) {
	dr := &DashboardResourceV2{}
	assert.Equal(t, "", dr.FolderUID())
}

func TestFolderUID_NoAnnotationKey(t *testing.T) {
	dr := &DashboardResourceV2{}
	dr.Annotations = map[string]string{"some.other/annotation": "value"}
	assert.Equal(t, "", dr.FolderUID())
}

func TestFolderUID_AnnotationSet(t *testing.T) {
	dr := &DashboardResourceV2{}
	dr.Annotations = map[string]string{AnnotationFolder: "folder-uid-123"}
	assert.Equal(t, "folder-uid-123", dr.FolderUID())
}

// ---------------------------------------------------------------------------
// DashboardResourceV2.SetFolderUID
// ---------------------------------------------------------------------------

func TestSetFolderUID_EmptyUID_NilAnnotations(t *testing.T) {
	// Deleting from a nil map should not panic.
	dr := &DashboardResourceV2{}
	assert.NotPanics(t, func() { dr.SetFolderUID("") })
	// Annotations remain nil because we never had to initialise them.
	assert.Nil(t, dr.Annotations)
}

func TestSetFolderUID_EmptyUID_ClearsExisting(t *testing.T) {
	dr := &DashboardResourceV2{}
	dr.Annotations = map[string]string{AnnotationFolder: "old-uid"}
	dr.SetFolderUID("")
	_, exists := dr.Annotations[AnnotationFolder]
	assert.False(t, exists, "annotation should be removed when uid is empty")
}

func TestSetFolderUID_SetsAnnotation(t *testing.T) {
	dr := &DashboardResourceV2{}
	dr.SetFolderUID("new-folder-uid")
	require.NotNil(t, dr.Annotations)
	assert.Equal(t, "new-folder-uid", dr.Annotations[AnnotationFolder])
}

func TestSetFolderUID_InitialisesAnnotationsMap(t *testing.T) {
	dr := &DashboardResourceV2{}
	assert.Nil(t, dr.Annotations)
	dr.SetFolderUID("some-uid")
	require.NotNil(t, dr.Annotations)
	assert.Equal(t, "some-uid", dr.FolderUID())
}

func TestSetFolderUID_OverwritesExisting(t *testing.T) {
	dr := &DashboardResourceV2{}
	dr.Annotations = map[string]string{AnnotationFolder: "old"}
	dr.SetFolderUID("new")
	assert.Equal(t, "new", dr.FolderUID())
}

// ---------------------------------------------------------------------------
// DashboardV2Gdg
// ---------------------------------------------------------------------------

func TestDashboardV2Gdg_Fields(t *testing.T) {
	dr := &DashboardResourceV2{}
	dr.Name = "my-dashboard"

	gdg := &DashboardV2Gdg{
		Resource:      dr,
		NestedPath:    "General/Subfolder",
		GdgApiVersion: GdgApiVersionV2,
	}

	assert.Equal(t, "my-dashboard", gdg.Resource.Name)
	assert.Equal(t, "General/Subfolder", gdg.NestedPath)
	assert.Equal(t, GdgApiVersionV2, gdg.GdgApiVersion)
}
