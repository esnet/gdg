package service

import (
	"errors"
	"fmt"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
	"log/slog"
)

// OrgPreferencesApi Contract definition
type OrgPreferencesApi interface {
	GetOrgPreferences(orgName string) (*models.Preferences, error)
	UploadOrgPreferences(orgName string, pref *models.Preferences) error
}

// GetOrgPreferences returns the preferences for a given Org
// orgName: The name of the organization whose preferences we should retrieve
func (s *DashNGoImpl) GetOrgPreferences(orgName string) (*models.Preferences, error) {
	if !s.grafanaConf.IsAdminEnabled() {
		return nil, errors.New("no valid Grafana Admin configured, cannot retrieve Organizations Preferences")
	}
	f := func() (interface{}, error) {
		orgPreferences, err := s.GetClient().OrgPreferences.GetOrgPreferences()
		if err != nil {
			return nil, err
		}
		return orgPreferences.GetPayload(), nil
	}
	result, err := s.scopeIntoOrg(orgName, f)
	if err != nil {
		return nil, err
	}
	return result.(*models.Preferences), nil
}

// scopeIntoOrg changes the organization, performs an operation, and reverts the Org to the previous value.
func (s *DashNGoImpl) scopeIntoOrg(orgName string, runTask func() (interface{}, error)) (interface{}, error) {
	currentOrg := s.getAssociatedActiveOrg(s.GetClient())
	orgNameBackup := s.grafanaConf.OrganizationName
	s.grafanaConf.OrganizationName = orgName
	orgEntity, err := s.getOrgIdFromSlug(slug.Make(orgName))
	if err != nil {
		return nil, err
	}
	defer func() {
		s.grafanaConf.OrganizationName = orgNameBackup
		//restore scoped Org
		err = s.SetUserOrganizations(currentOrg.ID)
		if err != nil {
			slog.Warn("unable to restore previous Org")
		}
	}()

	err = s.SetUserOrganizations(orgEntity.OrgID)
	if err != nil {
		return nil, fmt.Errorf("unable to scope into requested org. %w", err)
	}

	res, err := runTask()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// UploadOrgPreferences Updates the preferences for a given organization.  Returns error if org is not found.
func (s *DashNGoImpl) UploadOrgPreferences(orgName string, pref *models.Preferences) error {
	runTask := func() (interface{}, error) {
		if pref == nil {
			return nil, fmt.Errorf("preferences are nil, cannot update")
		}

		update := &models.UpdatePrefsCmd{}
		update.HomeDashboardUID = pref.HomeDashboardUID
		update.Language = pref.Language
		update.Timezone = pref.Timezone
		update.Theme = pref.Theme
		update.WeekStart = pref.WeekStart

		status, err := s.GetClient().OrgPreferences.UpdateOrgPreferences(update)
		if err != nil {
			return nil, err
		}
		return status, nil
	}
	_, err := s.scopeIntoOrg(orgName, runTask)
	if err != nil {
		return err
	}
	slog.Info("Organization Preferences were updated")
	return nil
}
