package api

import (
	"context"

	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
)

//ListOrganizations: List all dashboards
func (s *DashNGoImpl) ListOrganizations() []sdk.Org {
	ctx := context.Background()
	orgs, err := s.adminClient.GetAllOrgs(ctx)
	if err != nil {
		log.WithError(err).Errorf("Unable to retrieve Organization List")
	}
	if s.grafanaConf.Organization != "" {
		var ID uint
		for _, org := range orgs {
			log.Errorf("%d %s\n", org.ID, org.Name)
			if org.Name == s.grafanaConf.Organization {
				ID = org.ID
			}
		}
		if ID > 0 {
			status, err := s.client.SwitchActualUserContext(ctx, ID)
			if err != nil {
				log.Fatalf("%s for %v\n", err, status)
			}
		}
	}
	orgs, err = s.client.GetAllOrgs(ctx)
	if err != nil {
		panic(err)
	}
	return orgs
}
