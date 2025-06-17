package service

import (
	"errors"
	"fmt"

	"github.com/grafana/grafana-openapi-client-go/models"
)

// GetOrgPreferences returns the preferences for a given Org
// orgName: The name of the organization whose preferences we should retrieve
func (s *DashNGoImpl) GetOrgPreferences(orgName string) (*models.Preferences, error) {
	if !s.grafanaConf.IsGrafanaAdmin() {
		return nil, errors.New("no valid Grafana Admin configured, cannot retrieve Organizations Preferences")
	}
	orgPreferences, err := s.GetBasicClientWithOpts(GetOrgNameClientOpts(orgName)).OrgPreferences.GetOrgPreferences()
	if err != nil {
		return nil, err
	}
	return orgPreferences.GetPayload(), nil
}

// UploadOrgPreferences Updates the preferences for a given organization.  Returns error if org is not found.
func (s *DashNGoImpl) UploadOrgPreferences(orgName string, preferenceRequest *models.Preferences) error {
	if !s.grafanaConf.IsGrafanaAdmin() {
		return errors.New("no valid Grafana Admin configured, cannot update Organizations Preferences")
	}

	if preferenceRequest == nil {
		return fmt.Errorf("preferences are nil, cannot update")
	}

	update := &models.UpdatePrefsCmd{}
	update.HomeDashboardUID = preferenceRequest.HomeDashboardUID
	update.Language = preferenceRequest.Language
	update.Timezone = preferenceRequest.Timezone
	update.Theme = preferenceRequest.Theme
	update.WeekStart = preferenceRequest.WeekStart

	_, err := s.GetBasicClientWithOpts(GetOrgNameClientOpts(orgName)).OrgPreferences.UpdateOrgPreferences(update)
	if err != nil {
		return err
	}

	return nil
}
