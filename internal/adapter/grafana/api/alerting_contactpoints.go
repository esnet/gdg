package api

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/esnet/gdg/internal/config/config_tooling"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	emailReceiver = "email receiver"
	contactsFile  = "contacts"
)

var _ outbound.AlertContactPoints = (*DashNGoImpl)(nil)

func (s *DashNGoImpl) ListContactPoints() ([]*models.ContactPointExport, error) {
	p := provisioning.NewGetContactpointsExportParams()
	p.Download = new(false)
	p.Decrypt = new(false)
	p.Format = new("json")
	result, err := s.GetClient().Provisioning.GetContactpointsExport(p)
	if err != nil {
		return nil, err
	}

	data := make([]*models.ContactPointExport, 0)
	connSettings := s.grafanaConf.GetAlertSettings().ContactSettings
	for _, item := range result.GetPayload().ContactPoints {
		// Skipping broken email Receiver from v12 can't manage
		if item.Name == emailReceiver && len(item.Receivers) > 0 && item.Receivers[0].UID == "" {
			continue
		}

		if connSettings.FiltersEnabled() && config_tooling.IsExcluded(item, connSettings.FilterRules) {
			slog.Debug("Skipping contact point, since it fails filter checks", "contact_point", item.Name)
			continue
		}
		data = append(data, item)

	}

	return data, nil
}

func (s *DashNGoImpl) DownloadContactPoints() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	p := provisioning.NewGetContactpointsExportParams()
	p.Download = new(true)
	p.Decrypt = new(true)
	p.Format = new("json")
	data, err := s.GetClient().Provisioning.GetContactpointsExport(p)
	if err != nil {
		log.Fatalf("unable to retrieve Contact Points, err: %s", err.Error())
	}

	// Apply the same email-receiver skip and contact-point filters as
	// ListContactPoints so Download is consistent with every other operation.
	connSettings := s.grafanaConf.GetAlertSettings().ContactSettings
	var filtered []*models.ContactPointExport
	for _, item := range data.GetPayload().ContactPoints {
		if item.Name == emailReceiver && len(item.Receivers) > 0 && item.Receivers[0].UID == "" {
			continue
		}
		if connSettings.FiltersEnabled() && config_tooling.IsExcluded(item, connSettings.FilterRules) {
			slog.Debug("Skipping contact point, since it fails filter checks", "contact_point", item.Name)
			continue
		}
		filtered = append(filtered, item)
	}

	dsPath := s.resources.BuildResourcePath(s.grafanaConf, contactsFile, domain.AlertingResource, s.isLocal(), false)
	if dsPacked, err = json.MarshalIndent(filtered, "", "	"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if !s.gdgConfig.PluginConfig.Disabled && s.gdgConfig.PluginConfig.CipherPlugin != nil {
		newData, encodeErr := s.encoder.Encode(domain.AlertingResource, dsPacked)
		if encodeErr != nil {
			slog.Error("unable to encode sensitive data using cipher plugin. All data was saved in plaintext.", "err", encodeErr)
		}
		dsPacked = newData
	}
	if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}

	return dsPath, nil
}

func (s *DashNGoImpl) UploadContactPoints() ([]string, error) {
	var (
		rawDS  []byte
		result []string
	)
	var data []models.ContactPointExport
	currentContacts, err := s.ListContactPoints()
	if err != nil {
		return nil, err
	}
	m := make(map[string]*models.ReceiverExport)
	for ndx, i := range currentContacts {
		for _, item := range i.Receivers {
			m[item.UID] = i.Receivers[ndx] //currentContacts[ndx]
		}
	}

	fileLocation := s.resources.BuildResourcePath(s.grafanaConf, contactsFile, domain.AlertingResource, s.isLocal(), false)
	if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
		return nil, fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if !s.gdgConfig.PluginConfig.Disabled && s.gdgConfig.PluginConfig.CipherPlugin != nil {
		newData, encodeErr := s.encoder.Decode(domain.AlertingResource, rawDS)
		if encodeErr != nil {
			slog.Error("unable to encode sensitive data using cipher plugin. All data was saved in plaintext. ", "err", encodeErr)
		}
		rawDS = newData
	}
	if err = json.Unmarshal(rawDS, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}
	connSettings := s.grafanaConf.GetAlertSettings().ContactSettings
	for _, i := range data {
		if connSettings.FiltersEnabled() && config_tooling.IsExcluded(i, connSettings.FilterRules) {
			slog.Debug("Skipping local JSON file since source fails datatype filter checks", "datasource", i.Name)
			continue
		}
		for _, r := range i.Receivers {
			if r.UID == "" {
				slog.Info("No valid UID found for record, skipping", slog.Any("type", r.Type))
				continue
			}
			if _, ok := m[r.UID]; ok {
				// do update
				p := provisioning.NewPutContactpointParams()
				p.UID = r.UID
				p.XDisableProvenance = new("true")
				p.Body = &models.EmbeddedContactPoint{
					DisableResolveMessage: false,
					Name:                  i.Name,
					Provenance:            "",
					Settings:              r.Settings,
					Type:                  new(r.Type),
					UID:                   r.UID,
				}
				_, err := s.GetClient().Provisioning.PutContactpoint(p)
				if err != nil {
					slog.Error("failed to update contact point", slog.Any("uid", r.UID))
					continue
				}
				result = append(result, i.Name)

			} else {
				p := provisioning.NewPostContactpointsParams()
				p.XDisableProvenance = new("true")
				p.Body = &models.EmbeddedContactPoint{
					DisableResolveMessage: false,
					Name:                  i.Name,
					UID:                   r.UID,
					Provenance:            "",
					Settings:              r.Settings,
					Type:                  new(r.Type),
				}
				_, err = s.GetClient().Provisioning.PostContactpoints(p)
				if err != nil {
					slog.Error("failed to create contact point", slog.Any("uid", r.UID))
					continue
				}

				result = append(result, i.Name)
			}
		}
	}

	return result, nil
}

func (s *DashNGoImpl) ClearContactPoints() ([]string, error) {
	var (
		err     error
		results []string
	)
	contacts, err := s.ListContactPoints()
	if err != nil {
		return nil, err
	}

	for _, contact := range contacts {
		for _, receiver := range contact.Receivers {
			_, err = s.GetClient().Provisioning.DeleteContactpoints(receiver.UID)
			if err != nil {
				slog.Error("unable to delete contact point",
					slog.Any("name", contact.Name),
					slog.Any("uid", receiver.UID),
				)
				continue
			}
			results = append(results, contact.Name)
		}
	}

	return results, nil
}
