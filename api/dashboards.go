package api

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/config"
	gapi "github.com/esnet/grafana-swagger-api-golang"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/dashboards"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/folders"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/orgs"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/search"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/signed_in_user"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/apphelpers"

	"github.com/tidwall/pretty"

	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

// DashboardsApi Contract definition
type DashboardsApi interface {
	ListDashboards(filter Filter) []*models.Hit
	ImportDashboards(filter Filter) []string
	ExportDashboards(filter Filter)
	DeleteAllDashboards(filter Filter) []string
}

// ListDashboards List all dashboards optionally filtered by folder name. If folderFilters
// is blank, defaults to the configured Monitored folders
func (s *DashNGoImpl) ListDashboards(filters Filter) []*models.Hit {

	var orgsPayload []*models.OrgDTO
	orgList, err := s.client.Orgs.SearchOrgs(orgs.NewSearchOrgsParams(), s.getAdminAuth())
	if err != nil {
		log.Warnf("Error getting organizations: %s", err.Error())
		orgsPayload = make([]*models.OrgDTO, 0)
	} else {
		orgsPayload = orgList.GetPayload()
	}
	if s.grafanaConf.Organization != "" {
		var ID int64
		for _, org := range orgsPayload {
			log.Errorf("%d %s\n", org.ID, org.Name)
			if org.Name == s.grafanaConf.Organization {
				ID = org.ID
			}
		}
		if ID > 0 {
			params := signed_in_user.NewUserSetUsingOrgParams()
			params.OrgID = ID
			status, err := s.client.SignedInUser.UserSetUsingOrg(params, s.getAuth())
			if err != nil {
				log.Fatalf("%s for %v\n", err, status)
			}
		}
	}

	// Fallback on defaults
	if filters == nil {
		filters = &DashboardFilter{}
	}

	var boardsList = make([]*models.Hit, 0)
	var boardLinks = make([]*models.Hit, 0)

	var page uint = 1
	var limit uint = 5000 // Upper bound of Grafana API call

	var tagsParams = make([]string, 0)
	if !apphelpers.GetCtxDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
		tagsParams = append(tagsParams, filters.GetTags()...)
	}

	for {
		searchParams := search.NewSearchParams()
		searchParams.Tag = tagsParams
		searchParams.Limit = gapi.ToPtr(int64(limit))
		searchParams.Page = gapi.ToPtr(int64(page))
		searchParams.Type = gapi.ToPtr("dash-db")

		pageBoardLinks, err := s.client.Search.Search(searchParams, s.getAuth())
		if err != nil {
			log.Fatal("Failed to retrieve dashboards", err)
		}
		boardLinks = append(boardLinks, pageBoardLinks.GetPayload()...)
		if len(pageBoardLinks.GetPayload()) < int(limit) {
			break
		}
		page += 1
	}

	folderFilters := filters.GetFolders()
	var validFolder bool
	var validUid bool
	for _, link := range boardLinks {
		validFolder = false
		if apphelpers.GetCtxDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
			validFolder = true
		} else if funk.ContainsString(folderFilters, link.FolderTitle) {
			validFolder = true
		} else if funk.ContainsString(folderFilters, DefaultFolderName) && link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
			validFolder = true
		}
		if !validFolder {
			continue
		}
		link.Slug = updateSlug(link.URI)
		validUid = filters.GetFilter("DashFilter") == "" || link.Slug == filters.GetFilter("DashFilter")
		if link.FolderID == 0 {
			link.FolderTitle = DefaultFolderName
		}

		if validFolder && validUid {
			boardsList = append(boardsList, link)
		}
	}

	return boardsList

}

// ImportDashboards saves all dashboards matching query to configured location
func (s *DashNGoImpl) ImportDashboards(filter Filter) []string {
	var (
		boardLinks []*models.Hit
		rawBoard   []byte
		err        error
		metaData   *dashboards.GetDashboardByUIDOK
	)

	boardLinks = s.ListDashboards(filter)
	var boards []string = make([]string, 0)
	for _, link := range boardLinks {
		dp := dashboards.NewGetDashboardByUIDParams()
		dp.UID = link.UID

		if metaData, err = s.client.Dashboards.GetDashboardByUID(dp, s.getAuth()); err != nil {
			log.Errorf("%s for %s\n", err, link.URI)
			continue
		}

		rawBoard, err = json.Marshal(metaData.Payload.Dashboard)
		if err != nil {
			log.Errorf("unable to serialize dashboard %s", dp.UID)
			continue
		}

		fileName := fmt.Sprintf("%s/%s.json", buildResourceFolder(link.FolderTitle, config.DashboardResource), metaData.Payload.Meta.Slug)
		if err = s.storage.WriteFile(fileName, pretty.Pretty(rawBoard), os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err, metaData.Payload.Meta.Slug)
		} else {
			boards = append(boards, fileName)
		}

	}
	return boards
}

// createFolder Creates a new folder with the given name.
func (s *DashNGoImpl) createdFolder(folderName string) (int64, error) {
	createdFolderRequest := folders.NewCreateFolderParams()
	createdFolderRequest.Body = &models.CreateFolderCommand{
		Title: folderName,
	}
	folder, err := s.client.Folders.CreateFolder(createdFolderRequest, s.getAuth())
	if err != nil {
		return 0, err
	}
	return folder.GetPayload().ID, nil

}

// ExportDashboards finds all the dashboards in the configured location and exports them to grafana.
// if the folder doesn't exist, it'll be created.
func (s *DashNGoImpl) ExportDashboards(filters Filter) {

	var (
		rawBoard   []byte
		folderName string = ""
		folderId   int64
	)
	path := apphelpers.GetCtxDefaultGrafanaConfig().GetPath(config.DashboardResource)
	filesInDir, err := s.storage.FindAllFiles(path, true)
	if err != nil {
		log.WithError(err).Fatal("unable to find any files to export from storage engine")
	}
	//Delete all dashboards that match prior to import
	s.DeleteAllDashboards(filters)

	folderMap := getFolderNameIDMap(s.ListFolder(nil))

	// Fallback on defaults
	if filters == nil {
		filters = &DashboardFilter{}
	}
	validFolders := filters.GetFolders()
	for _, file := range filesInDir {
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")
		validateMap := map[string]string{FolderFilter: folderName, DashFilter: baseFile}
		//If folder OR slug is filtered, then skip if it doesn't match
		if !filters.Validate(validateMap) {
			continue
		}
		if !strings.HasSuffix(file, ".json") {
			log.Warnf("Only json files are supported, skipping %s", file)
			continue
		}

		if rawBoard, err = s.storage.ReadFile(file); err != nil {
			log.Println(err)
			continue
		}
		var board = make(map[string]interface{})
		if err = json.Unmarshal(rawBoard, &board); err != nil {
			log.WithError(err).Printf("Failed to unmarshall file: %s", file)
			continue
		}

		elements := strings.Split(file, string(os.PathSeparator))
		if len(elements) >= 2 {
			folderName = elements[len(elements)-2]
		}
		if folderName == "" || folderName == DefaultFolderName {
			folderId = DefaultFolderId
			folderName = DefaultFolderName
		}
		if !slices.Contains(validFolders, folderName) && !apphelpers.GetCtxDefaultGrafanaConfig().GetFilterOverrides().IgnoreDashboardFilters {
			log.Debugf("Skipping file %s, doesn't match any valid folders", file)
			continue
		}

		if folderName == DefaultFolderName {
			folderId = DefaultFolderId
		} else {
			if val, ok := folderMap[folderName]; ok {
				folderId = val
			} else {
				if filters.Validate(validateMap) {
					id, folderErr := s.createdFolder(folderName)
					if folderErr != nil {
						log.Panic("Unable to create required folder")
					} else {
						folderMap[folderName] = id
						folderId = id
					}
				}
			}
		}

		data := make(map[string]interface{}, 0)

		err = json.Unmarshal(rawBoard, &data)
		//zero out ID.  Can't create a new dashboard if an ID already exists.
		data["id"] = nil
		importDashReq := dashboards.NewImportDashboardParams()
		importDashReq.Body = &models.ImportDashboardRequest{
			FolderID:  folderId,
			Overwrite: true,
			Dashboard: data,
		}

		if _, exportError := s.client.Dashboards.ImportDashboard(importDashReq, s.getAuth()); exportError != nil {
			log.WithError(err).Printf("error on Exporting dashboard %s", file)
			continue
		}

	}
}

// DeleteAllDashboards clears all current dashboards being monitored.  Any folder not white listed
// will not be affected
func (s *DashNGoImpl) DeleteAllDashboards(filter Filter) []string {
	var dashboardListing = make([]string, 0)

	items := s.ListDashboards(filter)
	for _, item := range items {
		if filter.Validate(map[string]string{FolderFilter: item.FolderTitle, DashFilter: item.Slug}) {
			dp := dashboards.NewDeleteDashboardByUIDParams()
			dp.UID = item.UID
			_, err := s.client.Dashboards.DeleteDashboardByUID(dp, s.getAuth())
			if err == nil {
				dashboardListing = append(dashboardListing, item.Title)
			}
		}
	}
	return dashboardListing

}
