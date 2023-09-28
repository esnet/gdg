package service

import (
	"fmt"

	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/tools"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/service_accounts"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

type ServiceAccountApi interface {
	ListServiceAccounts() []*api.ServiceAccountDTOWithTokens
	ListServiceAccountsTokens(id int64) ([]*models.TokenDTO, error)
	DeleteAllServiceAccounts() []string
	DeleteServiceAccountTokens(serviceId int64) []string
	CreateServiceAccountToken(name int64, role string, expiration int64) (*models.NewAPIKeyResult, error)
	CreateServiceAccount(name, role string, expiration int64) (*models.ServiceAccountDTO, error)
}

func (s *DashNGoImpl) CreateServiceAccount(name, role string, expiration int64) (*models.ServiceAccountDTO, error) {
	p := service_accounts.NewCreateServiceAccountParams()
	p.Body = &models.CreateServiceAccountForm{
		Name: name,
		Role: role,
	}
	data, err := s.client.ServiceAccounts.CreateServiceAccount(p, s.getAuth())
	if err != nil {
		log.WithField("serivceName", name).
			WithField("role", role).
			Fatal("unable to create a service request")
	}

	return data.GetPayload(), nil
}

func (s *DashNGoImpl) CreateServiceAccountToken(serviceAccountId int64, name string, expiration int64) (*models.NewAPIKeyResult, error) {
	p := service_accounts.NewCreateTokenParams()
	p.Body = &models.AddServiceAccountTokenCommand{
		Name:          name,
		SecondsToLive: expiration,
	}
	p.ServiceAccountID = serviceAccountId
	token, err := s.client.ServiceAccounts.CreateToken(p, s.getAuth())
	if err != nil {
		log.Error(err.Error())
		log.Fatalf("unable to create token '%s' for service account ID: %d", name, serviceAccountId)

	}

	return token.GetPayload(), nil
}

func (s *DashNGoImpl) ListServiceAccounts() []*api.ServiceAccountDTOWithTokens {
	p := service_accounts.NewSearchOrgServiceAccountsWithPagingParams()
	p.Disabled = tools.PtrOf(false)
	p.Perpage = tools.PtrOf(int64(5000))

	resp, err := s.client.ServiceAccounts.SearchOrgServiceAccountsWithPaging(p, s.getAuth())
	if err != nil {
		log.Fatal("unable to retrieve service accounts")
	}
	data := resp.GetPayload()
	result := lo.Map(data.ServiceAccounts, func(entity *models.ServiceAccountDTO, _ int) *api.ServiceAccountDTOWithTokens {
		t := api.ServiceAccountDTOWithTokens{
			ServiceAccount: entity,
		}
		return &t
	})
	for _, item := range result {
		if item.ServiceAccount.Tokens > 0 {
			item.Tokens, err = s.ListServiceAccountsTokens(item.ServiceAccount.ID)
			if err != nil {
				log.Warnf("failed to retrieve tokens for service account %d", item.ServiceAccount.ID)
			}
		}

	}

	return result
}

func (s *DashNGoImpl) ListServiceAccountsTokens(id int64) ([]*models.TokenDTO, error) {

	p := service_accounts.NewListTokensParams()
	p.ServiceAccountID = id
	response, err := s.extended.ListTokens(p)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve service account for %d response", id)
	}

	return response, nil
}

func (s *DashNGoImpl) DeleteAllServiceAccounts() []string {
	var accountNames []string
	accounts := s.ListServiceAccounts()
	for _, account := range accounts {
		p := service_accounts.NewDeleteServiceAccountParams()
		p.ServiceAccountID = account.ServiceAccount.ID
		_, err := s.client.ServiceAccounts.DeleteServiceAccount(p, s.getAuth())
		if err != nil {
			log.Warnf("Failed to delete service account %d", p.ServiceAccountID)
		} else {
			accountNames = append(accountNames, fmt.Sprintf("service account %d has been deleted", p.ServiceAccountID))
		}
	}

	return accountNames
}

func (s *DashNGoImpl) DeleteServiceAccountTokens(serviceId int64) []string {
	var result []string
	tokens, err := s.ListServiceAccountsTokens(serviceId)
	if err != nil {
		log.Fatalf("failed to retrieve tokens for the given service ID: %d", serviceId)
	}

	for _, token := range tokens {
		p := service_accounts.NewDeleteTokenParams()
		p.TokenID = token.ID
		p.ServiceAccountID = serviceId
		_, err := s.client.ServiceAccounts.DeleteToken(p, s.getAuth())
		if err != nil {
			log.Errorf("unable to delete token ID: %d", token.ID)
			continue
		}
		result = append(result, token.Name)
	}

	return result
}
