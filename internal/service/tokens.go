package service

import (
	"fmt"
	"log/slog"

	"github.com/grafana/grafana-openapi-client-go/client/api_keys"
	"github.com/grafana/grafana-openapi-client-go/models"
	"log"
)

// ListAPIKeys returns a list of all known API Keys and service accounts
func (s *DashNGoImpl) ListAPIKeys() []*models.APIKeyDTO {
	params := api_keys.NewGetAPIkeysParams()
	keys, err := s.GetBasicAuthClient().APIKeys.GetAPIkeys(params)
	if err != nil {
		log.Fatal("unable to list API Keys")
	}
	return keys.GetPayload()
}

// DeleteAllTokens Deletes all known tokens
func (s *DashNGoImpl) DeleteAllTokens() []string {
	deleted := []string{}
	keys := s.ListAPIKeys()
	for _, key := range keys {
		err := s.deleteAPIKey(key.ID)
		if err != nil {
			slog.Warn("Failed to delete API key", "APIKeyID", key.ID, "APIKey", key.Name)
			continue
		}
		deleted = append(deleted, key.Name)
	}

	return deleted
}

// CreateAPIKey create a new key for the given role and expiration specified
func (s *DashNGoImpl) CreateAPIKey(name, role string, expiration int64) (*models.NewAPIKeyResult, error) {
	p := &models.AddAPIKeyCommand{
		Name: name,
		Role: role,
	}
	if expiration != 0 {
		p.SecondsToLive = expiration
	}
	newKey, err := s.GetClient().APIKeys.AddAPIkey(p)
	if err != nil {
		return nil, fmt.Errorf("unable to create a new API Key")
	}
	return newKey.GetPayload(), nil

}
func (s *DashNGoImpl) deleteAPIKey(id int64) error {
	_, err := s.GetClient().APIKeys.DeleteAPIkey(id)
	if err != nil {
		return fmt.Errorf("failed to delete API Key: %d", id)
	}
	return nil

}
