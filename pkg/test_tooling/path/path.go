package path

import (
	"os"
	"strings"
)

const TestEnvKey = "TESTING"

func FixTestDir(packageName string, newPath string) error {
	err := os.Setenv(TestEnvKey, "1")
	if err != nil {
		return err
	}
	dir, _ := os.Getwd()
	if strings.Contains(dir, packageName) {
		err = os.Chdir(newPath)
		if err != nil {
			return err
		}
	}
	return nil
}
