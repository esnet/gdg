package domain

import (
	"fmt"
	"path"

	"github.com/gosimple/slug"
)

type ResourceType string

const (
	ConnectionPermissionResource ResourceType = "connections-permissions"
	ConnectionResource           ResourceType = "connections"
	DashboardPermissionsResource ResourceType = "dashboards-permissions"
	DashboardResource            ResourceType = "dashboards"
	FolderPermissionResource     ResourceType = "folders-permissions"
	FolderResource               ResourceType = "folders"
	LibraryElementResource       ResourceType = "libraryelements"
	OrganizationResource         ResourceType = "organizations"
	OrganizationMetaResource     ResourceType = "org"
	TeamResource                 ResourceType = "teams"
	UserResource                 ResourceType = "users"
	TemplatesResource            ResourceType = "templates"
	SecureSecretsResource        ResourceType = "secure"
	AlertingResource             ResourceType = "alerting"
	AlertingRulesResource        ResourceType = "alerting-rules"
)

var orgNamespacedResource = map[ResourceType]bool{
	ConnectionPermissionResource: true,
	DashboardPermissionsResource: true,
	ConnectionResource:           true,
	DashboardResource:            true,
	FolderPermissionResource:     true,
	FolderResource:               true,
	LibraryElementResource:       true,
	TeamResource:                 true,
	AlertingResource:             true,
	AlertingRulesResource:        true,
}

// isNamespaced returns true if the resource type is namespaced
func (s *ResourceType) isNamespaced() bool {
	return orgNamespacedResource[*s]
}

// String returns the string representation of the resource type
func (s *ResourceType) String() string {
	return string(*s)
}

// GetPath returns the path of the resource type, if Namespaced, will delimit the path by org Id
func (s *ResourceType) GetPath(basePath string, orgName string) string {
	if s.isNamespaced() {
		orgName = slug.Make(orgName)
		return path.Join(basePath, fmt.Sprintf("%s_%s", OrganizationMetaResource, orgName), s.String())

	}
	return path.Join(basePath, s.String())
}
