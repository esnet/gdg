package backup_test

import (
	"log"
	"os"
	"testing"

	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service/mocks"

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

func GetOptionMockSvc(testSvc *mocks.GrafanaService) func() support.RootOption {
	return func() support.RootOption {
		return func(response *support.RootCommand) {
			response.SetUpTest(testSvc)
		}
	}
}
