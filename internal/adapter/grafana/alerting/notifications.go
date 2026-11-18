package alerting

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/esnet/gdg/internal/domain"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	policiesFile = "policies"
)

// DownloadAlertNotifications retrieves alert notifications, serializes them to JSON,
// writes the file to disk/s3, and returns the file path or an error.
func (a *serviceImpl) DownloadAlertNotifications() (string, error) {
	var (
		dsPacked []byte
		err      error
	)
	tpls, err := a.ListAlertNotifications()
	if err != nil {
		return "", err
	}

	dsPath := a.Resources.BuildResourcePath(a.GrafanaConf, policiesFile, domain.AlertingResource, a.IsLocal(), false)
	if dsPacked, err = json.MarshalIndent(tpls, "", "	"); err != nil {
		return "", fmt.Errorf("unable to serialize data to JSON. %w", err)
	}
	if err = a.Storage.WriteFile(dsPath, dsPacked); err != nil {
		return "", fmt.Errorf("unable to write file. %w", err)
	}

	return dsPath, nil
}

// ListAlertNotifications retrieves the current alert notification policy tree.
func (a *serviceImpl) ListAlertNotifications() (*models.Route, error) {
	res, err := a.GetClient().Provisioning.GetPolicyTree()
	if err != nil {
		return nil, err
	}
	return res.GetPayload(), nil
}

// ClearAlertNotifications resets the policy tree to clear alert notifications.
func (a *serviceImpl) ClearAlertNotifications() error {
	_, err := a.GetClient().Provisioning.ResetPolicyTree()
	if err != nil {
		slog.Error("unable to reset policy tree")
		return err
	}

	return nil
}

// UploadAlertNotifications uploads alert notification policies from file to Grafana and returns updated list.
func (a *serviceImpl) UploadAlertNotifications() (*models.Route, error) {
	var (
		err   error
		rawDS []byte
		data  *models.Route
	)

	fileLocation := a.Resources.BuildResourcePath(a.GrafanaConf, policiesFile, domain.AlertingResource, a.IsLocal(), false)
	if rawDS, err = a.Storage.ReadFile(fileLocation); err != nil {
		return nil, fmt.Errorf("failed to read file.  file: %s, err: %w", fileLocation, err)
	}
	if err = json.Unmarshal(rawDS, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshall file, file:%s, err: %w", fileLocation, err)
	}
	p := provisioning.NewPutPolicyTreeParams()
	p.XDisableProvenance = new("true")
	p.Body = data
	_, err = a.GetClient().Provisioning.PutPolicyTree(p)
	if err != nil {
		return nil, err
	}
	return a.ListAlertNotifications()
}
