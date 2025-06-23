package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/esnet/gdg/internal/config/domain"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	templatesFile = "templates"
)

func (s *DashNGoImpl) DownloadAlertTemplates() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	tpls, err := s.ListAlertTemplates()
	if err != nil {
		return "", err
	}

	dsPath := buildResourcePath(templatesFile, domain.AlertingResource, s.isLocal(), false)
	if dsPacked, err = json.MarshalIndent(tpls, "", "	"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}

	return dsPath, nil
}

func (s *DashNGoImpl) ListAlertTemplates() ([]*models.NotificationTemplate, error) {
	p := provisioning.NewGetTemplatesParams()
	tpl, err := s.GetClient().Provisioning.GetTemplatesWithParams(p)
	if err != nil {
		return nil, err
	}
	return tpl.GetPayload(), nil
}

func (s *DashNGoImpl) ClearAlertTemplates() ([]string, error) {
	tpls, err := s.ListAlertTemplates()
	if err != nil {
		return nil, err
	}
	var result []string
	for _, tpl := range tpls {
		p := provisioning.NewDeleteTemplateParams()
		p.Name = tpl.Name
		_, err = s.GetClient().Provisioning.DeleteTemplate(p)
		if err != nil {
			slog.Error("unable to delete template", "template", tpl.Name)
			continue
		}
		result = append(result, tpl.Name)
	}
	return result, nil
}

func (s *DashNGoImpl) UploadAlertTemplates() ([]string, error) {
	var (
		err   error
		rawDS []byte
	)
	data := make([]*models.NotificationTemplate, 0)
	currentTemplates, err := s.ListAlertTemplates()
	if err != nil {
		return nil, err
	}
	m := make(map[string]*models.NotificationTemplate)
	for ndx, i := range currentTemplates {
		m[i.Name] = currentTemplates[ndx]
	}

	fileLocation := buildResourcePath(templatesFile, domain.AlertingResource, s.isLocal(), false)
	if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
		return nil, fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if err = json.Unmarshal(rawDS, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}
	var result []string
	for _, tpl := range data {
		p := provisioning.NewPutTemplateParams()
		p.Name = tpl.Name
		p.Body = &models.NotificationTemplateContent{Template: tpl.Template}
		if val, ok := m[p.Name]; ok {
			p.Body.Version = val.Version
		}
		_, err = s.GetClient().Provisioning.PutTemplate(p)
		if err != nil {
			slog.Error("unable to upload template", "template", p.Name, "err", err)
			continue
		}
		result = append(result, p.Name)
	}
	return result, nil
}
