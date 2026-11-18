// Package alerting implements outbound.AlertingService (currently contact
// points) as a standalone service embedding core.Service for shared
// infrastructure. All contact point operations use the legacy provisioning
// API; the K8s receivers API does not expose decrypted secrets, making it
// unsuitable for backup/restore workflows.
package alerting

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/esnet/gdg/internal/adapter/grafana/core"
	"github.com/esnet/gdg/internal/config/config_tooling"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// Contact point constants shared across the legacy and (future) K8s implementation paths.
const (
	emailReceiver = "email receiver"
	contactsFile  = "contacts"
)

// Compile-time assertion: serviceImpl satisfies outbound.AlertingService.
var _ outbound.AlertingService = (*serviceImpl)(nil)

// serviceImpl is the standalone implementation of outbound.AlertingService.
// It embeds core.Service for all shared infrastructure — there is no
// back-reference to the top-level api.DashNGoImpl.
type serviceImpl struct {
	core.Service
}

// NewService constructs a serviceImpl from a core.Service value.
func NewService(base core.Service) outbound.AlertingService {
	return &serviceImpl{Service: base}
}

// ---------------------------------------------------------------------------
// AlertContactPoints — public CRUD methods (provisioning API)
// ---------------------------------------------------------------------------

func (a *serviceImpl) ListContactPoints() ([]*models.ContactPointExport, error) {
	p := provisioning.NewGetContactpointsExportParams()
	p.Download = new(false)
	p.Decrypt = new(false)
	p.Format = new("json")
	result, err := a.GetClient().Provisioning.GetContactpointsExport(p)
	if err != nil {
		return nil, err
	}

	data := make([]*models.ContactPointExport, 0)
	connSettings := a.GrafanaConf.GetAlertSettings().ContactSettings
	for _, item := range result.GetPayload().ContactPoints {
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

func (a *serviceImpl) DownloadContactPoints() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	p := provisioning.NewGetContactpointsExportParams()
	p.Download = new(true)
	p.Decrypt = new(true)
	p.Format = new("json")
	data, err := a.GetClient().Provisioning.GetContactpointsExport(p)
	if err != nil {
		log.Fatalf("unable to retrieve Contact Points, err: %s", err.Error())
	}

	connSettings := a.GrafanaConf.GetAlertSettings().ContactSettings
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

	dsPath := a.Resources.BuildResourcePath(a.GrafanaConf, contactsFile, domain.AlertingResource, a.IsLocal(), false)
	if dsPacked, err = json.MarshalIndent(filtered, "", "\t"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if a.GdgConfig.PluginConfig.CipherEnabled() {
		newData, encodeErr := a.Encoder.Encode(domain.AlertingResource, dsPacked)
		if encodeErr != nil {
			slog.Error("unable to encode sensitive data using cipher plugin. All data was saved in plaintext.", "err", encodeErr)
		}
		dsPacked = newData
	}
	if err = a.Storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}
	return dsPath, nil
}

func (a *serviceImpl) UploadContactPoints() ([]string, error) {
	var (
		rawDS  []byte
		result []string
	)
	var data []models.ContactPointExport
	currentContacts, err := a.ListContactPoints()
	if err != nil {
		return nil, err
	}
	m := make(map[string]*models.ReceiverExport)
	for ndx, i := range currentContacts {
		for _, item := range i.Receivers {
			m[item.UID] = i.Receivers[ndx]
		}
	}

	fileLocation := a.Resources.BuildResourcePath(a.GrafanaConf, contactsFile, domain.AlertingResource, a.IsLocal(), false)
	if rawDS, err = a.Storage.ReadFile(fileLocation); err != nil {
		return nil, fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if a.GdgConfig.PluginConfig.CipherEnabled() {
		newData, encodeErr := a.Encoder.Decode(domain.AlertingResource, rawDS)
		if encodeErr != nil {
			slog.Error("unable to encode sensitive data using cipher plugin. All data was saved in plaintext. ", "err", encodeErr)
		}
		rawDS = newData
	}
	if err = json.Unmarshal(rawDS, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}
	connSettings := a.GrafanaConf.GetAlertSettings().ContactSettings
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
				if _, putErr := a.GetClient().Provisioning.PutContactpoint(p); putErr != nil {
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
				if _, postErr := a.GetClient().Provisioning.PostContactpoints(p); postErr != nil {
					slog.Error("failed to create contact point", slog.Any("uid", r.UID))
					continue
				}
				result = append(result, i.Name)
			}
		}
	}
	return result, nil
}

func (a *serviceImpl) ClearContactPoints() ([]string, error) {
	contacts, err := a.ListContactPoints()
	if err != nil {
		return nil, err
	}
	var results []string
	for _, contact := range contacts {
		for _, receiver := range contact.Receivers {
			if _, delErr := a.GetClient().Provisioning.DeleteContactpoints(receiver.UID); delErr != nil {
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
