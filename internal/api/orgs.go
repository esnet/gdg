package api

import (
	"context"
	"github.com/grafana/grafana-openapi-client-go/models"
	"net/http"
)

func (extended *ExtendedApi) UserOrg() (int64, error) {
	result := new(models.OrgDetailsDTO)
	err := extended.getRequestBuilder().
		Path("api/org").
		ToJSON(result).
		Method(http.MethodGet).Fetch(context.Background())
	if err != nil {
		return 0, err
	}
	return result.ID, err

}
