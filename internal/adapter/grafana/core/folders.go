package core

import (
	"context"
	"fmt"
	"log"
	"path"
	"regexp"

	v2 "github.com/esnet/gdg/internal/adapter/filters/v2"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/encode"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
)

// NewFolderFilter builds a Filter that matches folders against the configured
// monitored-folders list.
func NewFolderFilter(cfg *configDomain.GDGAppConfiguration) outbound.Filter {
	filterObj := v2.NewBaseFilter()
	err := v2.RegisterTypedReader[*domain.NestedHit](filterObj, func(ctx context.Context, filterType domain.FilterType, val *domain.NestedHit) (any, error) {
		switch filterType {
		case domain.FolderFilter:
			return val.NestedPath, nil
		default:
			return nil, fmt.Errorf("unsupported filter type: %s", filterType)
		}
	})
	if err != nil {
		log.Fatalf("unable to register a valid reader for folder filter")
	}

	folderArr := cfg.GetDefaultGrafanaConfig().GetMonitoredFolders(false)
	filterObj.AddValidation(domain.FolderFilter, func(ctx context.Context, value any, expected any) error {
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
	return filterObj
}

// GetFolderMapping returns a mapping of any comparable T to any value based on the folder entity.
// key is a function that takes the folder type as a parameter and returns the key to use for the resulting map.
// val is a function that takes the folder type as a parameter and returns the value to set the map value to.
func GetFolderMapping[T comparable, V any](folders []*domain.NestedHit, key func(fld *domain.NestedHit) T, val func(fld *domain.NestedHit) V) map[T]V {
	m := make(map[T]V)
	for _, f := range folders {
		m[key(f)] = val(f)
	}
	return m
}

// GetFolderUIDEntityMapByList builds a mapping from folder UID to NestedHit.
func GetFolderUIDEntityMapByList(folders []*domain.NestedHit) map[string]*domain.NestedHit {
	return GetFolderMapping(folders, func(fld *domain.NestedHit) string {
		return fld.UID
	},
		func(fld *domain.NestedHit) *domain.NestedHit {
			return fld
		},
	)
}

// GetNestedFolder returns the nested folder path for a given folder, walking up
// the parent chain via folderUidMap.
func GetNestedFolder(folderTitle, folderUID string, folderUidMap map[string]*domain.NestedHit) string {
	folderPath := encode.Encode(folderTitle)
	currentFolderUid := folderUID
	for currentFolderUid != "" {
		parent, ok := folderUidMap[currentFolderUid]
		if ok && parent.FolderUID != "" {
			currentFolderUid = parent.FolderUID
			folderPath = fmt.Sprintf("%s/%s", encode.Encode(parent.FolderTitle), folderPath)
		} else {
			currentFolderUid = ""
		}
	}
	return folderPath
}

// GetFolderByUid gets a given folder given a valid Uid.
func (b *Service) GetFolderByUid(uid string) (*domain.NestedHit, error) {
	res, err := b.GetClient().Folders.GetFolderByUID(uid)
	if err != nil {
		return nil, err
	}
	return b.FolderToHit(res.GetPayload()), nil
}

// GetFolderByNestedPath retrieves a single folder matching the given nested path,
// using the provided filter to obtain the list of folders. Returns an error if no
// folder is found or if multiple folders match.
func (b *Service) GetFolderByNestedPath(nestedPath string, filter outbound.Filter) (*domain.NestedHit, error) {
	folderList := b.ListFolders(filter)
	result := lo.Filter(folderList, func(item *domain.NestedHit, index int) bool {
		return item.NestedPath == nestedPath
	})
	if len(result) == 0 {
		return nil, fmt.Errorf("unable to find folder with nested path %s", nestedPath)
	}
	if len(result) > 1 {
		return nil, fmt.Errorf("unexpected behavior, too many folders matched for %s", nestedPath)
	}

	return result[0], nil
}

// FolderToHit converts a models.Folder struct to a models.Hit struct.
func (b *Service) FolderToHit(fld *models.Folder) *domain.NestedHit {
	res := new(domain.NestedHit)
	res.Hit = new(models.Hit)
	res.Title = fld.Title
	res.UID = fld.UID
	res.FolderUID = fld.ParentUID
	res.Type = models.HitType(domain.ApiConsts.SearchTypeFolder)
	res.URL = fld.URL
	paths := lo.Map(fld.Parents, func(item *models.Folder, index int) string {
		return encode.Encode(item.Title)
	})
	if val := path.Join(paths...); val == "" {
		res.NestedPath = encode.Encode(res.Title)
	} else {
		res.NestedPath = path.Join(val, encode.Encode(res.Title))
	}

	return res
}
