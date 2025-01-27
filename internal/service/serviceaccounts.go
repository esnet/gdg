package service

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/esnet/gdg/internal/types"

	"github.com/grafana/grafana-openapi-client-go/client/service_accounts"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
)

// TODO: create a method to simply delete a service account.
func (s *DashNGoImpl) CreateServiceAccount(name, role string, expiration int64) (*models.ServiceAccountDTO, error) {
	p := service_accounts.NewCreateServiceAccountParams()
	p.Body = &models.CreateServiceAccountForm{
		Name: name,
		Role: role,
	}
	data, err := s.GetClient().ServiceAccounts.CreateServiceAccount(p)
	if err != nil {
		log.Fatalf("unable to create a service request, serviceName: %s, role: %s", name, role)
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
	token, err := s.GetClient().ServiceAccounts.CreateToken(p)
	if err != nil {
		log.Fatalf("unable to create token '%s' for service account ID: %d, err: %v", name, serviceAccountId, err)
	}

	return token.GetPayload(), nil
}

func (s *DashNGoImpl) ListServiceAccounts() []*types.ServiceAccountDTOWithTokens {
	p := service_accounts.NewSearchOrgServiceAccountsWithPagingParams()
	p.Disabled = ptr.Of(false)
	p.Perpage = ptr.Of(int64(5000))

	resp, err := s.GetClient().ServiceAccounts.SearchOrgServiceAccountsWithPaging(p)
	if err != nil {
		log.Fatal("unable to retrieve service accounts")
	}
	data := resp.GetPayload()
	result := lo.Map(data.ServiceAccounts, func(entity *models.ServiceAccountDTO, _ int) *types.ServiceAccountDTOWithTokens {
		t := types.ServiceAccountDTOWithTokens{
			ServiceAccount: entity,
		}
		return &t
	})
	for _, item := range result {
		if item.ServiceAccount.Tokens > 0 {
			item.Tokens, err = s.ListServiceAccountsTokens(item.ServiceAccount.ID)
			if err != nil {
				slog.Warn("failed to retrieve tokens for service account", "serviceAccountId", item.ServiceAccount.ID)
			}
		}
	}

	return result
}

func (s *DashNGoImpl) ListServiceAccountsTokens(id int64) ([]*models.TokenDTO, error) {
	response, err := s.GetClient().ServiceAccounts.ListTokens(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve service account for %d response", id)
	}

	return response.GetPayload(), nil
}

func (s *DashNGoImpl) DeleteServiceAccount(accountId int64) error {
	_, err := s.GetClient().ServiceAccounts.DeleteServiceAccount(accountId)
	return err
}

func (s *DashNGoImpl) DeleteAllServiceAccounts() []string {
	var accountNames []string
	accounts := s.ListServiceAccounts()
	for _, account := range accounts {
		accountId := account.ServiceAccount.ID
		err := s.DeleteServiceAccount(accountId)
		if err != nil {
			slog.Warn("Failed to delete service account", "ServiceAccountId", accountId)
		} else {
			accountNames = append(accountNames, fmt.Sprintf("service account %d has been deleted", accountId))
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
		_, err := s.GetClient().ServiceAccounts.DeleteToken(token.ID, serviceId)
		if err != nil {
			slog.Error("unable to delete token", "tokenID", token.ID)
			continue
		}
		result = append(result, token.Name)
	}

	return result
}
