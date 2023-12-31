package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/orgs"
	"github.com/grafana/grafana-openapi-client-go/models"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
)

// OrganizationsApi Contract definition
type OrganizationsApi interface {
	ListOrganizations() []*models.OrgDTO
	DownloadOrganizations() []string
	UploadOrganizations() []string
	SetOrganization(id int64) error
	//Manage Active Organization
	GetUserOrganization() *models.OrgDetailsDTO
	GetTokenOrganization() *models.OrgDetailsDTO
	SetUserOrganizations(id int64) error
	InitOrganizations()
	//Org Users
	ListOrgUsers(orgId int64) []*models.OrgUserDTO
	AddUserToOrg(role string, userId, orgId int64) error
	DeleteUserFromOrg(userId, orgId int64) error
	UpdateUserInOrg(role string, userId, orgId int64) error
}

// InitOrganizations will context switch to configured organization and invoke a different call depending on the access level.
func (s *DashNGoImpl) InitOrganizations() {
	var orgInfo *models.OrgDetailsDTO

	if s.grafanaConf.IsAdminEnabled() || s.grafanaConf.IsBasicAuth() {
		orgInfo = s.GetUserOrganization()
		if orgInfo == nil {
			log.Fatal("Unable to retrieve requested user's org")
		}
		if orgInfo.ID != s.grafanaConf.GetOrganizationId() {
			err := s.SetUserOrganizations(s.grafanaConf.GetOrganizationId())
			if err != nil {
				log.Fatal("Unable to switch user's Org")
			}
		}

	} else {
		orgInfo = &models.OrgDetailsDTO{
			ID: s.grafanaConf.GetOrganizationId(),
		}

	}
}

// getOrganizations returns organization for a given id.
func (s *DashNGoImpl) getOrganization(id int64) (*models.OrgDetailsDTO, error) {
	data, err := s.GetClient().Orgs.GetOrgByID(id)
	if err != nil {
		return nil, err
	}

	return data.GetPayload(), nil

}

// SetOrganization sets organization for a given id.
func (s *DashNGoImpl) SetOrganization(id int64) error {
	//Removes Org filter
	if id <= 1 {
		slog.Warn("organization is not a valid value, resetting to default value of 1.")
		s.grafanaConf.OrganizationId = 1
	}

	if s.grafanaConf.IsAdminEnabled() || s.grafanaConf.IsBasicAuth() {
		organization, err := s.getOrganization(id)
		if err != nil {
			return errors.New("invalid org Id, org is not found")
		}
		s.grafanaConf.OrganizationId = organization.ID
	} else {
		tokenOrg := s.GetTokenOrganization()
		if tokenOrg.ID != id {
			log.Fatalf("you have no BasicAuth configured, and token org are non-changeable.  Please configure a different token associated with Org %d, OR configure basic auth.", id)
		}
		s.grafanaConf.OrganizationId = id
	}

	return config.Config().SaveToDisk(false)
}

// ListOrganizations List all dashboards
func (s *DashNGoImpl) ListOrganizations() []*models.OrgDTO {
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
	return orgList.GetPayload()
}

// DownloadOrganizations Download organizations
func (s *DashNGoImpl) DownloadOrganizations() []string {
	if !s.grafanaConf.IsAdminEnabled() {
		slog.Error("No valid Grafana Admin configured, cannot retrieve Organizations")
		return nil
	}
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)

	orgsListing := s.ListOrganizations()
	for _, organisation := range orgsListing {
		if dsPacked, err = json.MarshalIndent(organisation, "", "	"); err != nil {
			slog.Error("Unable to serialize organization object", "err", err, "organization", organisation.Name)
			continue
		}
		dsPath := buildResourcePath(slug.Make(organisation.Name), config.OrganizationResource)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			slog.Error("Unable to write file", "err", err.Error(), "organization", slug.Make(organisation.Name))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles
}

// UploadOrganizations Upload organizations to Grafana
func (s *DashNGoImpl) UploadOrganizations() []string {
	if !s.grafanaConf.IsAdminEnabled() {
		slog.Error("No valid Grafana Admin configured, cannot upload Organizations")
		return nil
	}
	var (
		result    []string
		rawFolder []byte
	)
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.OrganizationResource), false)
	if err != nil {
		log.Fatalf("Failed to read folders imports, err: %v", err)
	}
	orgListing := s.ListOrganizations()
	orgMap := map[string]bool{}
	for _, entry := range orgListing {
		orgMap[entry.Name] = true
	}

	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.OrganizationResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				slog.Error("failed to read file", "filename", fileLocation, "err", err)
				continue
			}
		}
		var newOrg models.CreateOrgCommand
		if err = json.Unmarshal(rawFolder, &newOrg); err != nil {
			slog.Warn("failed to unmarshall folder", "err", err)
			continue
		}
		if _, ok := orgMap[newOrg.Name]; ok {
			slog.Info("Organization already exists, skipping", "organization", newOrg.Name)
			continue
		}

		_, err = s.GetBasicAuthClient().Orgs.CreateOrg(&newOrg)
		if err != nil {
			slog.Error("failed to create folder", "organization", newOrg.Name)
			continue
		}
		result = append(result, newOrg.Name)

	}
	return result
}

// SwitchOrganization switch organization context
func (s *DashNGoImpl) SwitchOrganization(id int64) error {
	if !s.grafanaConf.IsBasicAuth() {
		slog.Warn("Basic auth required for Org switching.  Ignoring Org setting and continuing")
		return nil
	}
	valid := false
	if id > 1 {
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
			if orgEntry.ID == s.grafanaConf.GetOrganizationId() {
				valid = true
				break
			}
		}

	}
	//Fallback on default
	if id < 2 {
		id = 1 // DefaultOrgID
		valid = true
	}

	//We retrieved all the orgs successfully and none of them matched the requested ID
	if !valid {
		log.Fatalf("The Specified OrgId does not match any existing organization.  Please check your configuration and try again.")
	}

	status, err := s.GetBasicAuthClient().SignedInUser.UserSetUsingOrg(id)
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

func (s *DashNGoImpl) AddUserToOrg(role string, userId, orgId int64) error {
	userInfo, err := s.getUserById(userId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user with Id: %d", userId)
	}
	request := &models.AddOrgUserCommand{
		LoginOrEmail: userInfo.Login,
		Role:         role,
	}
	_, err = s.GetAdminClient().Orgs.AddOrgUser(orgId, request)
	return err
}

func (s *DashNGoImpl) DeleteUserFromOrg(userId, orgId int64) error {
	p := orgs.NewRemoveOrgUserParams()
	p.OrgID = orgId
	p.UserID = userId
	_, err := s.GetAdminClient().Orgs.RemoveOrgUser(userId, orgId)
	return err
}

func (s *DashNGoImpl) UpdateUserInOrg(role string, userId, orgId int64) error {
	p := orgs.NewUpdateOrgUserParams()
	p.OrgID = orgId
	p.UserID = userId
	p.Body = &models.UpdateOrgUserCommand{
		Role: role,
	}
	_, err := s.GetAdminClient().Orgs.UpdateOrgUser(p)
	return err
}
