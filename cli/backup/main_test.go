package backup_test

import (
	"log"
	"os"
	"testing"

	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"

	"github.com/esnet/gdg/pkg/test_tooling/path"
)

func TestMain(m *testing.M) {
	err := path.FixTestDir("backup", "../..")
	if err != nil {
		log.Fatal(err.Error())
	}
	code := m.Run()
	os.Exit(code)
}

func GetOptionMockSvc(testSvc *mocks.GrafanaService) func() domain.RootOption {
	return func() domain.RootOption {
		return func(response *domain.RootCommand) {
			response.SetUpTest(testSvc)
		}
	}
}
