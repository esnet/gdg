package tools

import (
	"log"
	"os"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling/path"
)

func TestMain(m *testing.M) {
	err := path.FixTestDir("tools", "../../..")
	if err != nil {
		log.Fatal(err.Error())
	}
	code := m.Run()
	os.Exit(code)
}
