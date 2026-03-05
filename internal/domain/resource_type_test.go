package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceType_String(t *testing.T) {
	cases := []struct {
		rt   ResourceType
		want string
	}{
		{DashboardResource, "dashboards"},
		{FolderResource, "folders"},
		{ConnectionResource, "connections"},
		{UserResource, "users"},
		{TeamResource, "teams"},
		{OrganizationResource, "organizations"},
		{OrganizationMetaResource, "org"},
		{AlertingResource, "alerting"},
		{AlertingRulesResource, "alerting-rules"},
		{LibraryElementResource, "libraryelements"},
		{SecureSecretsResource, "secure"},
		{TemplatesResource, "templates"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.rt.String())
		})
	}
}

func TestResourceType_IsNamespaced_True(t *testing.T) {
	namespaced := []ResourceType{
		ConnectionPermissionResource,
		DashboardPermissionsResource,
		ConnectionResource,
		DashboardResource,
		FolderPermissionResource,
		FolderResource,
		LibraryElementResource,
		TeamResource,
		AlertingResource,
		AlertingRulesResource,
	}
	for _, rt := range namespaced {
		rt := rt
		t.Run(rt.String(), func(t *testing.T) {
			assert.True(t, rt.isNamespaced(), "expected %q to be namespaced", rt)
		})
	}
}

func TestResourceType_IsNamespaced_False(t *testing.T) {
	notNamespaced := []ResourceType{
		UserResource,
		OrganizationResource,
		OrganizationMetaResource,
		TemplatesResource,
		SecureSecretsResource,
	}
	for _, rt := range notNamespaced {
		rt := rt
		t.Run(rt.String(), func(t *testing.T) {
			assert.False(t, rt.isNamespaced(), "expected %q NOT to be namespaced", rt)
		})
	}
}

func TestResourceType_GetPath_NotNamespaced(t *testing.T) {
	rt := UserResource
	got := rt.GetPath("/backups", "Main Org.")
	// Not namespaced → path.Join(basePath, rt.String())
	assert.Equal(t, "/backups/users", got)
}

func TestResourceType_GetPath_Namespaced(t *testing.T) {
	rt := DashboardResource
	got := rt.GetPath("/backups", "Main Org.")
	// Namespaced → slug of "Main Org." = "main-org"
	// path.Join("/backups", "org_main-org", "dashboards")
	assert.Equal(t, "/backups/org_main-org/dashboards", got)
}

func TestResourceType_GetPath_NamespacedSpecialChars(t *testing.T) {
	rt := FolderResource
	got := rt.GetPath("output", "My Org & Partners")
	// slug.Make("My Org & Partners") = "my-org-partners"
	assert.Contains(t, got, "org_")
	assert.Contains(t, got, "folders")
}

func TestResourceType_GetPath_EmptyBase(t *testing.T) {
	rt := UserResource
	got := rt.GetPath("", "")
	assert.Equal(t, "users", got)
}

func TestResourceType_GetPath_NamespacedEmptyBase(t *testing.T) {
	rt := DashboardResource
	got := rt.GetPath("", "Main Org.")
	assert.Equal(t, "org_main-org/dashboards", got)
}
