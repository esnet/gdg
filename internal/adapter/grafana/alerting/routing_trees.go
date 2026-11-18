package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/esnet/gdg/internal/domain"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// deleteOrResetRoutingTree deletes a routing tree, falling back to resetting it
// to domain.DefaultRoutingTreeSpec when the server refuses the delete (e.g. the
// org's protected/default tree). Shared by ClearRoutingTrees and the
// UploadRoutingTrees prune step.
func (a *serviceImpl) deleteOrResetRoutingTree(ctx context.Context, client dynamic.ResourceInterface, name string) error {
	if delErr := client.Delete(ctx, name, metav1.DeleteOptions{}); delErr == nil {
		return nil
	}
	return a.resetRoutingTree(ctx, client, name)
}

// ListRoutingTrees lists RoutingTree resources in the configured namespace via
// the App Platform (/apis) notifications.alerting.grafana.app API. If name is
// non-empty, only the tree with that exact metadata.name is returned.
func (a *serviceImpl) ListRoutingTrees(name string) ([]*domain.RoutingTreeGdg, error) {
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
		gdg := &domain.RoutingTreeGdg{
			Resource:      list.Items[i].Object,
			GdgApiVersion: apiVersion,
		}
		if name != "" && gdg.Name() != name {
			continue
		}
		results = append(results, gdg)
	}
	return results, nil
}

// DownloadRoutingTrees saves matching routing trees to the configured storage
// location, one file per tree named after its metadata.name. If name is
// non-empty, only that tree is downloaded.
func (a *serviceImpl) DownloadRoutingTrees(name string) ([]string, error) {
	trees, err := a.ListRoutingTrees(name)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, tree := range trees {
		treeName := tree.Name()
		if treeName == "" {
			slog.Warn("skipping routing tree without a name")
			continue
		}

		dsPacked, marshalErr := json.MarshalIndent(tree, "", "\t")
		if marshalErr != nil {
			return nil, fmt.Errorf("unable to serialize data to JSON. %w", marshalErr)
		}

		dsPath := a.Resources.BuildResourcePath(a.GrafanaConf, treeName, domain.RoutingTreeResource, a.IsLocal(), false)
		if writeErr := a.Storage.WriteFile(dsPath, dsPacked); writeErr != nil {
			return nil, fmt.Errorf("unable to write file. %w", writeErr)
		}
		files = append(files, dsPath)
	}
	return files, nil
}

// ClearRoutingTrees deletes matching routing trees. When the server rejects a
// delete (the org's protected/default tree typically cannot be removed), the
// tree is reset to domain.DefaultRoutingTreeSpec instead of being skipped.
// If name is non-empty, only that tree is cleared.
func (a *serviceImpl) ClearRoutingTrees(name string) ([]string, error) {
	client, _, _, err := a.routingTreeClient()
	if err != nil {
		return nil, err
	}

	trees, err := a.ListRoutingTrees(name)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var cleared []string
	for _, tree := range trees {
		treeName := tree.Name()
		if treeName == "" {
			slog.Warn("skipping routing tree without a name")
			continue
		}

		if err := a.deleteOrResetRoutingTree(ctx, client, treeName); err != nil {
			slog.Error("unable to delete or reset routing tree", "name", treeName, "err", err)
			continue
		}
		cleared = append(cleared, treeName)
	}
	return cleared, nil
}

// resetRoutingTree replaces a routing tree's spec with domain.DefaultRoutingTreeSpec,
// used when the server refuses to delete it outright (e.g. the protected default tree).
func (a *serviceImpl) resetRoutingTree(ctx context.Context, client dynamic.ResourceInterface, name string) error {
	existing, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to fetch routing tree %q for reset: %w", name, err)
	}
	existing.Object["spec"] = domain.DefaultRoutingTreeSpec()
	if _, err := client.Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("unable to reset routing tree %q: %w", name, err)
	}
	return nil
}

// UploadRoutingTrees makes the server fully reflect local storage: it
// creates/updates every matching local tree, then removes any matching
// server-side tree that has no local file (using the same delete-or-reset
// logic as ClearRoutingTrees for trees the server refuses to delete), since
// GDG is the source of truth for what should exist. If name is non-empty,
// only that tree is reconciled — both the upload and the prune step are
// scoped to it.
func (a *serviceImpl) UploadRoutingTrees(name string) ([]string, error) {
	client, ns, apiVersion, err := a.routingTreeClient()
	if err != nil {
		return nil, err
	}

	treesPath := a.GrafanaConf.GetPath(domain.RoutingTreeResource, a.GrafanaConf.GetOrganizationName())
	filesInDir, err := a.Storage.FindAllFiles(treesPath, true)
	if err != nil {
		return nil, fmt.Errorf("unable to find any routing trees to upload from storage engine, err: %w", err)
	}

	ctx := context.Background()
	local := make(map[string]bool)
	var uploaded []string
	for _, file := range filesInDir {
		if !strings.HasSuffix(file, ".json") {
			slog.Debug("Only json files are supported, skipping", "filename", file)
			continue
		}
		rawEntity, readErr := a.Storage.ReadFile(file)
		if readErr != nil {
			slog.Warn("Unable to read file", "filename", file, "err", readErr)
			continue
		}

		var gdg domain.RoutingTreeGdg
		if umErr := json.Unmarshal(rawEntity, &gdg); umErr != nil {
			slog.Warn("Failed to unmarshal routing tree resource", "filename", file, "err", umErr)
			continue
		}
		treeName := gdg.Name()
		if treeName == "" {
			slog.Warn("skipping routing tree file without a name", "filename", file)
			continue
		}
		if name != "" && treeName != name {
			continue
		}
		local[treeName] = true

		obj := &unstructured.Unstructured{Object: gdg.Resource}
		obj.SetNamespace(ns)
		obj.SetAPIVersion(apiVersion)
		obj.SetKind(domain.RoutingTreeKind)
		obj.SetResourceVersion("")

		if upErr := upsertRoutingTree(ctx, client, treeName, obj); upErr != nil {
			slog.Error("unable to upload routing tree", "name", treeName, "file", file, "err", upErr)
			continue
		}
		uploaded = append(uploaded, treeName)
	}

	serverTrees, err := a.ListRoutingTrees(name)
	if err != nil {
		return uploaded, fmt.Errorf("unable to list routing trees to prune orphaned entries: %w", err)
	}
	for _, tree := range serverTrees {
		treeName := tree.Name()
		if treeName == "" || local[treeName] {
			continue
		}
		if err := a.deleteOrResetRoutingTree(ctx, client, treeName); err != nil {
			slog.Error("unable to remove routing tree with no local file", "name", treeName, "err", err)
			continue
		}
	}

	return uploaded, nil
}

// upsertRoutingTree creates the routing tree resource if it does not already
// exist, or updates it (carrying over the current resourceVersion) if it does.
func upsertRoutingTree(ctx context.Context, client dynamic.ResourceInterface, name string, obj *unstructured.Unstructured) error {
	existing, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("unable to check for existing routing tree %q: %w", name, err)
		}
		_, err = client.Create(ctx, obj, metav1.CreateOptions{})
		return err
	}

	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = client.Update(ctx, obj, metav1.UpdateOptions{})
	return err
}
