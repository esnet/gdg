package backup_test

import (
	"io"
	"strings"
	"testing"

	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/test_tooling"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnectionCommand(t *testing.T) {
	testSvc := new(mocks.GrafanaService)
	getMockSvc := func() outbound.GrafanaService {
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

	testSvc.EXPECT().Login().Return()
	testSvc.EXPECT().InitOrganizations().Return()
	testSvc.EXPECT().ListConnections(mock.Anything).Return(resp)

	optionMockSvc := func() domain.RootOption {
		return func(response *domain.RootCommand) {
			response.SetUpTest(getMockSvc())
		}
	}
	r, w, cleanup := test_tooling.InterceptStdout()

	rootSvc := cli.NewRootService()
	err := cli.Execute(rootSvc, []string{"backup", "connections", "list"}, optionMockSvc())
	assert.Nil(t, err)
	defer cleanup()
	assert.NoError(t, w.Close())

	out, _ := io.ReadAll(r)
	outStr := string(out)
	assert.True(t, strings.Contains(outStr, "magicUid"))
	assert.True(t, strings.Contains(outStr, "Hello"))
}
