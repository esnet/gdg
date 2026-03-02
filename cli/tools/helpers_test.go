package tools_test

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports/mocks"
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
