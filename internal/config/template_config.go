package config

import (
	"log"
)

type TemplatingConfig struct {
	Entities TemplateEntities `mapstructure:"entities"`
}

type TemplateEntities struct {
	Dashboards []TemplateDashboards `mapstructure:"dashboards"`
}

type TemplateDashboards struct {
	TemplateName      string                    `mapstructure:"template_name"`
	DashboardEntities []TemplateDashboardEntity `mapstructure:"output"`
}

type TemplateDashboardEntity struct {
	Folder           string         `mapstructure:"folder"`
	OrganizationName string         `mapstructure:"organization_name"`
	DashboardName    string         `mapstructure:"dashboard_name"`
	TemplateData     map[string]any `mapstructure:"template_data"`
}

func InitTemplateConfig(override string) {
	if configData == nil {
		log.Fatal("GDG configuration was not able to be loaded, cannot continue")
	}
	var ext, appName string
	var configDirs []string
	if override == "" {
		configDirs, appName, ext = buildConfigSearchPath("config/templates.yml")
	} else {
		configDirs, appName, ext = buildConfigSearchPath(override)
	}
	configData.templatingConfig = new(TemplatingConfig)

	_, err := readViperConfig[TemplatingConfig](appName, configDirs, configData.templatingConfig, ext)
	if err != nil {
		log.Fatal("unable to read templating configuration")
	}
}
