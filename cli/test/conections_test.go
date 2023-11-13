package test

import (
	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"strings"
	"testing"
)

func TestConnectionCommand(t *testing.T) {
	testSvc := new(mocks.GrafanaService)
	getMockSvc := func() service.GrafanaService {
		return testSvc
	}
	resp := []models.DataSourceListItemDTO{
		{
			ID:        5,
			Name:      "Hello",
			UID:       "magicUid",
			Type:      "elasticsearch",
			IsDefault: false,
		},
	}

	testSvc.On("InitOrganizations").Return(nil)
	testSvc.On("ListConnections", mock.Anything).Return(resp)

	optionMockSvc := func() support.RootOption {
		return func(response *support.RootCommand) {
			response.GrafanaSvc = getMockSvc
		}
	}
	r, w, cleanup := InterceptStdout()

	err := cli.Execute("testing.yml", []string{"backup", "connections", "list"}, optionMockSvc())
	assert.Nil(t, err)
	defer cleanup()
	w.Close()

	out, _ := io.ReadAll(r)
	outStr := string(out)
	assert.True(t, strings.Contains(outStr, "magicUid"))
	assert.True(t, strings.Contains(outStr, "Hello"))
}
