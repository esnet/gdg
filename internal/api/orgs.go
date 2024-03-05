package api

import (
	"context"
	"errors"
	"github.com/grafana/grafana-openapi-client-go/models"
	"net/http"
)

// GetConfiguredOrgId needed to call grafana API in order to configure the Grafana API correctly.  Invoking
// this endpoint manually to avoid a circular dependency.
func (extended *ExtendedApi) GetConfiguredOrgId(orgName string) (int64, error) {
	var result []*models.UserOrgDTO
	err := extended.getRequestBuilder().
		Path("api/user/orgs").
		ToJSON(&result).
		Method(http.MethodGet).
		Fetch(context.Background())
	if err != nil {
		return 0, err
	}
	for _, entity := range result {
		if entity.Name == orgName {
			return entity.OrgID, nil
		}
	}
	return 0, errors.New("org not found")
}
