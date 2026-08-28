package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/esnet/gdg/internal/adapter/grafana/extended"
	"github.com/esnet/gdg/internal/adapter/storage"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/encode"
	"github.com/go-openapi/strfmt"
	"github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
)

// baseService holds the shared infrastructure that every service implementation
// (DashboardServiceImpl, and future FolderServiceImpl, ConnectionServiceImpl, etc.)
// needs. DashNGoImpl also embeds baseService so all existing methods continue to
// compile unchanged.
type baseService struct {
	extended    outbound.ExtendedApi
	gdgConfig   *configDomain.GDGAppConfiguration
	grafanaConf *configDomain.GrafanaConfig
	storage     outbound.Storage
	encoder     outbound.CipherEncoder
	resources   ports.Resources
}

// ---------------------------------------------------------------------------
// HTTP client helpers
// ---------------------------------------------------------------------------

func ignoreSSLBase(transportConfig *client.TransportConfig) {
	_, clientTransport := ignoreSSLErrorsBase()
	transportConfig.TLSConfig = clientTransport.TLSClientConfig
}

func ignoreSSLErrorsBase() (*http.Client, *http.Transport) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
	httpclient := &http.Client{Transport: customTransport}
	return httpclient, customTransport
}

func (b *baseService) getOrgNameClientOpts() NewClientOpts {
	cfg := b.gdgConfig
	orgName := cfg.GetDefaultGrafanaConfig().OrganizationName
	if orgName != "" {
		return func(transportConfig *client.TransportConfig) {
			orgId, err := extended.NewExtendedApi(cfg).GetConfiguredOrgId(orgName)
			if err != nil {
				slog.Error("unable to determine org ID, falling back", slog.Any("err", err))
				orgId = 1
			}
			transportConfig.OrgID = orgId
		}
	}
	return func(clientCfg *client.TransportConfig) {
		clientCfg.OrgID = configDomain.DefaultOrganizationId
	}
}

func (b *baseService) getNewClient(opts ...NewClientOpts) (*client.GrafanaHTTPAPI, *client.TransportConfig) {
	u, err := url.Parse(b.grafanaConf.GetURL())
	if err != nil {
		log.Fatal("invalid Grafana URL", b.grafanaConf.GetURL())
	}
	path, err := url.JoinPath(u.Path, "api")
	if err != nil {
		log.Fatal("invalid Grafana URL Path")
	}
	httpConfig := &client.TransportConfig{
		Host:         u.Host,
		BasePath:     path,
		Schemes:      []string{u.Scheme},
		NumRetries:   b.gdgConfig.GetAppGlobals().RetryCount,
		RetryTimeout: b.gdgConfig.GetAppGlobals().GetRetryTimeout(),
		Debug:        b.GetGlobals().ApiDebug,
	}
	if b.grafanaConf.IsBasicAuth() && len(opts) == 1 {
		opts = append(opts, b.getOrgNameClientOpts())
	}
	for _, opt := range opts {
		if opt != nil {
			opt(httpConfig)
		}
	}
	if b.gdgConfig.IgnoreSSL() {
		ignoreSSLBase(httpConfig)
	}
	return client.NewHTTPClientWithConfig(strfmt.Default, httpConfig), httpConfig
}

// GetClient returns a Grafana API client, preferring token auth over basic auth.
func (b *baseService) GetClient() *client.GrafanaHTTPAPI {
	if b.grafanaConf.GetAPIToken() != "" {
		grafanaClient, _ := b.getNewClient(func(clientCfg *client.TransportConfig) {
			clientCfg.APIKey = b.grafanaConf.GetAPIToken()
			clientCfg.Debug = b.GetGlobals().ApiDebug
		})
		return grafanaClient
	}
	return b.GetBasicAuthClient()
}

func (b *baseService) getDefaultBasicOpts() []NewClientOpts {
	return []NewClientOpts{func(clientCfg *client.TransportConfig) {
		clientCfg.BasicAuth = url.UserPassword(b.grafanaConf.UserName, b.grafanaConf.GetPassword())
		clientCfg.Debug = b.GetGlobals().ApiDebug
	}}
}

// GetBasicClientWithOpts returns a basic-auth client with additional options applied.
func (b *baseService) GetBasicClientWithOpts(opts ...NewClientOpts) *client.GrafanaHTTPAPI {
	allOpts := b.getDefaultBasicOpts()
	allOpts = append(allOpts, opts...)
	grafanaClient, _ := b.getNewClient(allOpts...)
	return grafanaClient
}

// GetBasicAuthClient returns a basic-auth Grafana API client.
func (b *baseService) GetBasicAuthClient() *client.GrafanaHTTPAPI {
	return b.GetBasicClientWithOpts()
}

// GetAdminClient returns the admin client; fatal if admin is not configured.
func (b *baseService) GetAdminClient() *client.GrafanaHTTPAPI {
	if !b.grafanaConf.IsGrafanaAdmin() || b.grafanaConf.UserName == "" {
		log.Fatal("Unable to get Grafana Admin SecureData.")
	}
	return b.GetBasicClientWithOpts()
}

// ---------------------------------------------------------------------------
// Config / globals helpers
// ---------------------------------------------------------------------------

// GetGlobals returns the global app configuration, initialising it if nil.
func (b *baseService) GetGlobals() *configDomain.AppGlobals {
	if b.gdgConfig.Global == nil {
		b.gdgConfig.Global = &configDomain.AppGlobals{}
	}
	return b.gdgConfig.Global
}

// isLocal reports whether the configured storage backend is local disk.
func (b *baseService) isLocal() bool {
	return b.storage.Name() == storage.LocalStorageType.String()
}

// ---------------------------------------------------------------------------
// Server info
// ---------------------------------------------------------------------------

// GetServerInfo returns basic Grafana server health information.
// It satisfies tools.GetVersion so version checks can be called directly on any
// service that embeds baseService.
func (b *baseService) GetServerInfo() map[string]any {
	response, err := b.GetClient().Health.GetHealth()
	if err != nil {
		log.Fatalf("Unable to get server health info, err: %v", err)
	}
	t := response.GetPayload()
	result := make(map[string]any)
	result[SrvInfoDBKey] = t.Database
	result[SrvInfoCommitKey] = t.Commit
	result[SrvInfoVersionKey] = t.Version
	result[SrvInfoEnterpriseCommitKey] = t.EnterpriseCommit
	return result
}

// ---------------------------------------------------------------------------
// Folder helpers — used by DashboardServiceImpl and any future service
// ---------------------------------------------------------------------------

// searchAllPages drives a legacy /api/search call to exhaustion, looping on
// Limit/Page the same way listDashboardsV1 already does for dashboards.
// Grafana's search endpoint caps a single page at up to 5000 results and
// simply returns whatever fits -- it does not error or signal truncation,
// so a caller that sends one unpaginated request silently gets only the
// first page. That's exactly the bug class that hit dashboard v2 listing
// (a single List() call silently truncated to the server's page size): a
// single Search() call here would silently drop folders once an org has
// more of them than one page holds, which in turn corrupts folder-path
// resolution for every dashboard under a missing folder with no error or
// log line. Loop until a short page confirms there's nothing left.
func searchAllPages(apiClient *client.GrafanaHTTPAPI, configure func(*search.SearchParams)) []*models.Hit {
	const pageSize int64 = 5000 // Grafana's documented per-page maximum.

	var all []*models.Hit
	var page int64 = 1
	for {
		p := search.NewSearchParams()
		if configure != nil {
			configure(p)
		}
		p.Limit = new(pageSize)
		p.Page = new(page)

		resp, err := apiClient.Search.Search(p)
		if err != nil {
			log.Fatal("unable to retrieve folder list.")
		}
		payload := resp.GetPayload()
		all = append(all, payload...)
		if int64(len(payload)) < pageSize {
			break
		}
		page++
	}
	return all
}

// listFolders fetches all folders from Grafana, optionally filtered.
// This is an infrastructure call — it does not use the DashboardService routing.
func (b *baseService) listFolders(filter outbound.Filter) []*domain.NestedHit {
	result := make([]*domain.NestedHit, 0)
	if b.grafanaConf.GetDashboardSettings().IgnoreFilters {
		filter = nil
	}

	rawHits := searchAllPages(b.GetClient(), func(p *search.SearchParams) {
		p.Type = &domain.ApiConsts.SearchTypeFolder
	})

	folderListing := make([]*domain.NestedHit, 0)
	lo.ForEach(rawHits, func(item *models.Hit, index int) {
		newItem := &domain.NestedHit{Hit: item}
		folderListing = append(folderListing, newItem)
	})
	folderUid := getFolderUIDEntityMapByList(folderListing)

	for ndx, val := range folderListing {
		nestedVal := getNestedFolder(val.Title, val.UID, folderUid)
		val.NestedPath = nestedVal
		if filter == nil || filter.Validate(context.Background(), domain.FolderFilter, val) {
			item := folderListing[ndx]
			item.NestedPath = nestedVal
			result = append(result, item)
		}
	}
	return result
}

// getFolderUIDEntityMap builds a UID→NestedHit map from the current folder list.
func (b *baseService) getFolderUIDEntityMap(filter outbound.Filter) map[string]*domain.NestedHit {
	return getFolderUIDEntityMapByList(b.listFolders(filter))
}

// getFolderNameUIDMap builds a NestedPath→UID map from a folder slice.
func (b *baseService) getFolderNameUIDMap(folders []*domain.NestedHit) map[string]string {
	return getFolderMapping(folders,
		func(fld *domain.NestedHit) string { return fld.NestedPath },
		func(fld *domain.NestedHit) string { return fld.UID },
	)
}

// createdFolders creates any missing folders for folderName (supports nested paths).
func (b *baseService) createdFolders(folderName string) (map[string]string, error) {
	return b.createdFoldersWithBaseUID(folderName, "")
}

// createdFoldersWithBaseUID creates any missing folders, assigning uid to the
// leaf folder. Each segment is created in order if it does not already exist.
func (b *baseService) createdFoldersWithBaseUID(folderName string, uid string) (map[string]string, error) {
	namedUIDMap := getFolderMapping(b.listFolders(NewFolderFilter(b.gdgConfig)),
		func(db *domain.NestedHit) string { return db.NestedPath },
		func(fld *domain.NestedHit) *domain.NestedHit { return fld },
	)

	createBaseFolder := func(title, parent, folderUID string) (string, error) {
		request := &models.CreateFolderCommand{
			Title:     title,
			ParentUID: parent,
			UID:       folderUID,
		}
		res, err := b.GetClient().Folders.CreateFolder(request)
		if err != nil {
			return "", err
		}
		return res.GetPayload().UID, nil
	}

	newFoldersMap := make(map[string]string)
	folderPath := strings.Builder{}
	parentUID := ""
	const pathSeparator = string(os.PathSeparator)

	if strings.Contains(folderName, pathSeparator) {
		elements := strings.Split(folderName, pathSeparator)
		for ndx, folder := range elements {
			var (
				cnt     int
				pathErr error
			)
			if ndx == 0 {
				cnt, pathErr = folderPath.WriteString(folder)
			} else {
				cnt, pathErr = fmt.Fprintf(&folderPath, "%s%s", pathSeparator, folder)
			}
			if pathErr != nil || cnt <= 0 {
				log.Fatal("unable to update folder path, critical logic error")
			}
			if val, ok := namedUIDMap[folderPath.String()]; ok {
				parentUID = val.UID
			} else {
				leafUID := ""
				if len(elements)-1 == ndx {
					leafUID = uid
				}
				newUID, err := createBaseFolder(encode.Decode(folder), parentUID, leafUID)
				if err != nil {
					return newFoldersMap, err
				}
				newFoldersMap[folderPath.String()] = newUID
				parentUID = newUID
			}
		}
	} else {
		data, err := createBaseFolder(encode.Decode(folderName), "", uid)
		if err == nil {
			newFoldersMap[folderName] = data
		}
		return newFoldersMap, err
	}

	return newFoldersMap, nil
}

// getDashboardByUid retrieves a single dashboard by UID from the legacy API.
// Available on baseService so non-dashboard service types (e.g. library elements)
// can look up dashboards without depending on DashboardServiceImpl.
func (b *baseService) getDashboardByUid(uid string) (*models.DashboardFullWithMeta, error) {
	data, err := b.GetClient().Dashboards.GetDashboardByUID(uid)
	if err != nil {
		return nil, err
	}
	return data.GetPayload(), nil
}

// IsEnterprise returns true if the connected Grafana instance runs an Enterprise licence.
func (b *baseService) IsEnterprise() bool {
	r, err := b.GetClient().Licensing.GetStatus()
	if err != nil {
		return false
	}
	return r.IsSuccess()
}
