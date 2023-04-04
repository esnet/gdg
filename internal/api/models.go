package api

import "github.com/esnet/grafana-swagger-api-golang/goclient/models"

type ServiceAccountDTOWithTokens struct {
	ServiceAccount *models.ServiceAccountDTO
	Tokens         []*models.TokenDTO
}
