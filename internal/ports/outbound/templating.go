package outbound

// Templating defines an interface for generating and listing dashboard templates based on templating configuration.
type Templating interface {
	Generate(templateName string) (map[string][]string, error)
	ListTemplates() []string
}
