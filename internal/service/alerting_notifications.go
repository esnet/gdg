package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/esnet/gdg/internal/config"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	policiesFile = "policies"
)

func (s *DashNGoImpl) DownloadAlertNotifications() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	tpls, err := s.ListAlertNotifications()
	if err != nil {
		return "", err
	}

	dsPath := buildResourcePath(policiesFile, config.AlertingResource, s.isLocal(), false)
	if dsPacked, err = json.MarshalIndent(tpls, "", "	"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if err = s.storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}

	return dsPath, nil
}

func (s *DashNGoImpl) ListAlertNotifications() (*models.Route, error) {
	res, err := s.GetClient().Provisioning.GetPolicyTree()
	if err != nil {
		return nil, err
	}
	return res.GetPayload(), nil
}

func (s *DashNGoImpl) ClearAlertNotifications() error {
	_, err := s.GetClient().Provisioning.ResetPolicyTree()
	if err != nil {
		slog.Error("unable to reset policy tree")
		return err
	}

	return nil
}

func (s *DashNGoImpl) UploadAlertNotifications() (*models.Route, error) {
	var (
		err   error
		rawDS []byte
		data  *models.Route
	)

	fileLocation := buildResourcePath(policiesFile, config.AlertingResource, s.isLocal(), false)
	if rawDS, err = s.storage.ReadFile(fileLocation); err != nil {
		return nil, fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if err = json.Unmarshal(rawDS, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}
	p := provisioning.NewPutPolicyTreeParams()
	p.Body = data
	_, err = s.GetClient().Provisioning.PutPolicyTree(p)
	if err != nil {
		return nil, err
	}
	return s.ListAlertNotifications()
}
