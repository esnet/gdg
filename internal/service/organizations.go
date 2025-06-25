package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	configDomain "github.com/esnet/gdg/internal/config/domain"

	"github.com/esnet/gdg/internal/service/domain"

	"github.com/esnet/gdg/internal/service/filters/v2"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/tools"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/orgs"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/gjson"
)

func setupOrgReaders(filterObj filters.V2Filter) {
	obj := models.OrgDTO{}
	err := filterObj.RegisterReader(reflect.TypeOf(obj), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.(models.OrgDTO)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.OrgFilter:
			return val.Name, nil

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Org Filter, aborting.")
	}
	err = filterObj.RegisterReader(reflect.TypeOf([]byte{}), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.OrgFilter:
			{
				r := gjson.GetBytes(val, "organization.name")
				if !r.Exists() || r.IsArray() {
					return nil, fmt.Errorf("no valid connection name found")
				}
				return r.String(), nil

			}

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Org Filter, aborting.")
	}
}

func NewOrganizationFilter(args ...string) filters.V2Filter {
	filterObj := v2.NewBaseFilter()
	setupOrgReaders(filterObj)
	if len(args) == 0 || args[0] == "" {
		return filterObj
	}

	filterObj.AddValidation(filters.OrgFilter, func(value any, expected any) error {
		val, expectedValue, convErr := v2.GetParams[string](value, expected, filters.OrgFilter)
		if convErr != nil {
			return convErr
		}
		if expectedValue == "" {
			return nil
		}
		if val != expectedValue {
			return fmt.Errorf("failed Org Name filter, expected %v, got %v", expectedValue, val)
		}
		return nil
	}, args[0])

	return filterObj
}

func (s *DashNGoImpl) sanitizeOrganizationMembership() {
	allOrgs := s.ListOrganizations(NewOrganizationFilter(), false)
	userOrgs, err := s.ListUserOrganizations()
	if err != nil {
		slog.Warn("unable to list user organizations, aborting sanitization operation")
		return
	}
	if len(userOrgs) == len(allOrgs) {
		slog.Debug("Nothing to do, returning")
		return
	}

	userInfo, err := s.GetUserInfo()
	if err != nil {
		slog.Error("Unable to retrieve current user, skipping Org bootstrap")
		return
	}
	userId := userInfo.ID
	currentUserOrgs := map[int64]bool{}
	for _, org := range userOrgs {
		currentUserOrgs[org.OrgID] = true
	}
	// https://github.com/grafana/grafana/issues/79062 Broken state for Orgs, ensuring not in range
	if tools.InRange([]tools.VersionRange{{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"}}, s) {
		slog.Error("version check fails.  Cannot programmatically fix org membership in version v10.2.1-v10.2.2.  " +
			"Please ensure the configured grafana admin is added to the orgs below")
		for _, org := range allOrgs {
			if _, ok := currentUserOrgs[org.Organization.ID]; !ok {
				slog.Info("user is not a member of org", slog.Any("userId", userInfo.ID), slog.String("userName", userInfo.Login), slog.String("Organization", org.Organization.Name))
			}
		}
		os.Exit(1)
	}

	slog.Info("You've configured a grafana admin, but you are not a member of every Org.  The grafanaAdmin needs to be a member of every organization in order for GDG to operate successfully.")
	tools.GetUserConfirmation("Would you like to continue and add the configured grafana admin to all organizations as an 'admin'? (y/n) ", "", true)

	for _, org := range allOrgs {
		if _, ok := currentUserOrgs[org.Organization.ID]; !ok {
			slog.Info("Adding user to organization", slog.String("org", org.Organization.Name))
			err = s.AddUserToOrg("admin", slug.Make(org.Organization.Name), userId)
			if err != nil {
				slog.Error("unable to add user to org", slog.Any("userId", userInfo.ID), slog.String("userName", userInfo.Login), slog.String("Organization", org.Organization.Name))
			} else {
				slog.Info("added user to org", slog.Any("userId", userInfo.ID), slog.String("userName", userInfo.Login), slog.String("Organization", org.Organization.Name))
			}

		}
	}
}

// InitOrganizations will context switch to configured organization and invoke a different call depending on the access level.
func (s *DashNGoImpl) InitOrganizations() {
	var orgInfo *models.OrgDetailsDTO
	var orgEntity models.OrgDetailsDTO

	if s.grafanaConf.IsGrafanaAdmin() || s.grafanaConf.IsBasicAuth() {
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

	} else if s.grafanaConf.GetOrganizationName() != "unknown" {
		slog.Warn("Tokens do no operate across multiple Organizations. Cannot verify or scope into the given org. Please be sure your token belongs to the correct organization", slog.String("configure organization", s.grafanaConf.GetOrganizationName()))
	}

	if s.grafanaConf.IsGrafanaAdmin() {
		slog.Info("Running Sanity Check of Organization Membership")
		s.sanitizeOrganizationMembership()
	}
}

func (s *DashNGoImpl) SetOrganizationByName(name string, useSlug bool) error {
	if s.grafanaConf.IsGrafanaAdmin() || s.grafanaConf.IsBasicAuth() {
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
func (s *DashNGoImpl) ListOrganizations(filter filters.V2Filter, withPreferences bool) []*domain.OrgsDTOWithPreferences {
	if !s.grafanaConf.IsGrafanaAdmin() {
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

	var resultsData []*domain.OrgsDTOWithPreferences
	for _, org := range orgList.GetPayload() {
		if filter.ValidateAll(*org) {
			if !withPreferences {
				resultsData = append(resultsData, &domain.OrgsDTOWithPreferences{Organization: org, Preferences: &models.Preferences{}})
			} else {
				preferences, err := s.GetOrgPreferences(org.Name)
				if err != nil {
					slog.Warn("unable to retrieve org preferences for org", slog.String("organization", org.Name))
					preferences = &models.Preferences{}
				}
				resultsData = append(resultsData, &domain.OrgsDTOWithPreferences{Organization: org, Preferences: preferences})
			}
		}
	}

	return resultsData
}

// DownloadOrganizations Download organizations
func (s *DashNGoImpl) DownloadOrganizations(filter filters.V2Filter) []string {
	if !s.grafanaConf.IsGrafanaAdmin() {
		slog.Error("No valid Grafana Admin configured, cannot retrieve Organizations")
		return nil
	}
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)

	orgsListing := s.ListOrganizations(filter, true)
	for _, organisation := range orgsListing {
		if dsPacked, err = json.MarshalIndent(organisation, "", "	"); err != nil {
			slog.Error("Unable to serialize organization object", "err", err, "organization", organisation.Organization.Name)
			continue
		}
		dsPath := buildResourcePath(slug.Make(organisation.Organization.Name), configDomain.OrganizationResource, s.isLocal(), s.globalConf.ClearOutput)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "err", err.Error(), "organization", slug.Make(organisation.Organization.Name))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles
}

// UploadOrganizations Upload organizations to Grafana
func (s *DashNGoImpl) UploadOrganizations(filter filters.V2Filter) []string {
	if !s.grafanaConf.IsGrafanaAdmin() {
		slog.Error("No valid Grafana Admin configured, cannot upload Organizations")
		return nil
	}
	var (
		result    []string
		rawFolder []byte
	)
	orgName := s.grafanaConf.GetOrganizationName()
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(configDomain.OrganizationResource, orgName), false)
	if err != nil {
		log.Fatalf("Failed to read folders imports, err: %v", err)
	}
	orgListing := s.ListOrganizations(filter, false)
	orgMap := map[string]bool{}
	for _, entry := range orgListing {
		orgMap[entry.Organization.Name] = true
	}

	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(configDomain.OrganizationResource, orgName), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
		}
		var jsonOrg domain.OrgsDTOWithPreferences
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
		if !filter.ValidateAll(rawFolder) {
			slog.Debug("Skipping org, failing filter check", "file", file)
			continue
		}
		updateProperties := func(org *domain.OrgsDTOWithPreferences) error {
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
		// Fallback on default
		valid = true
		orgId = configDomain.DefaultOrganizationId
	}

	// We retrieved all the orgs successfully and none of them matched the requested ID
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

	orgEntity, err := s.getOrgIdFromSlug(orgSlug, true)
	if err != nil {
		return fmt.Errorf("unable to find a valid org with slug value of %s", orgSlug)
	}

	_, err = s.GetAdminClient().Orgs.AddOrgUser(orgEntity.OrgID, request)
	return err
}

func (s *DashNGoImpl) DeleteUserFromOrg(orgSlugName string, userId int64) error {
	orgEntity, err := s.getOrgIdFromSlug(orgSlugName, false)
	if err != nil {
		return err
	}
	_, err = s.GetAdminClient().Orgs.RemoveOrgUser(userId, orgEntity.OrgID)
	return err
}

func (s *DashNGoImpl) getOrgIdFromSlug(slugName string, useAdminListing bool) (*models.UserOrgDTO, error) {
	// Get Org
	var orgId int64
	var orgEntity *models.UserOrgDTO

	if s.grafanaConf.IsGrafanaAdmin() && useAdminListing {
		organizations := s.ListOrganizations(NewOrganizationFilter(), false)
		for _, org := range organizations {
			if org.Organization != nil && slug.Make(org.Organization.Name) == slugName {
				orgId = org.Organization.ID
				orgEntity = &models.UserOrgDTO{
					OrgID: org.Organization.ID,
					Name:  org.Organization.Name,
					Role:  "None",
				}
				break
			}
		}
	} else {
		organizations, err := s.ListUserOrganizations()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve user organizations, %w", err)
		}
		for _, org := range organizations {
			if slug.Make(org.Name) == slugName {
				orgId = org.OrgID
				orgEntity = org
				break
			}
		}
	}

	if orgId == 0 {
		return nil, fmt.Errorf("unable to find org with matching slug name of %s", slugName)
	}
	return orgEntity, nil
}

func (s *DashNGoImpl) UpdateUserInOrg(role, orgSlug string, userId int64) error {
	p := orgs.NewUpdateOrgUserParams()
	orgEntity, err := s.getOrgIdFromSlug(orgSlug, false)
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
