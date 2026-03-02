package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"regexp"
	"strings"

	"github.com/esnet/gdg/internal/adapter/filters/v2"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/ptr"
	"github.com/gosimple/slug"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"

	"github.com/grafana/grafana-openapi-client-go/client/provisioning"
)

func setupAlertRulesReaders(filterObj ports.Filter, grafanaSvc ports.GrafanaService) {
	// Object Reader
	err := filterObj.RegisterReader(reflect.TypeFor[*domain.AlertRuleWithNestedFolder](), func(filterType domain.FilterType, a any) (any, error) {
		val, ok := a.(*domain.AlertRuleWithNestedFolder)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.Name:
			return ptr.ValueOrDefault(val.Title, ""), nil
		case domain.FolderFilter:
			return val.NestedPath, nil
		case domain.TagsFilter:
			return val.Labels, nil
		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("unable to register a valid Alert Rule for alert rules filter. %v", err)
	}

	// Raw Reader
	err = filterObj.RegisterReader(reflect.TypeFor[[]byte](), func(filterType domain.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.Name:
			{
				r := gjson.GetBytes(val, "title")
				if !r.Exists() || r.IsArray() {
					return nil, errors.New("no valid rule name was found")
				}
				return r.String(), nil
			}
		case domain.FolderFilter:
			{
				r := gjson.GetBytes(val, "folderUID")
				if !r.Exists() || r.IsArray() {
					return DefaultFolderName, nil
				}
				folderUid := r.String()
				folderObj, folderErr := grafanaSvc.(*DashNGoImpl).getFolderByUid(folderUid)
				if folderErr != nil {
					return nil, folderErr
				}
				return folderObj.NestedPath, nil
			}
		case domain.TagsFilter:
			{
				slog.Info("Trying to read labels filter")
				r := gjson.GetBytes(val, "labels")
				data := make(map[string]string)
				if !r.Exists() || r.IsArray() {
					slog.Debug("unable to read rules labels")
					return data, nil
				}
				ar := r.Map()
				for k, v := range ar {
					data[k] = v.String()
				}
				return data, nil
			}

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	if err != nil {
		log.Fatalf("unable to register a valid byte reader for alert rules filter")
	}
}

// NewAlertRuleFilter creates and configures a new Filter for alert rules. It registers readers and validations
// for both typed alert rule objects and raw byte data, supporting folder-based and label-based filtering.
// Folder filtering uses regex matching against monitored folders from configuration or an explicit folder override.
// Label filtering expects labels in "key=value" format. The function terminates the process if reader registration fails.
func NewAlertRuleFilter(cfg *configDomain.GDGAppConfiguration, grafanaSvc ports.GrafanaService, filterEntities domain.AlertRuleFilterParams) ports.Filter {
	filterObj := v2.NewBaseFilter()
	// Define how we read data
	setupAlertRulesReaders(filterObj, grafanaSvc)
	err := filterObj.RegisterDataProcessor(domain.FolderFilter, domain.ProcessorEntity{
		Name: "folderQuoteRegEx",
		Processor: func(item any) (any, error) {
			switch w := item.(type) {
			case string:
				slog.Debug("folder quote filter applied to string")
				quoteRegex, _ := regexp.Compile("['\"]+")
				w = quoteRegex.ReplaceAllString(w, "")
				return w, nil
			case []string:
				slog.Debug("folder quote filter applied to []string")
				return lo.Map(w, func(i string, index int) string {
					quoteRegex, _ := regexp.Compile("['\"]+")
					i = quoteRegex.ReplaceAllString(i, "")
					return i
				}), nil
			}
			return item, nil
		},
	})
	if err != nil {
		log.Fatalf("Unable to create a valid Dashboard Filter, aborting.")
	}

	// Define the rules we enforce
	// Folder Behavior
	var folderArr []string
	if filterEntities.IgnoreWatchedFolders {
		if filterEntities.Folder != "" {
			folderArr = []string{filterEntities.Folder}
		} else {
			folderArr = []string{".*"}
		}
	} else {
		folderArr = cfg.GetDefaultGrafanaConfig().GetMonitoredFolders(false)
		if filterEntities.Folder != "" {
			folderArr = []string{filterEntities.Folder}
		}
	}

	filterObj.AddValidation(domain.FolderFilter, func(value any, expected any) error {
		//if filterEntities.IgnoreWatchedFolders && filterEntities.Folder == "" {
		//	return nil
		//}
		val, expressions, convErr := v2.GetMismatchParams[string, []string](value, expected, domain.FolderFilter)
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

	filterObj.AddValidation(domain.Name, func(value any, expected any) error {
		val, expressions, convErr := v2.GetParams[string](value, expected, domain.Name)
		if convErr != nil {
			return convErr
		}
		//no filter active
		if expressions == "" {
			return nil
		}
		if expected == val {
			return nil
		}

		return fmt.Errorf("invalid folder filter. Expected: %v", expressions)
	}, filterEntities.Name)

	// Tags
	filterObj.AddValidation(domain.TagsFilter, func(value any, expected any) error {
		val, exp, convErr := v2.GetMismatchParams[map[string]string, []string](value, expected, domain.TagsFilter)

		if convErr != nil {
			return convErr
		}
		// no filter active, returning nil
		if len(exp) == 0 {
			return nil
		}
		for _, labelFilterVal := range exp {
			elements := strings.Split(labelFilterVal, "=")
			if len(elements) != 2 {
				return fmt.Errorf("invalid label format, key=value expected")
			}
			entryKey := strings.TrimSpace(elements[0])
			entryValue := strings.TrimSpace(elements[1])

			if val[entryKey] == entryValue {
				return nil
			}
		}

		return fmt.Errorf("failed validation test val:%s  expected: %s", val, exp)
	}, filterEntities.Label)

	return filterObj
}

func (s *DashNGoImpl) ListAlertRules(filter ports.Filter) ([]*domain.AlertRuleWithNestedFolder, error) {
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
		if filter == nil || filter.ValidateAll(entry) {
			results = append(results, entry)
		}
	}

	return results, nil
}

func (s *DashNGoImpl) UploadAlertRules(filter ports.Filter) error {
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
	currentRules, err := s.ListAlertRules(filter)
	if err != nil {
		return err
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
		if filter != nil && !filter.ValidateAll(rawEntity) {
			slog.Debug("Skipping file, failed alert rule filter", "file", file)
			continue
		}
		entity := new(domain.AlertRuleWithNestedFolder)
		if err = json.Unmarshal(rawEntity, &entity); err != nil {
			return fmt.Errorf("failed to unmarshall file, file:%s, err: %w", file, err)
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
		}
	}

	return nil
}

func (s *DashNGoImpl) DownloadAlertRules(filter ports.Filter) ([]string, error) {
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
		base := BuildResourceFolder(s.grafanaConf, link.NestedPath, domain.AlertingRulesResource, s.isLocal(), s.GetGlobals().ClearOutput)
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

func (s *DashNGoImpl) ClearAlertRules(filter ports.Filter) ([]string, error) {
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
