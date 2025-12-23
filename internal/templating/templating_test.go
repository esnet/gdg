package templating

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling/path"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestGenerate(t *testing.T) {
	// Setup
	assert := assert.New(t)
	assert.NoError(path.FixTestDir("templating", "../.."))
	gdgCfg := config.InitGdgConfig(common.DefaultTestConfig)
	tplCfg := config.InitTemplateConfig(common.DefaultTemplateConfig)
	template := NewTemplate(tplCfg, gdgCfg.GetDefaultGrafanaConfig())
	data, err := template.Generate("template_example")
	assert.Nil(err)
	assert.Equal(len(data), 1)
	generatedFiles := data["template_example"]
	assert.True(slices.Contains(generatedFiles, "test/data/org_main-org/dashboards/General/testing-foobar.json"))
	assert.True(slices.Contains(generatedFiles, "test/data/org_some-other-org/dashboards/Testing/template_example.json"))
	// Remove output to avoid conflicting with other tests
	defer func() {
		os.Remove(generatedFiles[0])
		os.Remove(generatedFiles[1])
	}()

	// Obtain first Config and validate output.
	cfg := config.InitTemplateConfig(common.DefaultTemplateConfig)
	templateCfg := cfg.Entities.Dashboards[0].DashboardEntities[0]
	rawData, err := os.ReadFile("test/data/org_main-org/dashboards/General/testing-foobar.json")
	assert.Nil(err)
	parser := gjson.ParseBytes(rawData)
	val := parser.Get("annotations.list.0.hashKey")
	assert.True(val.Exists())
	expected := service.GetSlug(templateCfg.TemplateData["title"].(string))
	val = parser.Get("annotations.list.0.datasource")
	expected = "elasticsearch"
	assert.Equal(val.String(), expected)
	expected = service.GetSlug(templateCfg.TemplateData["title"].(string))
	valArray := parser.Get("panels.0.link_text").Array()
	val = parser.Get("panels.0.link_url.0")
	lightsources := templateCfg.TemplateData["lightsources"].([]any)
	for ndx, entry := range valArray {
		assert.Equal(entry.String(), lightsources[ndx].(string))
		assert.True(strings.Contains(val.String(), entry.String()))

	}
}
