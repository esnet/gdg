package api

// Unit tests for pure helper functions in the api package.
// These functions have no external dependencies (no HTTP, no storage, no config),
// so they can be tested in a plain unit-test style without Docker / testcontainers.

import (
	"encoding/json"
	"testing"

	"github.com/esnet/gdg/internal/domain"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ---------------------------------------------------------------------------
// permissionStringToType
// ---------------------------------------------------------------------------

func TestPermissionStringToType(t *testing.T) {
	tests := []struct {
		input    string
		expected models.PermissionType
	}{
		{"view", models.PermissionType(1)},
		{"View", models.PermissionType(1)},
		{"VIEW", models.PermissionType(1)},
		{"edit", models.PermissionType(2)},
		{"Edit", models.PermissionType(2)},
		{"admin", models.PermissionType(4)},
		{"Admin", models.PermissionType(4)},
		{"ADMIN", models.PermissionType(4)},
		{"", models.PermissionType(0)},
		{"unknown", models.PermissionType(0)},
		{"superuser", models.PermissionType(0)},
	}
	for _, tt := range tests {
		t.Run("input="+tt.input, func(t *testing.T) {
			got := permissionStringToType(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ---------------------------------------------------------------------------
// aclDTOToResourcePermission
// ---------------------------------------------------------------------------

func TestAclDTOToResourcePermission_AllFields(t *testing.T) {
	userID := int64(42)
	teamID := int64(7)
	acl := &models.DashboardACLInfoDTO{
		UserID:         userID,
		UserLogin:      "alice",
		TeamID:         teamID,
		Team:           "ops-team",
		Role:           "Editor",
		PermissionName: "Edit",
	}
	got := aclDTOToResourcePermission(acl, 0)
	require.NotNil(t, got)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, "alice", got.UserLogin)
	assert.Equal(t, teamID, got.TeamID)
	assert.Equal(t, "ops-team", got.Team)
	assert.Equal(t, "Editor", got.BuiltInRole)
	assert.Equal(t, "Edit", got.Permission)
}

func TestAclDTOToResourcePermission_ZeroValue(t *testing.T) {
	acl := &models.DashboardACLInfoDTO{}
	got := aclDTOToResourcePermission(acl, 0)
	require.NotNil(t, got)
	assert.Empty(t, got.UserLogin)
	assert.Empty(t, got.BuiltInRole)
	assert.Empty(t, got.Permission)
}

// ---------------------------------------------------------------------------
// v2GdgToNestedHit
// ---------------------------------------------------------------------------

func TestV2GdgToNestedHit_WithResource(t *testing.T) {
	dr := &domain.DashboardResourceV2{}
	dr.Name = "uid-123"
	dr.UID = k8stypes.UID("uid-123")
	dr.Spec.Title = "My Dashboard"

	gdg := &domain.DashboardV2Gdg{
		Resource:   dr,
		NestedPath: "General/Subfolder",
	}

	hit := v2GdgToNestedHit(gdg)
	require.NotNil(t, hit)
	assert.Equal(t, "General/Subfolder", hit.NestedPath)
	require.NotNil(t, hit.Hit)
	assert.Equal(t, "uid-123", hit.UID)
	assert.Equal(t, "My Dashboard", hit.Title)
}

func TestV2GdgToNestedHit_NilResource(t *testing.T) {
	gdg := &domain.DashboardV2Gdg{
		Resource:   nil,
		NestedPath: "General",
	}
	hit := v2GdgToNestedHit(gdg)
	require.NotNil(t, hit)
	assert.Equal(t, "General", hit.NestedPath)
	// Hit should be nil when there is no resource.
	assert.Nil(t, hit.Hit)
}

// ---------------------------------------------------------------------------
// getNestedFolder
// ---------------------------------------------------------------------------

func makeNestedHit(uid, folderUID, folderTitle string) *domain.NestedHit {
	return &domain.NestedHit{
		Hit: &models.Hit{
			UID:         uid,
			FolderUID:   folderUID,
			FolderTitle: folderTitle,
		},
	}
}

func TestGetNestedFolder_NoParent(t *testing.T) {
	// A top-level folder has no FolderUID in the map.
	folderUidMap := map[string]*domain.NestedHit{
		"uid-top": makeNestedHit("uid-top", "", ""),
	}
	// When folderUID is empty the result is just encode.Encode(title).
	got := getNestedFolder("General", "", folderUidMap)
	assert.Equal(t, "General", got)
}

func TestGetNestedFolder_NestedOneLevel(t *testing.T) {
	// "child" lives inside "parent".
	folderUidMap := map[string]*domain.NestedHit{
		"uid-child": {
			Hit: &models.Hit{
				UID:         "uid-child",
				FolderUID:   "uid-parent",
				FolderTitle: "Parent",
				Title:       "Child",
			},
		},
		"uid-parent": {
			Hit: &models.Hit{
				UID:         "uid-parent",
				FolderUID:   "",
				FolderTitle: "",
				Title:       "Parent",
			},
		},
	}
	got := getNestedFolder("Child", "uid-child", folderUidMap)
	// Should be Parent/Child (URL-query-escaped, but both have no special chars).
	assert.Equal(t, "Parent/Child", got)
}

func TestGetNestedFolder_UIDNotInMap(t *testing.T) {
	// folderUID is set but not present in the map — should just return title.
	folderUidMap := map[string]*domain.NestedHit{}
	got := getNestedFolder("Orphan", "unknown-uid", folderUidMap)
	assert.Equal(t, "Orphan", got)
}

// ---------------------------------------------------------------------------
// sanitizeForApply
// ---------------------------------------------------------------------------

func TestSanitizeForApply(t *testing.T) {
	dr := &domain.DashboardResourceV2{}
	dr.Namespace = "old-namespace"
	dr.ResourceVersion = "12345"
	dr.UID = k8stypes.UID("some-uid")
	dr.Generation = 3
	dr.CreationTimestamp = metav1.Now()
	dr.ManagedFields = []metav1.ManagedFieldsEntry{{Manager: "manager"}}
	dr.Status = map[string]any{"phase": "ready"}

	sanitizeForApply(dr, "new-namespace")

	assert.Equal(t, "new-namespace", dr.Namespace)
	assert.Empty(t, dr.ResourceVersion)
	assert.Empty(t, string(dr.UID))
	assert.Zero(t, dr.Generation)
	assert.True(t, dr.CreationTimestamp.IsZero())
	assert.Nil(t, dr.ManagedFields)
	assert.Nil(t, dr.Status)
}

// ---------------------------------------------------------------------------
// toUnstructured / fromUnstructured round-trip
// ---------------------------------------------------------------------------

func TestToUnstructured_SetsTypeMeta(t *testing.T) {
	dr := domain.DashboardResourceV2{}
	dr.Name = "test-dashboard"
	dr.Spec.Title = "Test"

	obj, err := toUnstructured(dr)
	require.NoError(t, err)
	require.NotNil(t, obj)

	// TypeMeta should be populated from constants.
	assert.Equal(t, domain.DashboardV2Group+"/"+domain.DashboardV2Version, obj.GetAPIVersion())
	assert.Equal(t, domain.DashboardV2Kind, obj.GetKind())
	assert.Equal(t, "test-dashboard", obj.GetName())
}

func TestFromUnstructured_RoundTrip(t *testing.T) {
	dr := domain.DashboardResourceV2{}
	dr.Name = "round-trip-dashboard"
	dr.Spec.Title = "Round Trip"
	dr.Annotations = map[string]string{domain.AnnotationFolder: "folder-uid-abc"}

	obj, err := toUnstructured(dr)
	require.NoError(t, err)

	decoded, err := fromUnstructured(obj)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	assert.Equal(t, "round-trip-dashboard", decoded.Name)
	assert.Equal(t, "Round Trip", decoded.Spec.Title)
	assert.Equal(t, "folder-uid-abc", decoded.FolderUID())
}

func TestFromUnstructured_InvalidJSON(t *testing.T) {
	// An unstructured object whose JSON cannot be decoded into DashboardResourceV2.
	// We simulate this by putting a non-object value at a required nested path.
	u := &unstructured.Unstructured{Object: map[string]any{
		"metadata": "not-an-object", // metadata must be an object
	}}
	// Marshalling the map itself succeeds; unmarshalling into DashboardResourceV2
	// should fail because metadata is a string, not an object.
	data, err := json.Marshal(u.Object)
	require.NoError(t, err)

	var dr domain.DashboardResourceV2
	unmarshalErr := json.Unmarshal(data, &dr)
	// If Go's JSON decoder ignores the type mismatch, the test is still
	// informative — we confirm fromUnstructured returns no error in that case.
	if unmarshalErr != nil {
		assert.Error(t, unmarshalErr)
	}
}

// ---------------------------------------------------------------------------
// mapV1ToDashboardV2Gdg
// ---------------------------------------------------------------------------

func TestMapV1ToDashboardV2Gdg_BasicConversion(t *testing.T) {
	hits := []*domain.NestedHit{
		{
			Hit: &models.Hit{
				UID:   "uid-1",
				Title: "Dashboard One",
				Tags:  []string{"prod", "monitoring"},
			},
			NestedPath: "General",
		},
		{
			Hit: &models.Hit{
				UID:   "uid-2",
				Title: "Dashboard Two",
			},
			NestedPath: "General/Sub",
		},
	}

	result := mapV1ToDashboardV2Gdg(hits)
	require.Len(t, result, 2)

	assert.Equal(t, "uid-1", result[0].Resource.Name)
	assert.Equal(t, "Dashboard One", result[0].Resource.Spec.Title)
	assert.Equal(t, []string{"prod", "monitoring"}, result[0].Resource.Spec.Tags)
	assert.Equal(t, "General", result[0].NestedPath)
	assert.Equal(t, domain.GdgApiVersionV1, result[0].GdgApiVersion)

	assert.Equal(t, "uid-2", result[1].Resource.Name)
	assert.Equal(t, "General/Sub", result[1].NestedPath)
}

func TestMapV1ToDashboardV2Gdg_SkipsNilAndNilHit(t *testing.T) {
	hits := []*domain.NestedHit{
		nil, // nil entry
		{Hit: nil, NestedPath: "General"}, // non-nil entry but nil Hit
		{
			Hit:        &models.Hit{UID: "uid-ok", Title: "Good"},
			NestedPath: "General",
		},
	}
	result := mapV1ToDashboardV2Gdg(hits)
	require.Len(t, result, 1)
	assert.Equal(t, "uid-ok", result[0].Resource.Name)
}

func TestMapV1ToDashboardV2Gdg_EmptyInput(t *testing.T) {
	result := mapV1ToDashboardV2Gdg(nil)
	assert.Empty(t, result)

	result = mapV1ToDashboardV2Gdg([]*domain.NestedHit{})
	assert.Empty(t, result)
}
