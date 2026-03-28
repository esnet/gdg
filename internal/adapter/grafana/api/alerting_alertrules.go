package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/ptr"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
	"github.com/samber/lo"
	"github.com/tidwall/sjson"
)

type contextKey string

const (
	fileContextKey contextKey = "file"
)

func (s *DashNGoImpl) ListAlertRules(filter outbound.Filter) ([]*domain.AlertRuleWithNestedFolder, error) {
	data, err := s.GetClient().Provisioning.GetAlertRules()
	if err != nil {
		return nil, err
	}

	folderUidMap := s.getFolderUIDEntityMap(nil)
	var results []*domain.AlertRuleWithNestedFolder

	for _, item := range data.GetPayload() {
		entry := &domain.AlertRuleWithNestedFolder{
			ProvisionedAlertRule: item,
		}

		if folder, ok := folderUidMap[ptr.ValueOrDefault(item.FolderUID, "")]; ok {
			entry.NestedPath = folder.NestedPath
		}
		file := getDestinationFilePath(s.grafanaConf, entry, s.isLocal(), s.resources)
		folderNameLocation, err := s.resources.GetFolderFromResourcePath(s.grafanaConf, file, domain.AlertingRulesResource, s.storage.GetPrefix(), s.grafanaConf.GetOrganizationName())
		if err != nil {
			slog.Error(fmt.Sprintf("unable to determine alert rule folder name, falling back on default, %v", err))
			continue
		}
		ctx := getFilterContext(folderNameLocation)
		if filter == nil || filter.ValidateAll(ctx, entry) {
			results = append(results, entry)
		}
	}

	return results, nil
}

// getDestinationFilePath constructs the full file path for an alert rule JSON file based on the Grafana configuration,
// the alert rule's nested folder path, and whether the output is for local storage or should be cleared.
func getDestinationFilePath(grafanaConf *config_domain.GrafanaConfig, entry *domain.AlertRuleWithNestedFolder, local bool, s ports.Resources) string {
	base := s.BuildResourceFolder(grafanaConf, entry.NestedPath, domain.AlertingRulesResource, local, false)
	file := fmt.Sprintf("%s/%s.json", base, slug.Make(ptr.ValueOrDefault(entry.Title, "no-name")))

	return file
}

// getFilterContext creates and returns a context populated with the given file path, Grafana configuration, and storage.
func getFilterContext(file string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, fileContextKey, file)

	return ctx
}

func (s *DashNGoImpl) UploadAlertRules(filter outbound.Filter) ([]*domain.AlertRuleWithNestedFolder, error) {
	// TODO: once filtering in enabled we should delete any rules that we're not tracking in the folders that gdg manages
	var (
		success   []*domain.AlertRuleWithNestedFolder
		err       error
		rawEntity []byte
	)

	rulesPath := s.grafanaConf.GetPath(domain.AlertingRulesResource, s.grafanaConf.GetOrganizationName())
	filesInDir, err := s.storage.FindAllFiles(rulesPath, true)
	if err != nil {
		return nil, fmt.Errorf("unable to find any rules to export from storage engine, err: %w", err)
	}
	currentRules, err := s.ListAlertRules(filter)
	if err != nil {
		return nil, err
	}
	m := lo.Associate(currentRules, func(item *domain.AlertRuleWithNestedFolder) (string, *domain.AlertRuleWithNestedFolder) {
		return item.UID, item
	})

	for _, file := range filesInDir {
		if !strings.HasSuffix(file, ".json") {
			slog.Debug("Only json files are supported, skipping", "filename", file)
			continue
		}
		if rawEntity, err = s.storage.ReadFile(file); err != nil {
			slog.Warn("Unable to read file", "filename", file, "err", err)
			continue
		}

		folderNameLocation, err := s.resources.GetFolderFromResourcePath(s.grafanaConf, file, domain.AlertingRulesResource, s.storage.GetPrefix(), s.grafanaConf.GetOrganizationName())
		if err != nil {
			slog.Debug("unable to determine alert rule folder name", "err", err)
		}
		ctx := getFilterContext(folderNameLocation)

		if filter != nil && !filter.ValidateAll(ctx, rawEntity) {
			slog.Debug("Skipping file, failed alert rule filter", "file", file)
			continue
		}
		folderUidRaw := filter.GetReaderValue(ctx, domain.FolderFilter, rawEntity)
		var folderUid string
		var ok bool
		if folderUid, ok = folderUidRaw.(string); !ok {
			slog.Error("Could not retrieve folder ID, skipping", "file", file)
			continue
		}

		rawEntity, err = s.alertRulesFolderResolver(ctx, folderUid, rawEntity, file)
		if err != nil {
			slog.Error("failed sanity check for alert rule folder. Skipping", "file", file)
			continue
		}
		entity := new(domain.AlertRuleWithNestedFolder)
		if err = json.Unmarshal(rawEntity, &entity); err != nil {
			slog.Error("failed to unmarshall file.", "file", file, "err", err)
			continue
		}

		if _, ok := m[entity.UID]; ok {
			p := provisioning.NewPutAlertRuleParams()
			p.Body = entity.ProvisionedAlertRule
			p.UID = entity.UID

			p.XDisableProvenance = new("true")
			_, err = s.GetClient().Provisioning.PutAlertRule(p)
		} else {
			p := provisioning.NewPostAlertRuleParams()
			p.Body = entity.ProvisionedAlertRule
			p.XDisableProvenance = new("true")
			_, err = s.GetClient().Provisioning.PostAlertRule(p)
		}
		if err != nil {
			slog.Error("unable to import rule", "uid", entity.UID, "err", err)
		} else {
			success = append(success, entity)
		}
	}

	return success, nil
}

// Create a Folder Resolver to be invoked by AlertRules
func (s *DashNGoImpl) alertRulesFolderResolver(ctx context.Context, folderUid string, rawEntity []byte, file string) ([]byte, error) {
	folderNameLocation, fileLocationErr := s.resources.GetFolderFromResourcePath(s.grafanaConf, file, domain.AlertingRulesResource, s.storage.GetPrefix(), s.grafanaConf.GetOrganizationName())
	if fileLocationErr != nil {
		fmt.Printf("unable to determine alert rule folder name, %v", fileLocationErr)
	}
	folderObj, folderErr := s.getFolderByUid(folderUid)
	if folderErr != nil && fileLocationErr != nil {
		return nil, fmt.Errorf("unable to proceed, folder does not exist and unable to look up name from file path. %v, %v", folderErr, fileLocationErr)
	}
	folderByPath, folderPathLookupErr := s.getFolderByNestedPath(folderNameLocation, nil)
	// Folder exists but with a different UID
	if folderErr != nil && folderPathLookupErr == nil {
		return sjson.SetBytes(rawEntity, "folderUID", folderByPath.UID)
	}
	if folderErr != nil {
		_, err := s.createdFoldersWithBaseUID(folderNameLocation, folderUid)
		if err != nil {
			return nil, fmt.Errorf("unable to proceed, could not create missing folder with UID: %s and path: %s. folderPathLookupErr: %v", folderUid, folderNameLocation, err)
		}
		return rawEntity, nil
	}
	// At this point we should have a valid folder lookup and file path lookup.
	if !strings.EqualFold(folderObj.NestedPath, folderNameLocation) {
		return nil, fmt.Errorf("invalid state, folder exists but does not have the expected path. %s, %s", folderObj.NestedPath, folderNameLocation)
	}

	// everything looks good we can proceed
	return rawEntity, nil
}

func (s *DashNGoImpl) DownloadAlertRules(filter outbound.Filter) ([]string, error) {
	var (
		dsPacked []byte
		err      error
	)
	data, err := s.ListAlertRules(filter)
	if err != nil {
		return nil, err
	}
	var savedFiles []string
	for _, link := range data {
		fileName := getDestinationFilePath(s.grafanaConf, link, s.isLocal(), s.resources)
		if dsPacked, err = json.MarshalIndent(link, "", "	"); err != nil {
			return nil, fmt.Errorf("unable to serialize data to JSON. %w", err)
		}
		if err = s.storage.WriteFile(fileName, dsPacked); err != nil {
			return nil, fmt.Errorf("unable to write file. %w", err)
		}
		savedFiles = append(savedFiles, fileName)
	}

	return savedFiles, nil
}

func (s *DashNGoImpl) ClearAlertRules(filter outbound.Filter) ([]string, error) {
	rules, err := s.ListAlertRules(filter)
	if err != nil {
		return nil, err
	}
	var data []string
	for _, rule := range rules {
		p := provisioning.NewDeleteAlertRuleParams()
		p.UID = rule.UID
		_, err := s.GetClient().Provisioning.DeleteAlertRule(p)
		if err != nil {
			slog.Error("unable to delete rule", "rule", rule.UID)
			continue
		}
		data = append(data, fmt.Sprintf("%s/%s", rule.NestedPath, ptr.ValueOrDefault(rule.Title, "")))
	}

	return data, nil
}
