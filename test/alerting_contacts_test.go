package test

import (
	"bytes"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling/path"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

func TestContactsCrud(t *testing.T) {
	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", "testing"))
	assert.NoError(t, path.FixTestDir("test", ".."))
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer func() {
		err := cleanup()
		if err != nil {
			slog.Warn("Unable to clean up after alerting contacts crud tests")
		}
	}()
	contactPoints, err := apiClient.ListContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, len(contactPoints), 0, "Validate initial contact list is empty")
	contacts, err := apiClient.UploadContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, len(contacts), 1)
	assert.True(t, slices.Contains(contacts, "discord"))
	contactPoints, err = apiClient.ListContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, len(contactPoints), 1)
	data, err := apiClient.DownloadContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, "test/data/org_main-org/alerting/contacts.json", data)
	rawData, err := os.ReadFile(data)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(rawData, []byte("discord")))
	assert.False(t, bytes.Contains(rawData, []byte("email receiver")))
	contacts, err = apiClient.ClearContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, len(contacts), 1)
}
