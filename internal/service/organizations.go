package service

import (
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/signed_in_user"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"

	"github.com/esnet/grafana-swagger-api-golang/goclient/client/orgs"
	log "github.com/sirupsen/logrus"
)

// OrganizationsApi Contract definition
type OrganizationsApi interface {
	ListOrganizations() []*models.OrgDTO
}

// ListOrganizations: List all dashboards
func (s *DashNGoImpl) ListOrganizations() []*models.OrgDTO {

	orgList, err := s.client.Orgs.SearchOrgs(orgs.NewSearchOrgsParams(), s.getAdminAuth())
	if err != nil {
		log.WithError(err).Errorf("Unable to retrieve Organization List")
	}
	if s.grafanaConf.Organization != "" {
		var ID int64
		for _, org := range orgList.Payload {
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

	orgList, err = s.client.Orgs.SearchOrgs(orgs.NewSearchOrgsParams(), s.getAuth())
	if err != nil {
		panic(err)
	}
	return orgList.GetPayload()
}
