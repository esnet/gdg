package config_domain

type AlertSettings struct {
	ContactSettings ContactPointSettings `mapstructure:"contact_points" yaml:"contact_points,omitempty"`
}

type ContactPointSettings struct {
	FilterRules []MatchingRule `mapstructure:"filters" yaml:"filters,omitempty"`
}

// FiltersEnabled returns true if the filters are enabled for the resource type
func (cp *ContactPointSettings) FiltersEnabled() bool {
	return len(cp.FilterRules) > 0
}
