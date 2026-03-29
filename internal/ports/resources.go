package ports

import (
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
)

type Resources interface {
	BuildResourceFolder(cfg *config_domain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string
	BuildResourcePath(cfg *config_domain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string
	GetFolderFromResourcePath(cfg *config_domain.GrafanaConfig, filePath string, resourceType domain.ResourceType, prefix string, orgName string) (string, error)
	//
	GetSlug(title string) string
	UpdateSlug(board string) string
}
