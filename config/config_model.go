package config

//GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	URL              string            `yaml:"url"`
	APIToken         string            `yaml:"token"`
	UserName         string            `yaml:"user_name"`
	Password         string            `yaml:"password"`
	MonitoredFolders []string          `yaml:"watched"`
	Datasource       GrafanaDataSource `yaml:"datasource"`
}

func (s *GrafanaConfig) GetMonitoredFolders() []string {
	if len(s.MonitoredFolders) == 0 {
		return []string{"General"}
	}

	return s.MonitoredFolders

}

//Default datasource credentials
type GrafanaDataSource struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
