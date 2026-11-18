// Package alerting implements outbound.AlertingService as a standalone service
// embedding core.Service for shared infrastructure. It covers every alerting
// resource: contact points (contactpoints.go), alert rules (rules.go,
// rules_filters.go), notification templates (templates.go), notification
// policies (notifications.go), and mute timings (timings.go).
//
// Contact point operations use the legacy provisioning API; the K8s receivers
// API does not expose decrypted secrets, making it unsuitable for
// backup/restore workflows.
package alerting

import (
	"github.com/esnet/gdg/internal/adapter/grafana/core"
	"github.com/esnet/gdg/internal/ports/outbound"
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
