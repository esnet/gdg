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
	assert.NoError(t, path.FixTestDir("templating", "../.."))
	config.InitGdgConfig(common.DefaultTestConfig)
	config.InitTemplateConfig("templates-example")
	template := NewTemplate()
	data, err := template.Generate("template_example")
	assert.Nil(t, err)
	assert.Equal(t, len(data), 1)
	generatedFiles := data["template_example"]
	assert.True(t, slices.Contains(generatedFiles, "test/data/org_main-org/dashboards/General/testing-foobar.json"))
	assert.True(t, slices.Contains(generatedFiles, "test/data/org_some-other-org/dashboards/Testing/template_example.json"))
	// Remove output to avoid conflicting with other tests
	defer func() {
		os.Remove(generatedFiles[0])
		os.Remove(generatedFiles[1])
	}()

	// Obtain first Config and validate output.
	cfg := config.Config().GetTemplateConfig()
	templateCfg := cfg.Entities.Dashboards[0].DashboardEntities[0]
	rawData, err := os.ReadFile("test/data/org_main-org/dashboards/General/testing-foobar.json")
	assert.Nil(t, err)
	parser := gjson.ParseBytes(rawData)
	val := parser.Get("annotations.list.0.hashKey")
	assert.True(t, val.Exists())
	expected := service.GetSlug(templateCfg.TemplateData["title"].(string))
	val = parser.Get("annotations.list.0.datasource")
	expected = "elasticsearch"
	assert.Equal(t, val.String(), expected)
	expected = service.GetSlug(templateCfg.TemplateData["title"].(string))
	valArray := parser.Get("panels.0.link_text").Array()
	val = parser.Get("panels.0.link_url.0")
	lightsources := templateCfg.TemplateData["lightsources"].([]interface{})
	for ndx, entry := range valArray {
		assert.Equal(t, entry.String(), lightsources[ndx].(string))
		assert.True(t, strings.Contains(val.String(), entry.String()))

	}
}
