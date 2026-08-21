package tools_test

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/encode"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

func TestFolderEncode(t *testing.T) {
	folderName := "Some Folder Name"
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		err := cli.Execute(rootSvc, []string{"tools", "helpers", "folders", "encode", folderName}, optionMockSvc())
		return err
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()
	expected := encode.EncodeEscapeSpecialChars(folderName)
	assert.True(t, strings.Contains(outStr, "INF Encoded result output="))
	assert.True(t, strings.Contains(outStr, expected))
}

func TestFolderDecode(t *testing.T) {
	folderName := "Some\\+Folder\\+Name"
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		err := cli.Execute(rootSvc, []string{"tools", "helpers", "folders", "decode", folderName}, optionMockSvc())
		return err
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()
	expected := encode.DecodeEscapeSpecialChars(folderName)
	assert.True(t, strings.Contains(outStr, "INF Decoded result output="))
	assert.True(t, strings.Contains(outStr, expected))
}

// ── helpers cipher encode/decode (--value path) ──────────────────────────────

func TestCipherEncodeByValue(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().EncodeValue("super-secret").Return("cipher:encoded-value")
		return cli.Execute(rootSvc, []string{"tools", "helpers", "cipher", "encode", "--value", "super-secret"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "cipher:encoded-value")
}

func TestCipherDecodeByValue(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().DecodeValue("cipher:encoded-value").Return("super-secret")
		return cli.Execute(rootSvc, []string{"tools", "helpers", "cipher", "decode", "--value", "cipher:encoded-value"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "super-secret")
}

func TestCipherHelperAlias(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "helpers", "c"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.Contains(t, lower, "encode")
	assert.Contains(t, lower, "decode")
}
