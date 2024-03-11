package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/types"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/orgs"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/gjson"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
)

type organizationCrudApi interface {
	ListOrganizations(filter filters.Filter) []*types.OrgsDTOWithPreferences
	DownloadOrganizations(filter filters.Filter) []string
	UploadOrganizations(filter filters.Filter) []string
}

type organizationToolsApi interface {
	//Manage Active Organization
	SetOrganizationByName(name string, useSlug bool) error
	GetUserOrganization() *models.OrgDetailsDTO
	GetTokenOrganization() *models.OrgDetailsDTO
	SetUserOrganizations(id int64) error
	ListUserOrganizations() ([]*models.UserOrgDTO, error)
}

// organizationUserCrudApi  Manages user memberships to an org
type organizationUserCrudApi interface {
	ListOrgUsers(orgId int64) []*models.OrgUserDTO
	AddUserToOrg(role, orgSlug string, userId int64) error
	DeleteUserFromOrg(orgId string, userId int64) error
	UpdateUserInOrg(role, orgSlug string, userId int64) error
}

// OrganizationsApi Contract definition
type OrganizationsApi interface {
	organizationCrudApi
	organizationToolsApi
	organizationUserCrudApi
	OrgPreferencesApi
	InitOrganizations()
}

func NewOrganizationFilter(args ...string) filters.Filter {
	filterObj := filters.NewBaseFilter()
	if len(args) == 0 || args[0] == "" {
		return filterObj
	}

	filterObj.AddFilter(filters.OrgFilter, args[0])
	return filterObj
}

// InitOrganizations will context switch to configured organization and invoke a different call depending on the access level.
func (s *DashNGoImpl) InitOrganizations() {
	var orgInfo *models.OrgDetailsDTO
	var orgEntity models.OrgDetailsDTO

	if s.grafanaConf.IsAdminEnabled() || s.grafanaConf.IsBasicAuth() {
		orgInfo = s.GetUserOrganization()
		if orgInfo == nil {
			log.Fatal("Unable to retrieve requested user's org")
		}
		if orgInfo.Name != s.grafanaConf.GetOrganizationName() {
			userOrgs, err := s.ListUserOrganizations()
			if err != nil {
				log.Fatal("Unable to switch user's Org")
			}
			found := false
			for _, org := range userOrgs {
				if org.Name == s.grafanaConf.GetOrganizationName() {
					orgEntity.ID = org.OrgID
					orgEntity.Name = org.Name
					found = true
					break
				}
			}
			if !found {
				log.Fatalf("User does not have access to org: '%s', Unable to switch user's Org", s.grafanaConf.GetOrganizationName())
			}

		}
		if orgInfo.Name != s.grafanaConf.GetOrganizationName() {
			err := s.SetUserOrganizations(orgEntity.ID)
			if err != nil {
				log.Fatal("Unable to switch user's Org")
			}
		}

	} else {
		orgInfo = &models.OrgDetailsDTO{
			Name: s.grafanaConf.GetOrganizationName(),
		}

	}
}

func (s *DashNGoImpl) SetOrganizationByName(name string, useSlug bool) error {

	if s.grafanaConf.IsAdminEnabled() || s.grafanaConf.IsBasicAuth() {
		payload, err := s.ListUserOrganizations()
		if err != nil {
			return err
		}
		var requestOrg *models.UserOrgDTO

		for ndx, orgEntity := range payload {
			orgName := orgEntity.Name
			if useSlug {
				orgName = slug.Make(orgName)
			}
			if orgName == name {
				requestOrg = payload[ndx]
				break
			}
		}
		if requestOrg == nil {
			log.Fatalf("unable to set org.  Please ensure you have the correct permissions and the org name is correct")
		}
		s.grafanaConf.OrganizationName = requestOrg.Name
	} else {
		tokenOrg := s.GetTokenOrganization()
		orgName := tokenOrg.Name
		if useSlug {
			orgName = slug.Make(orgName)
		}
		if orgName != name {
			log.Fatalf("you have no BasicAuth configured, and token org are non-changeable.  Please configure a different token associated with Org %s, OR configure basic auth.", orgName)
		}
	}

	return config.Config().SaveToDisk(false)

}

// ListOrganizations List all dashboards
func (s *DashNGoImpl) ListOrganizations(filter filters.Filter) []*types.OrgsDTOWithPreferences {
	if !s.grafanaConf.IsAdminEnabled() {
		slog.Error("No valid Grafana Admin configured, cannot retrieve Organizations List")
		return nil
	}

	orgList, err := s.GetAdminClient().Orgs.SearchOrgs(orgs.NewSearchOrgsParams())
	if err != nil {
		var swaggerErr *orgs.SearchOrgsForbidden
		msg := "Cannot retrieve Orgs, you need additional permissions"
		switch {
		case errors.As(err, &swaggerErr):
			var castError *orgs.SearchOrgsForbidden
			errors.As(err, &castError)
			log.Fatalf("%s, message:%s", msg, *castError.GetPayload().Message)
		default:
			log.Fatalf("%s, err: %v", msg, err)
		}
	}

	var resultsData []*types.OrgsDTOWithPreferences
	for _, org := range orgList.GetPayload() {
		if filter.GetFilter(filters.OrgFilter) == "" || filter.GetFilter(filters.OrgFilter) == org.Name {
			preferences, err := s.GetOrgPreferences(org.Name)
			if err != nil {
				slog.Warn("unable to retrieve org preferences for org", slog.String("organization", org.Name))
				preferences = &models.Preferences{}
			}
			resultsData = append(resultsData, &types.OrgsDTOWithPreferences{Organization: org, Preferences: preferences})
		}

	}

	return resultsData
}

// DownloadOrganizations Download organizations
func (s *DashNGoImpl) DownloadOrganizations(filter filters.Filter) []string {
	if !s.grafanaConf.IsAdminEnabled() {
		slog.Error("No valid Grafana Admin configured, cannot retrieve Organizations")
		return nil
	}
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)

	orgsListing := s.ListOrganizations(filter)
	for _, organisation := range orgsListing {
		if dsPacked, err = json.MarshalIndent(organisation, "", "	"); err != nil {
			slog.Error("Unable to serialize organization object", "err", err, "organization", organisation.Organization.Name)
			continue
		}
		dsPath := buildResourcePath(slug.Make(organisation.Organization.Name), config.OrganizationResource)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "err", err.Error(), "organization", slug.Make(organisation.Organization.Name))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles
}

// UploadOrganizations Upload organizations to Grafana
func (s *DashNGoImpl) UploadOrganizations(filter filters.Filter) []string {
	if !s.grafanaConf.IsAdminEnabled() {
		slog.Error("No valid Grafana Admin configured, cannot upload Organizations")
		return nil
	}
	var (
		result    []string
		rawFolder []byte
	)
	//syncedMap := new(sync.Map)
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.OrganizationResource), false)
	if err != nil {
		log.Fatalf("Failed to read folders imports, err: %v", err)
	}
	orgListing := s.ListOrganizations(filter)
	orgMap := map[string]bool{}
	for _, entry := range orgListing {
		orgMap[entry.Organization.Name] = true
	}

	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.OrganizationResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
		}
		var jsonOrg types.OrgsDTOWithPreferences
		var newOrg models.CreateOrgCommand
		if err = json.Unmarshal(rawFolder, &jsonOrg); err != nil {
			slog.Warn("failed to unmarshall folder", "err", err)
			continue
		}
		if jsonOrg.Organization == nil {
			slog.Warn("unable to retrieve Org info from file", slog.String("file", file))
			continue
		}
		newOrg.Name = jsonOrg.Organization.Name
		rawOrgName := gjson.GetBytes(rawFolder, "name").String()
		if filter.GetFilter(filters.OrgFilter) != "" && rawOrgName != filter.GetFilter(filters.OrgFilter) {
			continue
		}
		updateProperties := func(org *types.OrgsDTOWithPreferences) error {
			if org.Preferences == nil || org.Organization == nil {
				slog.Warn("Properties or Organization is nil, ignore update request")
				return nil
			}
			return s.UploadOrgPreferences(org.Organization.Name, org.Preferences)
		}
		if _, ok := orgMap[newOrg.Name]; ok {
			slog.Info("Organization already exists, skipping", "organization", newOrg.Name)
			err = updateProperties(&jsonOrg)
			if err != nil {
				slog.Warn("unable to update Org properties for org.", slog.String("organization", newOrg.Name))
			}
			continue
		}

		_, err = s.GetBasicAuthClient().Orgs.CreateOrg(&newOrg)
		if err != nil {
			slog.Error("failed to create organization", "organization", newOrg.Name)
			continue
		}
		result = append(result, newOrg.Name)
		err = updateProperties(&jsonOrg)
		if err != nil {
			slog.Warn("unable to update Org properties for org.", slog.String("organization", newOrg.Name))
		}

	}
	return result
}

// SwitchOrganizationByName switch organization context
func (s *DashNGoImpl) SwitchOrganizationByName(orgName string) error {
	if !s.grafanaConf.IsBasicAuth() {
		slog.Warn("Basic auth required for Org switching.  Ignoring Org setting and continuing")
		return nil
	}
	valid := false
	var orgId int64 = 1
	if orgName != "" {
		var orgsPayload []*models.OrgDTO
		orgList, err := s.GetBasicAuthClient().Orgs.SearchOrgs(orgs.NewSearchOrgsParams())
		if err != nil {
			slog.Warn("Error fetch organizations requires (SuperAdmin Basic SecureData), assuming valid ID was requested.  Cannot validate OrgId")
			valid = true
			orgsPayload = make([]*models.OrgDTO, 0)
		} else {
			orgsPayload = orgList.GetPayload()
		}
		for _, orgEntry := range orgsPayload {
			slog.Debug("", "orgID", orgEntry.ID, "OrgName", orgEntry.Name)
			if orgEntry.Name == s.grafanaConf.GetOrganizationName() {
				valid = true
				orgId = orgEntry.ID
				break
			}
		}

	} else {
		//Fallback on default
		valid = true
		orgId = config.DefaultOrganizationId
	}

	//We retrieved all the orgs successfully and none of them matched the requested ID
	if !valid {
		log.Fatalf("The Specified OrgId does not match any existing organization.  Please check your configuration and try again.")
	}

	status, err := s.GetBasicAuthClient().SignedInUser.UserSetUsingOrg(orgId)
	if err != nil {
		log.Fatalf("%s for %v\n", err, status)
		return err
	}

	return nil
}

// GetUserOrganization returns the organizations the user is a member of.
func (s *DashNGoImpl) GetUserOrganization() *models.OrgDetailsDTO {
	return s.getAssociatedActiveOrg(s.GetBasicAuthClient())
}

// GetTokenOrganization returns the organizations associated with the given token. (This property is immutable)
func (s *DashNGoImpl) GetTokenOrganization() *models.OrgDetailsDTO {
	return s.getAssociatedActiveOrg(s.GetClient())
}

// getAssociatedActiveOrg returns the Org associated with the given authentication mechanism.
func (s *DashNGoImpl) getAssociatedActiveOrg(apiClient *client.GrafanaHTTPAPI) *models.OrgDetailsDTO {
	payload, err := apiClient.Org.GetCurrentOrg()
	if err != nil {
		log.Fatalf("Unable to retrieve current organization, err: %v", err)
	}
	return payload.GetPayload()
}

func (s *DashNGoImpl) ListUserOrganizations() ([]*models.UserOrgDTO, error) {
	payload, err := s.GetBasicAuthClient().SignedInUser.GetSignedInUserOrgList()
	if err != nil {
		return nil, err
	}

	return payload.GetPayload(), nil

}

func (s *DashNGoImpl) SetUserOrganizations(id int64) error {
	payload, err := s.GetBasicAuthClient().SignedInUser.UserSetUsingOrg(id)
	if err == nil {
		slog.Debug(payload.GetPayload().Message)
	}
	return err
}

func (s *DashNGoImpl) UpdateCurrentOrganization(name string) error {
	p := &models.UpdateOrgForm{Name: name}
	_, err := s.GetClient().Org.UpdateCurrentOrg(p)
	return err
}

func (s *DashNGoImpl) ListOrgUsers(orgId int64) []*models.OrgUserDTO {
	p := orgs.NewGetOrgUsersParams()
	p.OrgID = orgId
	resp, err := s.GetAdminClient().Orgs.GetOrgUsers(orgId)
	if err != nil {
		log.Fatalf("failed to get org users, err: %v", err)
	}
	return resp.GetPayload()
}

func (s *DashNGoImpl) AddUserToOrg(role, orgSlug string, userId int64) error {
	userInfo, err := s.getUserById(userId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user with Id: %d", userId)
	}
	request := &models.AddOrgUserCommand{
		LoginOrEmail: userInfo.Login,
		Role:         role,
	}

	orgEntity, err := s.getOrgIdFromSlug(orgSlug)
	if err != nil {
		return fmt.Errorf("unable to find a valid org with slug value of %s", orgSlug)
	}

	_, err = s.GetAdminClient().Orgs.AddOrgUser(orgEntity.OrgID, request)
	return err
}

func (s *DashNGoImpl) DeleteUserFromOrg(orgSlugName string, userId int64) error {
	orgEntity, err := s.getOrgIdFromSlug(orgSlugName)
	if err != nil {
		return err
	}
	_, err = s.GetAdminClient().Orgs.RemoveOrgUser(userId, orgEntity.OrgID)
	return err
}

func (s *DashNGoImpl) getOrgIdFromSlug(slugName string) (*models.UserOrgDTO, error) {
	//Get Org
	organizations, err := s.ListUserOrganizations()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve user organizations, %w", err)
	}
	var orgId int64
	var orgEntity *models.UserOrgDTO
	for _, org := range organizations {
		if slug.Make(org.Name) == slugName {
			orgId = org.OrgID
			orgEntity = org
			break
		}
	}
	if orgId == 0 {
		return nil, fmt.Errorf("unable to find org with matching slug name of %s", slugName)
	}
	return orgEntity, nil

}

func (s *DashNGoImpl) UpdateUserInOrg(role, orgSlug string, userId int64) error {
	p := orgs.NewUpdateOrgUserParams()
	orgEntity, err := s.getOrgIdFromSlug(orgSlug)
	if err != nil {
		return err
	}
	p.OrgID = orgEntity.OrgID
	p.UserID = userId
	p.Body = &models.UpdateOrgUserCommand{
		Role: role,
	}

	_, err = s.GetAdminClient().Orgs.UpdateOrgUser(p)
	return err
}
