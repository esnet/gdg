package types

import "github.com/grafana/grafana-openapi-client-go/models"

type ServiceAccountDTOWithTokens struct {
	ServiceAccount *models.ServiceAccountDTO
	Tokens         []*models.TokenDTO
}
