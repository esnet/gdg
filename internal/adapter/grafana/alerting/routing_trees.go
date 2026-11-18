package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/esnet/gdg/internal/domain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListRoutingTrees lists every RoutingTree resource in the configured namespace
// via the App Platform (/apis) notifications.alerting.grafana.app API.
func (a *serviceImpl) ListRoutingTrees() ([]*domain.RoutingTreeGdg, error) {
	client, _, apiVersion, err := a.routingTreeClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	list, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list routing trees via app platform api: %w", err)
	}

	results := make([]*domain.RoutingTreeGdg, 0, len(list.Items))
	for i := range list.Items {
		results = append(results, &domain.RoutingTreeGdg{
			Resource:      list.Items[i].Object,
			GdgApiVersion: apiVersion,
		})
	}
	return results, nil
}

// DownloadRoutingTrees saves every routing tree to the configured storage
// location, one file per tree named after its metadata.name.
func (a *serviceImpl) DownloadRoutingTrees() ([]string, error) {
	trees, err := a.ListRoutingTrees()
	if err != nil {
		return nil, err
	}

	var files []string
	for _, tree := range trees {
		name := tree.Name()
		if name == "" {
			slog.Warn("skipping routing tree without a name")
			continue
		}

		dsPacked, marshalErr := json.MarshalIndent(tree, "", "\t")
		if marshalErr != nil {
			return nil, fmt.Errorf("unable to serialize data to JSON. %w", marshalErr)
		}

		dsPath := a.Resources.BuildResourcePath(a.GrafanaConf, name, domain.RoutingTreeResource, a.IsLocal(), false)
		if writeErr := a.Storage.WriteFile(dsPath, dsPacked); writeErr != nil {
			return nil, fmt.Errorf("unable to write file. %w", writeErr)
		}
		files = append(files, dsPath)
	}
	return files, nil
}
