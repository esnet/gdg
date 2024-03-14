package config

import (
	"fmt"
	"github.com/gosimple/slug"
	"path"
)

type ResourceType string

const (
	ConnectionPermissionResource ResourceType = "connections-permissions"
	ConnectionResource           ResourceType = "connections"
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
)

var orgNamespacedResource = map[ResourceType]bool{
	ConnectionPermissionResource: true,
	ConnectionResource:           true,
	DashboardResource:            true,
	FolderPermissionResource:     true,
	FolderResource:               true,
	LibraryElementResource:       true,
	TeamResource:                 true,
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
func (s *ResourceType) GetPath(basePath string) string {
	if s.isNamespaced() {
		orgName := slug.Make(Config().GetDefaultGrafanaConfig().GetOrganizationName())
		return path.Join(basePath, fmt.Sprintf("%s_%s", OrganizationMetaResource, orgName), s.String())

	}
	return path.Join(basePath, s.String())
}
