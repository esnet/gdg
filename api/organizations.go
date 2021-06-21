package api

import (
	"context"
	"fmt"
	"os"

	"github.com/netsage-project/sdk"
)

//ListOrganizations: List all dashboards
func (s *DashNGoImpl) ListOrganizations() []sdk.Org {
	ctx := context.Background()
	orgs, err := s.client.GetAllOrgs(ctx)
	if err != nil {
		panic(err)
	}
	if s.grafanaConf.Organization != "" {
		var ID uint
		for _, org := range orgs {
			fmt.Fprintf(os.Stderr, "%d %s\n", org.ID, org.Name)
			if org.Name == s.grafanaConf.Organization {
				ID = org.ID
			}
		}
		if ID > 0 {
			status, err := s.client.SwitchActualUserContext(ctx, ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s for %v\n", err, status)
				panic(err)
			}
		}
	}
	orgs, err = s.client.GetAllOrgs(ctx)
	if err != nil {
		panic(err)
	}
	return orgs
}
