package service

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"regexp"
	"strings"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/domain"
	modelsDomain "github.com/esnet/gdg/internal/service/domain"
	"github.com/esnet/gdg/internal/service/filters"
	v2 "github.com/esnet/gdg/internal/service/filters/v2"
	"github.com/gosimple/slug"

	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
)

func NewAlertRuleFilter() filters.V2Filter {
	filterObj := v2.NewBaseFilter()
	err := filterObj.RegisterReader(reflect.TypeOf(&modelsDomain.AlertRuleWithNestedFolder{}), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.(*modelsDomain.AlertRuleWithNestedFolder)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.AlertRuleFilterType:
			return val.NestedPath, nil
		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("unable to register a valid reader for folder filter")
	}

	folderArr := config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders(false)
	filterObj.AddValidation(filters.AlertRuleFilterType, func(value any, expected any) error {
		val, expressions, convErr := v2.GetMismatchParams[string, []string](value, expected, filters.AlertRuleFilterType)
		if convErr != nil {
			return convErr
		}
		for _, exp := range expressions {
			r, ReErr := regexp.Compile(exp)
			if ReErr != nil {
				return fmt.Errorf("invalid regex: %s", exp)
			}
			if r.MatchString(val) {
				return nil
			}
		}

		return fmt.Errorf("invalid folder filter. Expected: %v", expressions)
	}, folderArr)
	return filterObj
}

func (s *DashNGoImpl) ListAlertRules(filter filters.V2Filter) ([]*modelsDomain.AlertRuleWithNestedFolder, error) {
	data, err := s.GetClient().Provisioning.GetAlertRules()
	if err != nil {
		return nil, err
	}

	folderUidMap := s.getFolderUIDEntityMap(nil)
	var results []*modelsDomain.AlertRuleWithNestedFolder

	for _, item := range data.GetPayload() {
		entry := &modelsDomain.AlertRuleWithNestedFolder{
			ProvisionedAlertRule: item,
		}

		if folder, ok := folderUidMap[ptr.ValueOrDefault(item.FolderUID, "")]; ok {
			entry.NestedPath = folder.NestedPath
		}

		if filter == nil || filter.Validate(filters.AlertRuleFilterType, entry) {
			results = append(results, entry)
		}

	}

	return results, nil
}

func (s *DashNGoImpl) UploadAlertRules(filter filters.V2Filter) error {
	// TODO: once filtering in enabled we should delete any rules that we're not tracking in the folders that gdg manages
	var (
		err       error
		rawEntity []byte
	)

	rulesPath := s.grafanaConf.GetPath(domain.AlertingRulesResource, s.grafanaConf.GetOrganizationName())
	filesInDir, err := s.storage.FindAllFiles(rulesPath, true)
	if err != nil {
		return fmt.Errorf("unable to find any rules to export from storage engine, err: %w", err)
	}
	currentContacts, err := s.ListAlertRules(filter)
	if err != nil {
		return err
	}
	m := lo.Associate(currentContacts, func(item *modelsDomain.AlertRuleWithNestedFolder) (string, *modelsDomain.AlertRuleWithNestedFolder) {
		return item.UID, item
	})

	for _, file := range filesInDir {
		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json files are supported, skipping", "filename", file)
			continue
		}
		if rawEntity, err = s.storage.ReadFile(file); err != nil {
			slog.Warn("Unable to read file", "filename", file, "err", err)
			continue
		}
		entity := new(modelsDomain.AlertRuleWithNestedFolder)
		if err = json.Unmarshal(rawEntity, &entity); err != nil {
			return fmt.Errorf("failed to unmarshall file, file:%s, err: %w", file, err)
		}

		if _, ok := m[entity.UID]; ok {
			p := provisioning.NewPutAlertRuleParams()
			p.Body = entity.ProvisionedAlertRule
			p.UID = entity.UID
			p.XDisableProvenance = ptr.Of("true")
			_, err = s.GetClient().Provisioning.PutAlertRule(p)
		} else {
			p := provisioning.NewPostAlertRuleParams()
			p.Body = entity.ProvisionedAlertRule
			p.XDisableProvenance = ptr.Of("true")
			_, err = s.GetClient().Provisioning.PostAlertRule(p)
		}
		if err != nil {
			slog.Error("unable to import rule", "uid", entity.UID, "err", err)
		}
	}

	return nil
}

func (s *DashNGoImpl) DownloadAlertRules(filter filters.V2Filter) ([]string, error) {
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
		base := BuildResourceFolder(link.NestedPath, domain.AlertingRulesResource, s.isLocal(), s.globalConf.ClearOutput)
		fileName := fmt.Sprintf("%s/%s.json", base, slug.Make(ptr.ValueOrDefault(link.Title, "no-name")))
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

func (s *DashNGoImpl) ClearAlertRules(filter filters.V2Filter) ([]string, error) {
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
		data = append(data, ptr.ValueOrDefault(rule.Title, ""))
	}

	return data, nil
}
