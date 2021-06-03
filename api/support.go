package api

import (
	"strings"

	"github.com/netsage-project/grafana-dashboard-manager/config"
)

type DashboardFilter struct {
	FolderFilter string // Name of Folder
	DashFilter   string //name of dashboard
}

//GetFolders splits the comma delimited folder list and returns a slice
func (s *DashboardFilter) GetFolders() []string {
	if s.FolderFilter == "" {
		return config.GetDefaultGrafanaConfig().GetMonitoredFolders()
	}
	s.FolderFilter = quoteRegex.ReplaceAllString(s.FolderFilter, "")

	return strings.Split(s.FolderFilter, ",")
}

func (s DashboardFilter) ValidateDashboard(dashUid string) bool {
	if s.DashFilter == "" {
		return true
	}
	return dashUid == s.DashFilter
}

func (s DashboardFilter) Validate(folder, dashUid string) bool {
	return s.ValidateDashboard(dashUid) && s.ValidateFolder(folder)
}

func (s DashboardFilter) ValidateFolder(folder string) bool {
	if s.FolderFilter == "" {
		return true
	}
	return folder == s.FolderFilter
}

type DatasourceFilter struct {
	Name string //name of datasource
}

func (s DatasourceFilter) ValidateDatasource(name string) bool {
	if s.Name == "" {
		return true
	}
	return name == s.Name
}
