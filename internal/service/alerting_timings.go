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
	timingsFile = "timings"
)

// DownloadAlertTimings retrieves alert timings, serializes to JSON, writes to file, and returns the file path or error.
func (s *DashNGoImpl) DownloadAlertTimings() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	timings, err := s.ListAlertTimings()
	if err != nil {
		return "", err
	}

	dsPath := buildResourcePath(timingsFile, domain.AlertingResource, s.isLocal(), false)
	if dsPacked, err = json.MarshalIndent(timings, "", "	"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}

	return dsPath, nil
}

// ListAlertTimings retrieves the current mute timing intervals for alerts.
func (s *DashNGoImpl) ListAlertTimings() ([]*models.MuteTimeInterval, error) {
	data, err := s.GetClient().Provisioning.GetMuteTimings()
	if err != nil {
		return nil, err
	}
	return data.GetPayload(), nil
}

// ClearAlertTimings deletes all alert timing configurations and returns any error encountered.
func (s *DashNGoImpl) ClearAlertTimings() error {
	currentTimings, err := s.ListAlertTimings()
	if err != nil {
		return err
	}
	for _, t := range currentTimings {
		p := provisioning.NewDeleteMuteTimingParams()
		p.Name = t.Name
		_, deleteErr := s.GetClient().Provisioning.DeleteMuteTiming(p)
		if deleteErr != nil {
			slog.Error(deleteErr.Error())
		}
	}
	return nil
}

// UploadAlertTimings uploads mute timing intervals from a JSON file to the provisioning API.
// It returns the names of successfully uploaded timings or an error if reading/parsing fails.
func (s *DashNGoImpl) UploadAlertTimings() ([]string, error) {
	var (
		err   error
		rawDS []byte
	)

	data := make([]*models.MuteTimeInterval, 0)
	currentTimings, err := s.ListAlertTimings()
	if err != nil {
		return nil, err
	}
	m := make(map[string]*models.MuteTimeInterval)
	for ndx, i := range currentTimings {
		m[i.Name] = currentTimings[ndx]
	}

	fileLocation := buildResourcePath(timingsFile, domain.AlertingResource, s.isLocal(), false)
	if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
		return nil, fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if err = json.Unmarshal(rawDS, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}

	var result []string
	for _, entry := range data {
		if _, ok := m[entry.Name]; ok {
			p := provisioning.NewPutMuteTimingParams()
			p.Body = entry
			p.XDisableProvenance = ptr.Of("true")
			_, err = s.GetClient().Provisioning.PutMuteTiming(p)
		} else {
			p := provisioning.NewPostMuteTimingParams()
			p.Body = entry
			p.XDisableProvenance = ptr.Of("true")
			_, err = s.GetClient().Provisioning.PostMuteTiming(p)
		}
		if err != nil {
			slog.Error("unable to upload template", "template", entry.Name, "err", err)
			continue
		}
		result = append(result, entry.Name)
	}
	return result, nil
}
