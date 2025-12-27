package api

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// GetConfiguredOrgId needed to call grafana API in order to configure the Grafana API correctly.  Invoking
// this endpoint manually to avoid a circular dependency.
func (extended *ExtendedApi) GetConfiguredOrgId(orgName string) (int64, error) {
	var result []*models.UserOrgDTO
	fetch := func() error {
		req := extended.getRequestBuilder().
			Path("api/user/orgs").
			ToJSON(&result).
			Method(http.MethodGet)

		if extended.debug {
			log.Printf("%v", req)
		}

		return req.Fetch(context.Background())
	}

	/* There's something goofy here.  This seems to fail sporadically in grafana if we keep swapping orgs too fast.
	   This is a safety check that should ideally never be triggered, but if the URL fails, then we retry a few times
		before finally giving up.
	*/
	delay := time.Second * 5
	var count uint = 5
	// Giving user configured value preference over defaults
	if extended.appCfg.GetAppGlobals().RetryCount != 0 {
		count = uint(extended.appCfg.GetAppGlobals().RetryCount) // #nosec G115
	}
	if extended.appCfg.GetAppGlobals().GetRetryTimeout() != time.Millisecond*100 {
		delay = extended.appCfg.GetAppGlobals().GetRetryTimeout()
	}
	err := retry.Do(fetch,
		retry.Attempts(count),
		retry.Delay(delay),
		retry.OnRetry(func(n uint, err error) {
			slog.Info("Retrying request after error",
				slog.String("orgName", orgName),
				slog.Any("err", err))
		}))
	if err != nil {
		return 0, err
	}
	for _, entity := range result {
		if entity.Name == orgName {
			return entity.OrgID, nil
		}
	}
	return 0, errors.New("org not found")
}
