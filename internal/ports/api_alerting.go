package ports

import (
	customModels "github.com/esnet/gdg/internal/domain"
	"github.com/grafana/grafana-openapi-client-go/models"
)

type AlertingApi interface {
	AlertContactPoints
	AlertRules
	AlertTemplates
	AlertPolicies
	AlertTimings
}

type AlertRules interface {
	DownloadAlertRules(filter Filter) ([]string, error)
	ListAlertRules(filter Filter) ([]*customModels.AlertRuleWithNestedFolder, error)
	ClearAlertRules(filter Filter) ([]string, error)
	UploadAlertRules(filter Filter) ([]*customModels.AlertRuleWithNestedFolder, error)
}

type AlertContactPoints interface {
	ListContactPoints() ([]*models.EmbeddedContactPoint, error)
	DownloadContactPoints() (string, error)
	ClearContactPoints() ([]string, error)
	UploadContactPoints() ([]string, error)
}

type AlertTemplates interface {
	DownloadAlertTemplates() (string, error)
	ListAlertTemplates() ([]*models.NotificationTemplate, error)
	ClearAlertTemplates() ([]string, error)
	UploadAlertTemplates() ([]string, error)
}

type AlertPolicies interface {
	DownloadAlertNotifications() (string, error)
	ListAlertNotifications() (*models.Route, error)
	ClearAlertNotifications() error
	UploadAlertNotifications() (*models.Route, error)
}

type AlertTimings interface {
	DownloadAlertTimings() (string, error)
	ListAlertTimings() ([]*models.MuteTimeInterval, error)
	ClearAlertTimings() error
	UploadAlertTimings() ([]string, error)
}
