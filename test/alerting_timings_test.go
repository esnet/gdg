package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAlertingTimingsCrud(t *testing.T) {
	assert := assert.New(t)
	assert.NoError(os.Setenv("GDG_CONTEXT_NAME", common.TestContextName))

	assert.NoError(path.FixTestDir("test", ".."))
	config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
		return r.Err
	})
	assert.NoError(err)
	assert.NotNil(r)
	defer func() {
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	//
	slog.Info("Uploading Contact Points")
	_, err = apiClient.UploadContactPoints()
	assert.NoError(err)

	timingsList, err := apiClient.ListAlertTimings()
	assert.NoError(err)
	assert.Equal(len(timingsList), 0)
	items, err := apiClient.UploadAlertTimings()
	assert.NoError(err)
	assert.Equal(len(items), 1)
	timingsList, err = apiClient.ListAlertTimings()
	assert.NoError(err)

	assert.Equal(len(timingsList), 1)
	timingItem := timingsList[0]
	assert.Equal(timingItem.Name, "after-hours")
	assert.Equal(len(timingItem.TimeIntervals), 2)
	timedInterval := lo.FindOrElse(timingItem.TimeIntervals, nil, func(item *models.TimeIntervalItem) bool {
		return item.Location == "America/New_York"
	})
	assert.NotEmpty(timedInterval)

	expected := models.TimeIntervalItem{
		Location:    "America/New_York",
		DaysOfMonth: []string{"7:31"},
		Months:      []string{"1:11"},
		Weekdays:    []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
		Years:       []string{"2021:2031"},
		Times: []*models.TimeIntervalTimeRange{
			{
				EndTime:   "23:59",
				StartTime: "17:00",
			},
			{
				EndTime:   "09:00",
				StartTime: "01:00",
			},
		},
	}
	assert.True(diffStruct(timedInterval, &expected))

	_, err = apiClient.DownloadAlertTimings()
	assert.NoError(err)
	err = apiClient.ClearAlertTimings()
	assert.NoError(err)
	timingsList, err = apiClient.ListAlertTimings()
	assert.NoError(err)
	assert.Equal(len(timingsList), 0)
	//validate download dataset
	_, err = apiClient.UploadAlertTimings() //upload the downloaded version
	assert.NoError(err)
	timingsList, err = apiClient.ListAlertTimings()
	assert.NoError(err)
	assert.Equal(len(timingsList), 1)
}
