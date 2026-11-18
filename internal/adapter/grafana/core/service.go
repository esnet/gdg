// Package core holds the shared infrastructure used by every Grafana service
// implementation: HTTP/OpenAPI client construction, App Platform (Kubernetes-style)
// helpers, folder helpers, and other cross-cutting concerns. Feature packages
// (dashboard, alerting, and the top-level api.DashNGoImpl) embed core.Service to
// get this behavior without duplicating it.
package core

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
	"time"

	"github.com/carlmjohnson/requests"
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
	"github.com/jellydator/ttlcache/v3"
	"github.com/samber/lo"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

// Server info keys, used by GetServerInfo and any callers inspecting its result.
const (
	SrvInfoDBKey               = "Database"
	SrvInfoCommitKey           = "Commit"
	SrvInfoVersionKey          = "Version"
	SrvInfoEnterpriseCommitKey = "EnterpriseCommit"
)

// NewClientOpts is a functional option applied to a TransportConfig when
// constructing a Grafana HTTP client.
type NewClientOpts func(transportConfig *client.TransportConfig)

// Service holds the shared infrastructure that every service implementation
// (DashboardServiceImpl, AlertingServiceImpl, api.DashNGoImpl, and any future
// service) needs. Each service embeds its own Service value, obtained via
// NewService, so there is no shared mutable state between them.
type Service struct {
	Extended       outbound.ExtendedApi
	GdgConfig      *configDomain.GDGAppConfiguration
	GrafanaConf    *configDomain.GrafanaConfig
	Storage        outbound.Storage
	Encoder        outbound.CipherEncoder
	LookupResolver outbound.LookupResolver
	LookupSvc      outbound.LookupService
	Resources      ports.Resources

	// discoveredVersions caches API group version discovery results with a 5-minute
	// TTL so long-running processes pick up version changes without paying a network
	// cost on every CRUD operation.
	discoveredVersions *ttlcache.Cache[string, string]
}

// NewService constructs a Service from the given dependencies.
func NewService(
	gdgConfig *configDomain.GDGAppConfiguration,
	grafanaConf *configDomain.GrafanaConfig,
	extendedAPI outbound.ExtendedApi,
	storageEngine outbound.Storage,
	encoder outbound.CipherEncoder,
	resources ports.Resources,
	lookupResolver outbound.LookupResolver,
	lookupSvc outbound.LookupService,
) Service {
	return Service{
		GdgConfig:      gdgConfig,
		GrafanaConf:    grafanaConf,
		Extended:       extendedAPI,
		Storage:        storageEngine,
		Encoder:        encoder,
		Resources:      resources,
		LookupResolver: lookupResolver,
		LookupSvc:      lookupSvc,
	}
}

// ---------------------------------------------------------------------------
// HTTP client helpers
// ---------------------------------------------------------------------------

func (b *Service) getOrgNameClientOpts() NewClientOpts {
	cfg := b.GdgConfig
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

// GetNewClient builds a Grafana OpenAPI client with the given options applied.
// Exported (rather than lowercase) so feature packages and tests outside core
// can still reach it; GetClient/GetBasicAuthClient/GetAdminClient cover the
// common cases and should be preferred.
func (b *Service) GetNewClient(opts ...NewClientOpts) (*client.GrafanaHTTPAPI, *client.TransportConfig) {
	u, err := url.Parse(b.GrafanaConf.GetURL())
	if err != nil {
		log.Fatal("invalid Grafana URL", b.GrafanaConf.GetURL())
	}
	path, err := url.JoinPath(u.Path, "api")
	if err != nil {
		log.Fatal("invalid Grafana URL Path")
	}

	// Build the transport stack from the inside out: TLS base → logging wrapper.
	// This mirrors the pattern in extended.go and ensures the logging middleware
	// always sees traffic regardless of TLS configuration.
	var base = http.DefaultTransport
	if b.GdgConfig.IgnoreSSL() {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
		base = t
	}
	if b.GdgConfig != nil && b.GdgConfig.TransportWrapper != nil {
		base = b.GdgConfig.TransportWrapper(base)
	}

	httpConfig := &client.TransportConfig{
		Host:         u.Host,
		BasePath:     path,
		Schemes:      []string{u.Scheme},
		NumRetries:   b.GdgConfig.GetAppGlobals().RetryCount,
		RetryTimeout: b.GdgConfig.GetAppGlobals().GetRetryTimeout(),
		Client:       &http.Client{Transport: base},
	}

	if b.GrafanaConf.IsBasicAuth() && len(opts) == 1 {
		opts = append(opts, b.getOrgNameClientOpts())
	}
	for _, opt := range opts {
		if opt != nil {
			opt(httpConfig)
		}
	}
	return client.NewHTTPClientWithConfig(strfmt.Default, httpConfig), httpConfig
}

// GetClient returns a Grafana API client, preferring token auth over basic auth.
func (b *Service) GetClient() *client.GrafanaHTTPAPI {
	if b.GrafanaConf.GetAPIToken() != "" {
		grafanaClient, _ := b.GetNewClient(func(clientCfg *client.TransportConfig) {
			clientCfg.APIKey = b.GrafanaConf.GetAPIToken()
		})
		return grafanaClient
	}
	return b.GetBasicAuthClient()
}

func (b *Service) getDefaultBasicOpts() []NewClientOpts {
	return []NewClientOpts{func(clientCfg *client.TransportConfig) {
		clientCfg.BasicAuth = url.UserPassword(b.GrafanaConf.UserName, b.GrafanaConf.GetPassword())
	}}
}

// GetBasicClientWithOpts returns a basic-auth client with additional options applied.
func (b *Service) GetBasicClientWithOpts(opts ...NewClientOpts) *client.GrafanaHTTPAPI {
	allOpts := b.getDefaultBasicOpts()
	allOpts = append(allOpts, opts...)
	grafanaClient, _ := b.GetNewClient(allOpts...)
	return grafanaClient
}

// GetBasicAuthClient returns a basic-auth Grafana API client.
func (b *Service) GetBasicAuthClient() *client.GrafanaHTTPAPI {
	return b.GetBasicClientWithOpts()
}

// GetAdminClient returns the admin client; fatal if admin is not configured.
func (b *Service) GetAdminClient() *client.GrafanaHTTPAPI {
	if !b.GrafanaConf.IsGrafanaAdmin() || b.GrafanaConf.UserName == "" {
		log.Fatal("Unable to get Grafana Admin SecureData.")
	}
	return b.GetBasicClientWithOpts()
}

// GetOrgNameClientOpts exposes getOrgNameClientOpts for callers (e.g. org
// preferences) that need to target a client at a specific configured org by name.
func (b *Service) GetOrgNameClientOpts() NewClientOpts {
	return b.getOrgNameClientOpts()
}

// ---------------------------------------------------------------------------
// Config / globals helpers
// ---------------------------------------------------------------------------

// GetGlobals returns the global app configuration, initialising it if nil.
func (b *Service) GetGlobals() *configDomain.AppGlobals {
	if b.GdgConfig.Global == nil {
		b.GdgConfig.Global = &configDomain.AppGlobals{}
	}
	return b.GdgConfig.Global
}

// IsLocal reports whether the configured storage backend is local disk.
func (b *Service) IsLocal() bool {
	return b.Storage.Name() == storage.LocalStorageType.String()
}

// ---------------------------------------------------------------------------
// Server info
// ---------------------------------------------------------------------------

// GetServerInfo returns basic Grafana server health information.
func (b *Service) GetServerInfo() map[string]any {
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
// Folder helpers — used by DashboardServiceImpl, AlertRules, and any future service
// ---------------------------------------------------------------------------

// SearchAllPages drives a legacy /api/search call to exhaustion, looping on
// Limit/Page. Grafana's search endpoint caps a single page at up to 5000
// results and simply returns whatever fits.
func SearchAllPages(apiClient *client.GrafanaHTTPAPI, configure func(*search.SearchParams)) []*models.Hit {
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

// ListFolders fetches all folders from Grafana, optionally filtered.
// This is an infrastructure call — it does not use the DashboardService routing.
func (b *Service) ListFolders(filter outbound.Filter) []*domain.NestedHit {
	result := make([]*domain.NestedHit, 0)
	if b.GrafanaConf.GetDashboardSettings().IgnoreFilters {
		filter = nil
	}

	rawHits := SearchAllPages(b.GetClient(), func(p *search.SearchParams) {
		p.Type = &domain.ApiConsts.SearchTypeFolder
	})

	folderListing := make([]*domain.NestedHit, 0)
	lo.ForEach(rawHits, func(item *models.Hit, index int) {
		newItem := &domain.NestedHit{Hit: item}
		folderListing = append(folderListing, newItem)
	})
	folderUid := GetFolderUIDEntityMapByList(folderListing)

	for ndx, val := range folderListing {
		nestedVal := GetNestedFolder(val.Title, val.UID, folderUid)
		val.NestedPath = nestedVal
		if filter == nil || filter.Validate(context.Background(), domain.FolderFilter, val) {
			item := folderListing[ndx]
			item.NestedPath = nestedVal
			result = append(result, item)
		}
	}
	return result
}

// GetFolderUIDEntityMap builds a UID→NestedHit map from the current folder list.
func (b *Service) GetFolderUIDEntityMap(filter outbound.Filter) map[string]*domain.NestedHit {
	return GetFolderUIDEntityMapByList(b.ListFolders(filter))
}

// GetFolderNameUIDMap builds a NestedPath→UID map from a folder slice.
func (b *Service) GetFolderNameUIDMap(folders []*domain.NestedHit) map[string]string {
	return GetFolderMapping(folders,
		func(fld *domain.NestedHit) string { return fld.NestedPath },
		func(fld *domain.NestedHit) string { return fld.UID },
	)
}

// CreatedFolders creates any missing folders for folderName (supports nested paths).
func (b *Service) CreatedFolders(folderName string) (map[string]string, error) {
	return b.CreatedFoldersWithBaseUID(folderName, "")
}

// CreatedFoldersWithBaseUID creates any missing folders, assigning uid to the
// leaf folder. Each segment is created in order if it does not already exist.
func (b *Service) CreatedFoldersWithBaseUID(folderName string, uid string) (map[string]string, error) {
	namedUIDMap := GetFolderMapping(b.ListFolders(NewFolderFilter(b.GdgConfig)),
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

// GetDashboardByUid retrieves a single dashboard by UID from the legacy API.
// Available on Service so non-dashboard service types (e.g. library elements)
// can look up dashboards without depending on DashboardServiceImpl.
func (b *Service) GetDashboardByUid(uid string) (*models.DashboardFullWithMeta, error) {
	data, err := b.GetClient().Dashboards.GetDashboardByUID(uid)
	if err != nil {
		return nil, err
	}
	return data.GetPayload(), nil
}

// GetCurrentOrg reports the org for apiClient's identity via GET /api/org,
// which works for basic auth, tokens, and anonymous access alike (unlike
// /api/user/orgs). fatal=true dies on error, for callers that require an org;
// fatal=false returns the error, for callers that should fail soft.
func (b *Service) GetCurrentOrg(apiClient *client.GrafanaHTTPAPI, fatal bool) (*models.OrgDetailsDTO, error) {
	resp, err := apiClient.Org.GetCurrentOrg()
	if err != nil {
		if fatal {
			log.Fatalf("Unable to retrieve current organization, err: %v", err)
		}
		return nil, err
	}
	return resp.GetPayload(), nil
}

// ---------------------------------------------------------------------------
// App Platform (Kubernetes-style API) helpers
// ---------------------------------------------------------------------------

// K8sRestConfig maps GDG's existing Grafana connection settings onto a
// Kubernetes rest.Config pointed at the Grafana App Platform (/apis) endpoint.
func (b *Service) K8sRestConfig() (*rest.Config, error) {
	u, err := url.Parse(b.GrafanaConf.GetURL())
	if err != nil {
		return nil, fmt.Errorf("invalid Grafana URL %q: %w", b.GrafanaConf.GetURL(), err)
	}

	host := strings.TrimSuffix(fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path), "/")

	cfg := &rest.Config{
		Host:    host,
		APIPath: "apis",
	}
	if b.GdgConfig != nil && b.GdgConfig.TransportWrapper != nil {
		cfg.Wrap(transport.WrapperFunc(b.GdgConfig.TransportWrapper))
	}

	if token := b.GrafanaConf.GetAPIToken(); token != "" {
		cfg.BearerToken = token
	} else {
		cfg.Username = b.GrafanaConf.UserName
		cfg.Password = b.GrafanaConf.GetPassword()
	}

	if b.GdgConfig.IgnoreSSL() {
		cfg.TLSClientConfig = rest.TLSClientConfig{Insecure: true}
	}

	return cfg, nil
}

// K8sNamespace maps the active org to an App Platform namespace
// (org 1 -> "default", org N -> "org-N").
func (b *Service) K8sNamespace() string {
	orgID := int64(configDomain.DefaultOrganizationId)

	switch {
	case !b.GrafanaConf.IsBasicAuth():
		if org, err := b.GetCurrentOrg(b.GetClient(), false); err == nil && org != nil {
			orgID = org.ID
		} else {
			slog.Warn("unable to determine the current organization via /api/org; defaulting to org 1/\"default\"", "err", err)
		}
	case b.GrafanaConf.OrganizationName != "":
		if id, err := extended.NewExtendedApi(b.GdgConfig).GetConfiguredOrgId(b.GrafanaConf.OrganizationName); err == nil {
			orgID = id
		}
	}

	if orgID <= configDomain.DefaultOrganizationId {
		return "default"
	}
	return fmt.Sprintf("org-%d", orgID)
}

// apiVersionCacheTTL is how long a discovered API group version is considered
// fresh before re-querying the App Platform.
const apiVersionCacheTTL = 5 * time.Minute

// DiscoverAPIGroupVersion queries the Grafana App Platform discovery endpoint
// (GET /apis) to find the preferred version for the given API group.
// Results are cached for apiVersionCacheTTL. Returns fallback on any failure.
func (b *Service) DiscoverAPIGroupVersion(group, fallback string) string {
	if b.discoveredVersions == nil {
		b.discoveredVersions = ttlcache.New[string, string](
			ttlcache.WithTTL[string, string](apiVersionCacheTTL),
			ttlcache.WithDisableTouchOnHit[string, string](),
		)
	}
	if item := b.discoveredVersions.Get(group); item != nil {
		return item.Value()
	}
	version := b.fetchAPIGroupVersion(group, fallback)
	b.discoveredVersions.Set(group, version, ttlcache.DefaultTTL)
	return version
}

// apiGroupList is a minimal representation of the Kubernetes APIGroupList
// response returned by GET /apis on a Grafana App Platform instance.
type apiGroupList struct {
	Groups []apiGroup `json:"groups"`
}

type apiGroup struct {
	Name             string          `json:"name"`
	PreferredVersion apiGroupVersion `json:"preferredVersion"`
}

type apiGroupVersion struct {
	Version string `json:"version"`
}

// fetchAPIGroupVersion performs the actual GET /apis discovery call.
// It honours IgnoreSSL and TransportWrapper. This is the raw-HTTP path used
// for non-OpenAPI discovery requests.
//
// TODO: replace with k8s.io/client-go/discovery once k8s.io/apimachinery@v0.38
// is released and the structured-merge-diff v6/v7 conflict is resolved.
func (b *Service) fetchAPIGroupVersion(group, fallback string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result apiGroupList
	req := requests.URL(strings.TrimRight(b.GrafanaConf.GetURL(), "/") + "/apis").
		Client(b.GdgConfig.BuildHTTPClient())

	token := b.GrafanaConf.GetAPIToken()
	if token != "" {
		req = req.Header("Authorization", "Bearer "+token)
	} else {
		req = req.BasicAuth(b.GrafanaConf.UserName, b.GrafanaConf.GetPassword())
	}

	if err := req.ToJSON(&result).Fetch(ctx); err != nil {
		slog.Warn("API group version discovery: request failed", "group", group, "err", err)
		return fallback
	}

	for _, g := range result.Groups {
		if g.Name == group {
			if v := g.PreferredVersion.Version; v != "" {
				slog.Debug("discovered API group preferred version", "group", group, "version", v)
				return v
			}
		}
	}
	slog.Warn("API group not found in discovery, using fallback", "group", group, "fallback", fallback)
	return fallback
}

// IsEnterprise returns true if the connected Grafana instance runs an Enterprise licence.
func (b *Service) IsEnterprise() bool {
	r, err := b.GetClient().Licensing.GetStatus()
	if err != nil {
		return false
	}
	return r.IsSuccess()
}
