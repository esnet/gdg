package tools_test

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/esnet/gdg/internal/tools/encode"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

func TestFolderEncode(t *testing.T) {
	folderName := "Some Folder Name"
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() support.RootOption) error {
		err := cli.Execute([]string{"tools", "helpers", "folders", "encode", folderName}, optionMockSvc())
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
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() support.RootOption) error {
		err := cli.Execute([]string{"tools", "helpers", "folders", "decode", folderName}, optionMockSvc())
		return err
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()
	expected := encode.DecodeEscapeSpecialChars(folderName)
	assert.True(t, strings.Contains(outStr, "INF Decoded result output="))
	assert.True(t, strings.Contains(outStr, expected))
}
