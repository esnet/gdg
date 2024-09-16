package service

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/tools"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

func (s *DashNGoImpl) ListContactPoints() []*models.EmbeddedContactPoint {
	p := provisioning.NewGetContactpointsParams()
	result, err := s.GetClient().Provisioning.GetContactpoints(p)
	if err != nil {
		log.Fatalf("unable to retrieve contact points, err:%s", err.Error())
	}
	return result.GetPayload()
}

func (s *DashNGoImpl) DownloadContactPoints() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	p := provisioning.NewGetContactpointsExportParams()
	p.Download = tools.PtrOf(true)
	p.Decrypt = tools.PtrOf(true)
	p.Format = tools.PtrOf("json")
	data, err := s.GetClient().Provisioning.GetContactpointsExport(p)
	if err != nil {
		log.Fatalf("unable to retrieve Contact Points, err: %s", err.Error())
	}

	dsPath := buildResourcePath("contacts", config.AlertingResource)
	if dsPacked, err = json.MarshalIndent(data.GetPayload(), "", "	"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}

	return dsPath, nil
}

func (s *DashNGoImpl) UploadContactPoints() ([]string, error) {
	var (
		err    error
		rawDS  []byte
		result []string
	)
	data := new(models.AlertingFileExport)
	currentContacts := s.ListContactPoints()
	m := make(map[string]*models.EmbeddedContactPoint)
	for ndx, i := range currentContacts {
		m[i.UID] = currentContacts[ndx]
	}

	fileLocation := buildResourcePath("contacts", config.AlertingResource)
	if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
		return nil, fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if err = json.Unmarshal(rawDS, data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}
	for _, i := range data.ContactPoints {
		for _, r := range i.Receivers {
			if _, ok := m[r.UID]; ok {
				// do update
				p := provisioning.NewPutContactpointParams()
				p.UID = r.UID
				p.Body = &models.EmbeddedContactPoint{
					DisableResolveMessage: false,
					Name:                  i.Name,
					Provenance:            "",
					Settings:              r.Settings,
					Type:                  tools.PtrOf(r.Type),
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
				p.Body = &models.EmbeddedContactPoint{
					DisableResolveMessage: false,
					Name:                  i.Name,
					UID:                   r.UID,
					Provenance:            "",
					Settings:              r.Settings,
					Type:                  tools.PtrOf(r.Type),
				}
				_, err := s.GetClient().Provisioning.PostContactpoints(p)
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
	contacts := s.ListContactPoints()
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
