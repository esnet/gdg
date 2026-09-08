package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"regexp"

	v2 "github.com/esnet/gdg/internal/adapter/filters/v2"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	templatesFile = "templates"
)

func setupTemplateReaders(filterObj outbound.Filter) {
	err := v2.RegisterTypedReader[*models.NotificationTemplate](filterObj, func(ctx context.Context, filterType domain.FilterType, val *models.NotificationTemplate) (any, error) {
		switch filterType {
		case domain.Name:
			return val.Name, nil
		default:
			return nil, fmt.Errorf("unsupported filter type: %s", filterType)
		}
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Alert Templates Filter, aborting.")
	}
}

func NewAlertTemplatesFilter(regexPattern string) outbound.Filter {
	filterEntity := v2.NewBaseFilter()
	setupTemplateReaders(filterEntity)
	v2.RegisterTypedValidation[string](filterEntity, domain.Name, regexPattern, func(ctx context.Context, val, expression string) error {
		if expression == "" {
			return nil
		}
		r, ReErr := regexp.Compile(expression)
		if ReErr != nil {
			return fmt.Errorf("invalid regex: %s", expression)
		}
		if r.MatchString(val) {
			return nil
		}
		return fmt.Errorf("invalid template filter. Expected: %v", expression)
	})
	return filterEntity
}

func (s *DashNGoImpl) DownloadAlertTemplates(filter outbound.Filter) (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	tpls, err := s.ListAlertTemplates(filter)
	if err != nil {
		return "", err
	}

	dsPath := s.resources.BuildResourcePath(s.grafanaConf, templatesFile, domain.AlertingResource, s.isLocal(), false)
	if dsPacked, err = json.MarshalIndent(tpls, "", "	"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}

	return dsPath, nil
}

func (s *DashNGoImpl) ListAlertTemplates(filter outbound.Filter) ([]*models.NotificationTemplate, error) {
	p := provisioning.NewGetTemplatesParams()
	tpl, err := s.GetClient().Provisioning.GetTemplatesWithParams(p)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	var result []*models.NotificationTemplate
	for _, item := range tpl.GetPayload() {
		if filter.Validate(ctx, domain.Name, item) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *DashNGoImpl) ClearAlertTemplates(filter outbound.Filter) ([]string, error) {
	tpls, err := s.ListAlertTemplates(filter)
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

func (s *DashNGoImpl) UploadAlertTemplates(filter outbound.Filter) ([]string, error) {
	var (
		err   error
		rawDS []byte
	)
	data := make([]*models.NotificationTemplate, 0)
	currentTemplates, err := s.ListAlertTemplates(filter)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*models.NotificationTemplate)
	for ndx, i := range currentTemplates {
		m[i.Name] = currentTemplates[ndx]
	}

	fileLocation := s.resources.BuildResourcePath(s.grafanaConf, templatesFile, domain.AlertingResource, s.isLocal(), false)
	if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
		return nil, fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if err = json.Unmarshal(rawDS, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}
	ctx := context.Background()
	var result []string
	for _, tpl := range data {
		if !filter.Validate(ctx, domain.Name, tpl) {
			slog.Debug("Skipping template, failed regex filter", slog.String("template", tpl.Name))
			continue
		}
		p := provisioning.NewPutTemplateParams()
		p.Name = tpl.Name
		p.XDisableProvenance = new("true")
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
