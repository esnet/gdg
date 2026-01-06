package service

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/esnet/gdg/pkg/config/domain"

	"github.com/esnet/gdg/internal/storage"

	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	emailReceiver = "email receiver"
	contactsFile  = "contacts"
)

func (s *DashNGoImpl) ListContactPoints() ([]*models.EmbeddedContactPoint, error) {
	p := provisioning.NewGetContactpointsParams()
	result, err := s.GetClient().Provisioning.GetContactpoints(p)
	if err != nil {
		return nil, err
	}
	data := lo.Filter(result.GetPayload(), func(item *models.EmbeddedContactPoint, index int) bool {
		return item.UID != "" && item.Name != emailReceiver
	})

	return data, nil
}

func (s *DashNGoImpl) DownloadContactPoints() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	p := provisioning.NewGetContactpointsExportParams()
	p.Download = ptr.Of(true)
	p.Decrypt = ptr.Of(true)
	p.Format = ptr.Of("json")
	data, err := s.GetClient().Provisioning.GetContactpointsExport(p)
	if err != nil {
		log.Fatalf("unable to retrieve Contact Points, err: %s", err.Error())
	}
	// filter default contactPoints
	payload := data.GetPayload()
	payload.ContactPoints = lo.Filter(payload.ContactPoints, func(item *models.ContactPointExport, index int) bool {
		return item.Name != emailReceiver
	})

	dsPath := buildResourcePath(s.grafanaConf, contactsFile, domain.AlertingResource, s.isLocal(), false)
	if dsPacked, err = json.MarshalIndent(payload.ContactPoints, "", "	"); err != nil {
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

func (s *DashNGoImpl) isLocal() bool {
	return s.storage.Name() == storage.LocalStorageType.String()
}

func (s *DashNGoImpl) UploadContactPoints() ([]string, error) {
	var (
		err    error
		rawDS  []byte
		result []string
	)
	var data []models.ContactPointExport
	currentContacts, err := s.ListContactPoints()
	if err != nil {
		return nil, err
	}
	m := make(map[string]*models.EmbeddedContactPoint)
	for ndx, i := range currentContacts {
		m[i.UID] = currentContacts[ndx]
	}

	fileLocation := buildResourcePath(s.grafanaConf, contactsFile, domain.AlertingResource, s.isLocal(), false)
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
	for _, i := range data {
		for _, r := range i.Receivers {
			if r.UID == "" {
				slog.Info("No valid UID found for record, skipping", slog.Any("type", r.Type))
				continue
			}
			if _, ok := m[r.UID]; ok {
				// do update
				p := provisioning.NewPutContactpointParams()
				p.UID = r.UID
				p.XDisableProvenance = ptr.Of("true")
				p.Body = &models.EmbeddedContactPoint{
					DisableResolveMessage: false,
					Name:                  i.Name,
					Provenance:            "",
					Settings:              r.Settings,
					Type:                  ptr.Of(r.Type),
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
				p.XDisableProvenance = ptr.Of("true")
				p.Body = &models.EmbeddedContactPoint{
					DisableResolveMessage: false,
					Name:                  i.Name,
					UID:                   r.UID,
					Provenance:            "",
					Settings:              r.Settings,
					Type:                  ptr.Of(r.Type),
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
		_, err = s.GetClient().Provisioning.DeleteContactpoints(contact.UID)
		if err != nil {
			slog.Error("unable to delete contact point",
				slog.Any("name", contact.Name),
				slog.Any("uid", contact.UID),
			)
			continue
		}
		results = append(results, contact.Name)
	}

	return results, nil
}
