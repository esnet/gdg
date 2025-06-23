package domain

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

func (s *TemplatingConfig) GetTemplate(name string) (*TemplateDashboards, bool) {
	for ndx, t := range s.Entities.Dashboards {
		if t.TemplateName == name {
			return &s.Entities.Dashboards[ndx], true
		}
	}

	return nil, false
}
