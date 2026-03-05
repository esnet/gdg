package version

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionDefaults(t *testing.T) {
	assert.Equal(t, "DEVEL", Version)
}

func TestGoVersionMatchesRuntime(t *testing.T) {
	assert.Equal(t, runtime.Version(), GoVersion)
}

func TestOsArchNotEmpty(t *testing.T) {
	assert.NotEmpty(t, OsArch)
}

func TestPrintVersionInfo_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintVersionInfo()
	})
}
