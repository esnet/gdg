package resources

import (
	"fmt"
	"os"
	"path/filepath"

	commonApi "github.com/esnet/gdg/internal/adapter/grafana/common"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/pkg/tools"
)

const (
	pathSeparator = string(os.PathSeparator)
)

func BuildResourceFolder(cfg *config_domain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string {
	if (resourceType == domain.DashboardResource || resourceType == domain.AlertingRulesResource) && folderName == "" {
		folderName = commonApi.DefaultFolderName
	}
	v := fmt.Sprintf("%s/%s", cfg.GetPath(resourceType, cfg.GetOrganizationName()), folderName)
	if createDestination {
		tools.CreateDestinationPath(cfg.GetPath(resourceType, cfg.GetOrganizationName()), clearOutput, v)
	}
	return v
}

// BuildResourcePath returns the full file path for a resource within the configured output directory.
// The path format is: <output_path>/<org_name>/<resource_type>/<folderName>.json
// If createDestination is true, the directory is created on disk.
func BuildResourcePath(cfg *config_domain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string {
	v := fmt.Sprintf("%s%s%s.json", cfg.GetPath(resourceType, cfg.GetOrganizationName()), pathSeparator, folderName)
	if createDestination {
		tools.CreateDestinationPath(cfg.GetPath(resourceType, cfg.GetOrganizationName()), clearOutput, filepath.Dir(v))
	}
	return v
}
