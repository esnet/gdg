package alerting

import (
	"fmt"

	"github.com/esnet/gdg/internal/domain"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// routingTreeClient returns a namespaced dynamic client for RoutingTree resources,
// the resolved namespace, and the resolved API version string. The API group
// version is auto-discovered from the Grafana App Platform; domain.RoutingTreeVersion
// is used when discovery fails.
func (a *serviceImpl) routingTreeClient() (dynamic.ResourceInterface, string, string, error) {
	cfg, err := a.K8sRestConfig()
	if err != nil {
		return nil, "", "", err
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, "", "", fmt.Errorf("unable to build dynamic client: %w", err)
	}
	version := a.DiscoverAPIGroupVersion(domain.RoutingTreeGroup, domain.RoutingTreeVersion)
	gvr := schema.GroupVersionResource{
		Group:    domain.RoutingTreeGroup,
		Version:  version,
		Resource: domain.RoutingTreeResourceName,
	}
	ns := a.K8sNamespace()
	apiVersion := domain.RoutingTreeGroup + "/" + version
	return dyn.Resource(gvr).Namespace(ns), ns, apiVersion, nil
}
