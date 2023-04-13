package api

import (
	"context"
	"fmt"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/service_accounts"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"net/http"
)

func (extended *ExtendedApi) ListTokens(query *service_accounts.ListTokensParams) ([]*models.TokenDTO, error) {
	tokens := make([]*models.TokenDTO, 0)
	path := fmt.Sprintf("/api/serviceaccounts/%d/tokens", query.ServiceAccountID)
	err := extended.req.
		Path(path).
		ToJSON(&tokens).
		Method(http.MethodGet).Fetch(context.Background())
	return tokens, err

}
