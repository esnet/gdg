package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configDomain "github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/pkg/config/domain"

	"github.com/esnet/gdg/internal/tools"
	"github.com/gosimple/slug"
)

const pathSeparator = string(os.PathSeparator)

var (
	DefaultFolderName   = "General"
	searchTypeDashboard = "dash-db"
	SearchTypeFolder    = "dash-folder"
)

func GetSlug(title string) string {
	return strings.ToLower(slug.Make(title))
}

// Update the slug in the board returned
func updateSlug(board string) string {
	elements := strings.Split(board, "/")
	if len(elements) > 1 {
		return elements[len(elements)-1]
	}

	return ""
}

// getFolderFromResourcePath if a use encodes a path separator in path, we can't determine the folder name.  This strips away
// all the known components of a resource type leaving only the folder name.
func getFolderFromResourcePath(cfg *configDomain.GrafanaConfig, filePath string, resourceType domain.ResourceType, prefix string, orgName string) (string, error) {
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

func BuildResourceFolder(cfg *configDomain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string {
	if (resourceType == domain.DashboardResource || resourceType == domain.AlertingRulesResource) && folderName == "" {
		folderName = DefaultFolderName
	}
	v := fmt.Sprintf("%s/%s", cfg.GetPath(resourceType, cfg.GetOrganizationName()), folderName)
	if createDestination {
		tools.CreateDestinationPath(cfg.GetPath(resourceType, cfg.GetOrganizationName()), clearOutput, v)
	}
	return v
}

func buildResourcePath(cfg *configDomain.GrafanaConfig, folderName string, resourceType domain.ResourceType, createDestination bool, clearOutput bool) string {
	v := fmt.Sprintf("%s%s%s.json", cfg.GetPath(resourceType, cfg.GetOrganizationName()), pathSeparator, folderName)
	if createDestination {
		tools.CreateDestinationPath(cfg.GetPath(resourceType, cfg.GetOrganizationName()), clearOutput, filepath.Dir(v))
	}
	return v
}
