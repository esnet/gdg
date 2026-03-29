package resources

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/tools"
	"github.com/gosimple/slug"
)

const (
	pathSeparator = string(os.PathSeparator)
)

func NewHelpers() ports.Resources {
	return &Helpers{}
}

type Helpers struct{}

func (h Helpers) BuildResourceFolder(cfg *config_domain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string {
	//func BuildResourceFolder(cfg *config_domain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string {
	if (resourceType == domain.DashboardResource || resourceType == domain.AlertingRulesResource) && folderName == "" {
		folderName = domain.ApiConsts.DefaultFolderName
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
func (h Helpers) BuildResourcePath(cfg *config_domain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string {
	//func BuildResourcePath(cfg *config_domain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string {
	v := fmt.Sprintf("%s%s%s.json", cfg.GetPath(resourceType, cfg.GetOrganizationName()), pathSeparator, folderName)
	if createDestination {
		tools.CreateDestinationPath(cfg.GetPath(resourceType, cfg.GetOrganizationName()), clearOutput, filepath.Dir(v))
	}
	return v
}

// GetSlug converts the given title into a lowercase, URL-friendly slug.
func (h Helpers) GetSlug(title string) string {
	return strings.ToLower(slug.Make(title))
}

// UpdateSlug extracts and returns the last segment from a slash-delimited board path.
// Returns an empty string if the path contains no slash separator.
func (h Helpers) UpdateSlug(board string) string {
	elements := strings.Split(board, "/")
	if len(elements) > 1 {
		return elements[len(elements)-1]
	}

	return ""
}

func (h Helpers) GetFolderFromResourcePath(cfg *config_domain.GrafanaConfig, filePath string, resourceType domain.ResourceType, prefix string, orgName string) (string, error) {
	basePath := fmt.Sprintf("%s/", cfg.GetPath(resourceType, orgName))
	if prefix != "" {
		if prefix[0] != filePath[0] && prefix[0] == '/' {
			prefix = prefix[1:]
		}
		basePath = filepath.Join(prefix, basePath)
	}
	if basePath[0] == os.PathSeparator {
		basePath = basePath[1:]
	}

	folderName := strings.Replace(filePath, basePath, "", 1)
	ndx := strings.LastIndex(folderName, string(os.PathSeparator))
	if ndx != -1 {
		folderName = folderName[0:ndx]
		if len(folderName) > 1 && folderName[0] == os.PathSeparator {
			folderName = folderName[1:]
		}
		return folderName, nil
	}
	return "", errors.New("unable to parse resource to retrieve folder name")
}
