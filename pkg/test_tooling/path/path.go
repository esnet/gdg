package path

import (
	"os"
	"strings"
)

const testEnv = "TESTING"

func FixTestDir(packageName string, newPath string) error {
	err := os.Setenv(testEnv, "1")
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
