package test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/safaci2000/grafana-openapi-client-go/models"
	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

func TestPoliesCrud(t *testing.T) {
	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", common.TestContextName))

	assert.NoError(t, path.FixTestDir("test", ".."))
	config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		err := r.CleanUp()
		if err != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	// Upload Contact points first.
	_, err = apiClient.UploadContactPoints()
	assert.NoError(t, err)
	//
	policies, err := apiClient.ListAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, len(policies.Routes), 0, "Validate initial contact list is empty")
	policiesListing, err := apiClient.UploadAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, len(policiesListing.Routes), 2)
	route := lo.FindOrElse(policiesListing.Routes, nil, func(item *models.Route) bool {
		return item.Receiver == "slack"
	})
	assert.NotNil(t, route)
	assert.Equal(t, len(route.ObjectMatchers[0]), 3)
	assert.Equal(t, route.ObjectMatchers[0][2], "23")

	policies, err = apiClient.ListAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, len(policies.Routes), 2)
	data, err := apiClient.DownloadAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, "test/data/org_main-org/alerting/policies.json", data)
	rawData, err := os.ReadFile(data)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(rawData, []byte("grafana_folder")))
	assert.True(t, bytes.Contains(rawData, []byte("alertname")))
	err = apiClient.ClearAlertNotifications()
	assert.NoError(t, err)
	policies, err = apiClient.ListAlertNotifications()
	assert.Equal(t, len(policies.Routes), 0)
}
