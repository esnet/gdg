package config

import (
	"log"

	"github.com/spf13/viper"
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
	Folder           string                 `mapstructure:"folder"`
	OrganizationName string                 `mapstructure:"organization_name"`
	DashboardName    string                 `mapstructure:"dashboard_name"`
	TemplateData     map[string]interface{} `mapstructure:"template_data"`
}

func InitTemplateConfig(override string) {
	if configData == nil {
		log.Fatal("GDG configuration was not able to be loaded, cannot continue")
	}
	appName := "templates"
	configDirs := buildConfigSearchPath(override, &appName)
	configData.templatingConfig = new(TemplatingConfig)

	v, err := readViperConfig[TemplatingConfig](appName, configDirs, configData.templatingConfig)
	if err != nil {
		log.Fatal("unable to read templating configuration")
	}
	if configData.viperConfiguration == nil {
		configData.viperConfiguration = make(map[string]*viper.Viper)
	}
	configData.viperConfiguration[ViperTemplateConfig] = v
}
