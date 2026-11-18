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
//
// Every method accepts a name filter: pass "" to operate on every routing
// tree, or a specific metadata.name to scope the operation to just that tree.
//
// ClearRoutingTrees deletes each matching tree; when the server rejects the
// delete (the org's protected default tree typically cannot be removed), it
// is reset to domain.DefaultRoutingTreeSpec instead of being skipped.
//
// UploadRoutingTrees makes the server fully reflect local storage: it
// creates/updates every matching local tree, then removes any matching
// server-side tree that has no local file, using the same delete-or-reset
// logic as ClearRoutingTrees for trees the server refuses to delete.
type AlertRoutingTrees interface {
	ListRoutingTrees(name string) ([]*customModels.RoutingTreeGdg, error)
	DownloadRoutingTrees(name string) ([]string, error)
	ClearRoutingTrees(name string) ([]string, error)
	UploadRoutingTrees(name string) ([]string, error)
}
