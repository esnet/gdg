package outbound

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
	AlertRoutingTrees
}

// AlertingService is the standalone port satisfied by the alerting package's
// ServiceImpl. It covers every alerting operation — contact points, rules,
// templates, notification policies, and mute timings.
type AlertingService interface {
	AlertingApi
}

type AlertRules interface {
	DownloadAlertRules(filter Filter) ([]string, error)
	ListAlertRules(filter Filter) ([]*customModels.AlertRuleWithNestedFolder, error)
	ClearAlertRules(filter Filter) ([]string, error)
	UploadAlertRules(filter Filter) ([]*customModels.AlertRuleWithNestedFolder, error)
}

type AlertContactPoints interface {
	ListContactPoints() ([]*models.ContactPointExport, error)
	DownloadContactPoints() (string, error)
	ClearContactPoints() ([]string, error)
	UploadContactPoints() ([]string, error)
}

type AlertTemplates interface {
	DownloadAlertTemplates(filter Filter) (string, error)
	ListAlertTemplates(filter Filter) ([]*models.NotificationTemplate, error)
	ClearAlertTemplates(filter Filter) ([]string, error)
	UploadAlertTemplates(filter Filter) ([]string, error)
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

// AlertRoutingTrees covers the App Platform (K8s-style) notification routing
// tree API (notifications.alerting.grafana.app/v1beta1, resource
// "routingtrees"). Unlike AlertPolicies, which manages a single flat policy
// tree via the legacy provisioning API, this supports multiple named routing
// trees per namespace/org and exposes the full recursive route structure.
// Read-only for now: list and download only.
type AlertRoutingTrees interface {
	ListRoutingTrees() ([]*customModels.RoutingTreeGdg, error)
	DownloadRoutingTrees() ([]string, error)
}
