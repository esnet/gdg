package backup

import (
	"io"
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	testSvc.EXPECT().InitOrganizations().Return()
	testSvc.EXPECT().ListConnections(mock.Anything).Return(resp)

	optionMockSvc := func() support.RootOption {
		return func(response *support.RootCommand) {
			response.GrafanaSvc = getMockSvc
		}
	}
	r, w, cleanup := test_tooling.InterceptStdout()

	err := cli.Execute([]string{"backup", "connections", "list"}, optionMockSvc())
	assert.Nil(t, err)
	defer cleanup()
	assert.NoError(t, w.Close())

	out, _ := io.ReadAll(r)
	outStr := string(out)
	assert.True(t, strings.Contains(outStr, "magicUid"))
	assert.True(t, strings.Contains(outStr, "Hello"))
}
