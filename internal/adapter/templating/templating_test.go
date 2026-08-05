package templating

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/pkg/test_tooling/path"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// ---------------------------------------------------------------------------
// ListTemplates
// ---------------------------------------------------------------------------

func TestListTemplates_ReturnsAllNames(t *testing.T) {
	cfg := &config_domain.TemplatingConfig{
		Entities: config_domain.TemplateEntities{
			Dashboards: []config_domain.TemplateDashboards{
				{TemplateName: "alpha"},
				{TemplateName: "beta"},
				{TemplateName: "gamma"},
			},
		},
	}
	// gdgCfg is not used by ListTemplates, so a zero value is fine.
	tmpl := NewTemplate(cfg, &config_domain.GrafanaConfig{})
	names := tmpl.ListTemplates()
	require.Len(t, names, 3)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, names)
}

func TestListTemplates_EmptyConfig(t *testing.T) {
	cfg := &config_domain.TemplatingConfig{}
	tmpl := NewTemplate(cfg, &config_domain.GrafanaConfig{})
	names := tmpl.ListTemplates()
	assert.Empty(t, names)
}

func TestListTemplates_SingleEntry(t *testing.T) {
	cfg := &config_domain.TemplatingConfig{
		Entities: config_domain.TemplateEntities{
			Dashboards: []config_domain.TemplateDashboards{
				{TemplateName: "only-one"},
			},
		},
	}
	tmpl := NewTemplate(cfg, &config_domain.GrafanaConfig{})
	names := tmpl.ListTemplates()
	require.Len(t, names, 1)
	assert.Equal(t, "only-one", names[0])
}

// ---------------------------------------------------------------------------
// QuotedStringJoin (package-level fns entry)
// ---------------------------------------------------------------------------

func TestQuotedStringJoin_FormatsCorrectly(t *testing.T) {
	fn, ok := fns["QuotedStringJoin"].(func([]any) string)
	require.True(t, ok, "QuotedStringJoin should be registered in fns map")

	result := fn([]any{"hello", "world"})
	assert.Equal(t, `"hello","world"`, result)
}

func TestQuotedStringJoin_SingleItem(t *testing.T) {
	fn := fns["QuotedStringJoin"].(func([]any) string)
	result := fn([]any{"only"})
	assert.Equal(t, `"only"`, result)
}

func TestQuotedStringJoin_EmptySlice(t *testing.T) {
	fn := fns["QuotedStringJoin"].(func([]any) string)
	result := fn([]any{})
	assert.Equal(t, "", result)
}

func TestQuotedStringJoin_SpecialChars(t *testing.T) {
	fn := fns["QuotedStringJoin"].(func([]any) string)
	// Values that contain quotes are rendered via %q which escapes them.
	result := fn([]any{`say "hi"`})
	assert.Equal(t, `"say \"hi\""`, result)
}

// ---------------------------------------------------------------------------
// Generate (existing test)
// ---------------------------------------------------------------------------

func TestGenerate(t *testing.T) {
	// Setup
	assert := assert.New(t)
	assert.NoError(path.FixTestDir("templating", "../../.."))
	gdgCfg := config.NewConfig(common.DefaultTestConfig)
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
	resourceHelper := resources.NewHelpers()
	// Obtain first Config and validate output.
	cfg := config.InitTemplateConfig(common.DefaultTemplateConfig)
	templateCfg := cfg.Entities.Dashboards[0].DashboardEntities[0]
	rawData, err := os.ReadFile("test/data/org_main-org/dashboards/General/testing-foobar.json")
	assert.Nil(err)
	parser := gjson.ParseBytes(rawData)
	val := parser.Get("annotations.list.0.hashKey")
	assert.True(val.Exists())
	expected := resourceHelper.GetSlug(templateCfg.TemplateData["title"].(string))
	val = parser.Get("annotations.list.0.datasource")
	expected = "elasticsearch"
	assert.Equal(val.String(), expected)
	expected = resourceHelper.GetSlug(templateCfg.TemplateData["title"].(string))
	valArray := parser.Get("panels.0.link_text").Array()
	val = parser.Get("panels.0.link_url.0")
	lightSources := templateCfg.TemplateData["lightsources"].([]any)
	for ndx, entry := range valArray {
		assert.Equal(entry.String(), lightSources[ndx].(string))
		assert.True(strings.Contains(val.String(), entry.String()))

	}
}
