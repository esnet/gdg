package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/org"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/signed_in_user"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/go-openapi/runtime"
	"github.com/gosimple/slug"
	"path/filepath"
	"strings"

	"github.com/esnet/grafana-swagger-api-golang/goclient/client/orgs"
	log "github.com/sirupsen/logrus"
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
	params := orgs.NewGetOrgByIDParams()
	params.OrgID = id
	data, err := s.client.Orgs.GetOrgByID(params, s.getAuth())
	if err != nil {
		return nil, err
	}

	return data.GetPayload(), nil

}

// SetOrganization sets organization for a given id.
func (s *DashNGoImpl) SetOrganization(id int64) error {
	//Removes Org filter
	if id <= 1 {
		s.grafanaConf.OrganizationId = 1
	} else {
		if s.grafanaConf.IsAdminEnabled() || s.grafanaConf.IsBasicAuth() {
			organization, err := s.getOrganization(id)
			if err != nil {
				return errors.New("invalid org Id, org is not found")
			}
			s.grafanaConf.OrganizationId = organization.ID
		} else {
			s.grafanaConf.OrganizationId = id
		}
	}

	return config.Config().SaveToDisk(false)
}

// ListOrganizations List all dashboards
func (s *DashNGoImpl) ListOrganizations() []*models.OrgDTO {
	if !s.grafanaConf.IsAdminEnabled() {
		log.Errorf("No valid Grafana Admin configured, cannot retrieve Organizations List")
		return nil
	}

	orgList, err := s.client.Orgs.SearchOrgs(orgs.NewSearchOrgsParams(), s.getGrafanaAdminAuth())
	if err != nil {
		var swaggerErr *orgs.SearchOrgsForbidden
		msg := "Cannot retrieve Orgs, you need additional permissions"
		switch {
		case errors.As(err, &swaggerErr):
			var castError *orgs.SearchOrgsForbidden
			errors.As(err, &castError)
			log.WithField("message", *castError.GetPayload().Message).Fatal(msg)
		default:
			log.WithError(err).Fatal(msg)
		}
	}
	return orgList.GetPayload()
}

// DownloadOrganizations Download organizations
func (s *DashNGoImpl) DownloadOrganizations() []string {
	if !s.grafanaConf.IsAdminEnabled() {
		log.Errorf("No valid Grafana Admin configured, cannot retrieve Organizations")
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
			log.Errorf("%s for %s\n", err, organisation.Name)
			continue
		}
		dsPath := buildResourcePath(slug.Make(organisation.Name), config.OrganizationResource)
		if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
			log.Errorf("%s for %s\n", err.Error(), slug.Make(organisation.Name))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles
}

// UploadOrganizations Upload organizations to Grafana
func (s *DashNGoImpl) UploadOrganizations() []string {
	if !s.grafanaConf.IsAdminEnabled() {
		log.Errorf("No valid Grafana Admin configured, cannot upload Organizations")
		return nil
	}
	var (
		result    []string
		rawFolder []byte
	)
	if s.grafanaConf.IsAdminEnabled() {

	}
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.OrganizationResource), false)
	if err != nil {
		log.WithError(err).Fatal("Failed to read folders imports")
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
				log.WithError(err).Errorf("failed to read file %s", fileLocation)
				continue
			}
		}
		var newOrg models.CreateOrgCommand
		if err = json.Unmarshal(rawFolder, &newOrg); err != nil {
			log.WithError(err).Warn("failed to unmarshall folder")
			continue
		}
		if _, ok := orgMap[newOrg.Name]; ok {
			log.Infof("Organizaiton %s already exists, skipping", newOrg.Name)
			continue
		}

		params := orgs.NewCreateOrgParams()
		params.Body = &newOrg
		_, err = s.client.Orgs.CreateOrg(params, s.getBasicAuth())
		if err != nil {
			log.Errorf("failed to create folder %s", newOrg.Name)
			continue
		}
		result = append(result, newOrg.Name)

	}
	return result
}

// SwitchOrganization switch organization context
func (s *DashNGoImpl) SwitchOrganization(id int64) error {
	if !s.grafanaConf.IsBasicAuth() {
		log.Warnf("Basic auth required for Org switching.  Ignoring Org setting and continuing")
		return nil
	}
	valid := false
	if id > 1 {
		var orgsPayload []*models.OrgDTO
		orgList, err := s.client.Orgs.SearchOrgs(orgs.NewSearchOrgsParams(), s.getBasicAuth())
		if err != nil {
			log.Warn("Error fetch organizations requires (SuperAdmin Basic Auth), assuming valid ID was requested.  Cannot validate OrgId")
			valid = true
			orgsPayload = make([]*models.OrgDTO, 0)
		} else {
			orgsPayload = orgList.GetPayload()
		}
		for _, org := range orgsPayload {
			log.Debugf("%d %s\n", org.ID, org.Name)
			if org.ID == s.grafanaConf.GetOrganizationId() {
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

	params := signed_in_user.NewUserSetUsingOrgParams()
	params.OrgID = id
	status, err := s.client.SignedInUser.UserSetUsingOrg(params, s.getBasicAuth())
	if err != nil {
		log.WithError(err).Fatalf("%s for %v\n", err, status)
		return err
	}

	return nil
}

// GetUserOrganization returns the organizations the user is a member of.
func (s *DashNGoImpl) GetUserOrganization() *models.OrgDetailsDTO {
	return s.getAssociatedActiveOrg(s.getBasicAuth())
}

// GetTokenOrganization returns the organizations associated with the given token. (This property is immutable)
func (s *DashNGoImpl) GetTokenOrganization() *models.OrgDetailsDTO {
	return s.getAssociatedActiveOrg(s.getAuth())
}

// getAssociatedActiveOrg returns the Org associated with the given authentication mechanism.
func (s *DashNGoImpl) getAssociatedActiveOrg(auth runtime.ClientAuthInfoWriter) *models.OrgDetailsDTO {
	p := org.NewGetCurrentOrgParams()
	payload, err := s.client.Org.GetCurrentOrg(p, auth)
	if err != nil {
		log.WithError(err).Fatal("Unable to retrieve current organization")
	}
	return payload.GetPayload()
}

func (s *DashNGoImpl) SetUserOrganizations(id int64) error {
	p := signed_in_user.NewUserSetUsingOrgParams()
	p.OrgID = id
	payload, err := s.client.SignedInUser.UserSetUsingOrg(p, s.getBasicAuth())
	if err == nil {
		log.Debugf(payload.GetPayload().Message)
	}
	return err
}

func (s *DashNGoImpl) UpdateCurrentOrganization(name string) error {
	p := org.NewUpdateCurrentOrgParams()
	p.Body = &models.UpdateOrgForm{Name: name}
	_, err := s.client.Org.UpdateCurrentOrg(p, s.getAuth())
	return err
}

func (s *DashNGoImpl) ListOrgUsers(orgId int64) []*models.OrgUserDTO {
	p := orgs.NewGetOrgUsersParams()
	p.OrgID = orgId
	resp, err := s.client.Orgs.GetOrgUsers(p, s.getGrafanaAdminAuth())
	if err != nil {
		log.WithError(err).Fatal("failed to get org users")
	}
	return resp.GetPayload()
}

func (s *DashNGoImpl) AddUserToOrg(role string, userId, orgId int64) error {
	userInfo, err := s.getUserById(userId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user with Id: %d", userId)
	}
	p := orgs.NewAddOrgUserParams()
	p.OrgID = orgId
	p.Body = &models.AddOrgUserCommand{
		LoginOrEmail: userInfo.Login,
		Role:         role,
	}
	_, err = s.client.Orgs.AddOrgUser(p, s.getGrafanaAdminAuth())
	return err
}

func (s *DashNGoImpl) DeleteUserFromOrg(userId, orgId int64) error {
	p := orgs.NewRemoveOrgUserParams()
	p.OrgID = orgId
	p.UserID = userId
	_, err := s.client.Orgs.RemoveOrgUser(p, s.getGrafanaAdminAuth())
	return err
}

func (s *DashNGoImpl) UpdateUserInOrg(role string, userId, orgId int64) error {
	p := orgs.NewUpdateOrgUserParams()
	p.OrgID = orgId
	p.UserID = userId
	p.Body = &models.UpdateOrgUserCommand{
		Role: role,
	}
	_, err := s.client.Orgs.UpdateOrgUser(p, s.getGrafanaAdminAuth())
	return err
}
