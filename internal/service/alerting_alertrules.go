package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/esnet/gdg/internal/config/domain"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	rulesFile = "rules"
)

func (s *DashNGoImpl) ListAlertRules() ([]*models.ProvisionedAlertRule, error) {
	data, err := s.GetClient().Provisioning.GetAlertRules()
	if err != nil {
		return nil, err
	}
	return data.GetPayload(), nil
}

func (s *DashNGoImpl) UploadAlertRules() error {
	var (
		err   error
		rawDS []byte
	)
	data := make([]*models.ProvisionedAlertRule, 0)
	currentContacts, err := s.ListAlertRules()
	if err != nil {
		return err
	}
	m := make(map[string]*models.ProvisionedAlertRule)
	for ndx, i := range currentContacts {
		m[i.UID] = currentContacts[ndx]
	}

	fileLocation := buildResourcePath(rulesFile, domain.AlertingResource, s.isLocal(), false)
	if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
		return fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if err = json.Unmarshal(rawDS, &data); err != nil {
		return fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}
	for _, group := range data {
		p := provisioning.NewPostAlertRuleParams()
		p.Body = group
		if _, ok := m[group.UID]; ok {
			// delete previous rule
			pdel := provisioning.NewDeleteAlertRuleParams()
			pdel.UID = group.UID
			_, delErr := s.GetClient().Provisioning.DeleteAlertRule(pdel)
			if delErr != nil {
				slog.Error("unable to delete previous data, skipping rule update", "uid", group.UID, "err", err)
				continue
			}
		}
		_, err = s.GetClient().Provisioning.PostAlertRule(p)
		if err != nil {
			slog.Error("unable to import rule", "uid", group.UID, "err", err)
		}
	}
	return nil
}

func (s *DashNGoImpl) DownloadAlertRules() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	//p := provisioning.NewGetAlertRuleExportParams()
	//p.Format = ptr.Of("json")
	//p.Download = ptr.Of(true)
	//data, err := s.GetClient().Provisioning.GetAlertRuleExport(p)
	//if err != nil {
	//	return "", err
	//}
	data, err := s.ListAlertRules()
	if err != nil {
		return "", err
	}

	dsPath := buildResourcePath(rulesFile, domain.AlertingResource, s.isLocal(), false)
	if dsPacked, err = json.MarshalIndent(data, "", "	"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}

	return dsPath, nil
}

func (s *DashNGoImpl) ClearAlertRules() ([]string, error) {
	rules, err := s.ListAlertRules()
	if err != nil {
		return nil, err
	}
	var data []string
	for _, rule := range rules {
		p := provisioning.NewDeleteAlertRuleParams()
		p.UID = rule.UID
		_, err := s.GetClient().Provisioning.DeleteAlertRule(p)
		if err != nil {
			slog.Error("unable to delete rule", "rule", rule.UID)
			continue
		}
		data = append(data, ptr.ValOf(rule.Title))
	}

	return data, nil
}
