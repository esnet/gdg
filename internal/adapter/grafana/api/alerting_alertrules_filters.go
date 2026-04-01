package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/jellydator/ttlcache/v3"

	v2 "github.com/esnet/gdg/internal/adapter/filters/v2"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/pkg/ptr"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
)

const noAlertRulesData string = ""

var alertRuleCache = ttlcache.New[string, *domain.NestedHit](
	ttlcache.WithTTL[string, *domain.NestedHit](10*time.Minute),
	ttlcache.WithDisableTouchOnHit[string, *domain.NestedHit](),
)

// setupAlertRulesReaders registers filter readers for alert rules on the provided filter object. It registers
// two readers: one for typed *domain.AlertRuleWithNestedFolder objects and one for raw []byte JSON data.
// Each reader supports extracting values for UID, FolderFilter, and TagsFilter filter types. The typed
// reader extracts fields directly from the struct, while the raw byte reader parses JSON using gjson and
// resolves folder UIDs to nested paths via the Grafana service. The function terminates the process via
// log.Fatalf if either reader registration fails.
func setupAlertRulesReaders(filterObj outbound.Filter, grafanaSvc outbound.GrafanaService) {
	// Object Reader
	err := filterObj.RegisterReader(reflect.TypeFor[*domain.AlertRuleWithNestedFolder](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
		val, ok := a.(*domain.AlertRuleWithNestedFolder)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.UID:
			return val.UID, nil
		case domain.FolderFilter:
			// return val.NestedPath, nil
			return ptr.ValueOrDefault(val.FolderUID, ""), nil
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
	err = filterObj.RegisterReader(reflect.TypeFor[[]byte](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.UID:
			{
				r := gjson.GetBytes(val, "uid")
				if !r.Exists() || r.IsArray() {
					return nil, errors.New("no valid rule name was found")
				}
				return r.String(), nil
			}
		case domain.FolderFilter:
			{
				r := gjson.GetBytes(val, "folderUID")
				if !r.Exists() || r.IsArray() {
					return domain.ApiConsts.DefaultFolderName, nil
				}

				return r.String(), nil
			}
		case domain.TagsFilter:
			{
				slog.Debug("Trying to read labels filter")
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

// setupAlertRulesFolderFilter configures a folder-based filter for alert rules by registering a data processor
// that strips quote characters from folder names and adding a regex-based validation rule. The folder list is
// determined by the filter parameters: if IgnoreWatchedFolders is set, it uses the explicitly provided folder
// or a wildcard match-all pattern; otherwise, it defaults to the monitored folders from the Grafana configuration,
// optionally overridden by a specific folder in filterEntities. The function terminates the process if the
// filter registration fails.
func setupAlertRulesFolderFilter(filterObj outbound.Filter, filterEntities domain.AlertRuleFilterParams, cfg *configDomain.GDGAppConfiguration, grafanaSvc outbound.GrafanaService) {
	const folderSeparator = "|"
	err := filterObj.RegisterDataProcessor(domain.FolderFilter, domain.ProcessorEntity{
		Name: "folderQuoteRegEx",
		Processor: func(ctx context.Context, item any) (any, error) {
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

	filterObj.AddValidation(domain.FolderFilter, func(ctx context.Context, value any, expected any) error {
		val, expressions, convErr := v2.GetMismatchParams[string, []string](value, expected, domain.FolderFilter)
		if convErr != nil {
			return convErr
		}

		var folders []string
		folderName, folderLookupErr := alertRuleFilterLookUpFolderName(ctx, val, grafanaSvc)
		if folderLookupErr != nil {
			return folderLookupErr
		}
		if strings.Contains(folderName, folderSeparator) {
			folders = strings.Split(folderName, folderSeparator)
		} else {
			folders = append(folders, folderName)
		}
		for _, exp := range expressions {
			r, ReErr := regexp.Compile(exp)
			if ReErr != nil {
				return fmt.Errorf("invalid regex: %s", exp)
			}

			for _, folder := range folders {
				if r.MatchString(folder) {
					return nil
				}
			}
		}

		return fmt.Errorf("invalid folder filter. Expected: %v", expressions)
	}, folderArr)
}

// alertRuleFilterLookUpFolderName resolves a folder name for an alert rule filter by looking up the folder
// associated with the given folderUid. It retrieves the file name from the context to use as a fallback or
// validation reference. The folder is fetched from the Grafana API or from an in-memory TTL cache. If the
// folder's nested path does not match the file-based location (case-insensitive), both values are returned
// joined by a pipe delimiter to allow validation against either path. Returns an error if the file name
// cannot be extracted from the context.
func alertRuleFilterLookUpFolderName(ctx context.Context, folderUid string, grafanaSvc outbound.GrafanaService) (string, error) {
	var err error
	folderNameLocation, ok := ctx.Value(fileContextKey).(string)
	if !ok {
		return noAlertRulesData, errors.New("unable to get file name from context")
	}

	// Get folder from API or cached copy.
	var folderObj *domain.NestedHit
	cacheEntry := alertRuleCache.Get(folderUid)
	if cacheEntry == nil {
		folderObj, err = grafanaSvc.(*DashNGoImpl).getFolderByUid(folderUid)
		if err != nil {
			slog.Debug("unable to find folder by uid, using file location for validation", "nested path", folderNameLocation)
			return folderNameLocation, nil
		}
		alertRuleCache.Set(folderUid, folderObj, ttlcache.DefaultTTL)
	} else {
		folderObj = cacheEntry.Value()
	}

	if !strings.EqualFold(folderObj.NestedPath, folderNameLocation) {
		slog.Warn("folder path for UID location does not match folder look up, using both for validation", slog.String("folder", folderNameLocation), slog.String("folder", folderObj.NestedPath))
		return fmt.Sprintf("%s|%s", folderNameLocation, folderObj.NestedPath), nil
	}

	return folderNameLocation, nil
}

// NewAlertRuleFilter creates and configures a new Filter for alert rules. It registers readers and validations
// for both typed alert rule objects and raw byte data, supporting folder-based and label-based filtering.
// Folder filtering uses regex matching against monitored folders from configuration or an explicit folder override.
// Label filtering expects labels in "key=value" format. The function terminates the process if reader registration fails.
func NewAlertRuleFilter(cfg *configDomain.GDGAppConfiguration, grafanaSvc outbound.GrafanaService, filterEntities domain.AlertRuleFilterParams) outbound.Filter {
	filterObj := v2.NewBaseFilter()
	// Define how we read data
	setupAlertRulesReaders(filterObj, grafanaSvc)
	setupAlertRulesFolderFilter(filterObj, filterEntities, cfg, grafanaSvc)

	// AlertName filter
	filterObj.AddValidation(domain.UID, func(ctx context.Context, value any, expected any) error {
		val, expressions, convErr := v2.GetParams[string](value, expected, domain.UID)
		if convErr != nil {
			return convErr
		}
		// no filter active
		if expressions == "" {
			return nil
		}
		if expected == val {
			return nil
		}

		return fmt.Errorf("invalid folder filter. Expected: %v", expressions)
	}, filterEntities.UID)

	// Alert Label filter
	filterObj.AddValidation(domain.TagsFilter, func(ctx context.Context, value any, expected any) error {
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
