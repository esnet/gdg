package config

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/esnet/gdg/internal/config/domain"
)

type formSelection string

const (
	basicAuthForm formSelection = "basicAuth"
	tokenAuthForm formSelection = "tokenAuth"
	bothAuthForm  formSelection = "bothAuth"
)

func (s formSelection) String() string {
	return string(s)
}

func buildFormGroups(authType string, config *domain.GrafanaConfig) []*huh.Group {
	groups := make([]*huh.Group, 0)
	basicGrps := huh.NewGroup(
		huh.NewInput().
			Value(&config.UserName).
			Title("Grafana Username").Description("Grafana Username"),
		huh.NewInput().
			Value(&config.Password).
			Title("Grafana Password").
			Description("Grafana Username").
			EchoMode(huh.EchoModePassword),
	)
	tokenGrps := huh.NewGroup(
		huh.NewInput().
			Value(&config.APIToken).
			Title("Grafana Token").Description("Grafana Token"),
	).
		WithShowHelp(false).
		WithShowErrors(false)

	switch authType {
	case basicAuthForm.String():
		groups = append(groups, basicGrps)
	case tokenAuthForm.String():
		groups = append(groups, tokenGrps)
	case bothAuthForm.String():
		groups = append(groups, []*huh.Group{basicGrps, tokenGrps}...)
	}
	groups = append(groups, huh.NewGroup(
		huh.NewInput().
			Description("Destination Folder?").
			Value(&config.OutputPath),
		huh.NewInput().
			Description("What is the Grafana URL include http(s)?").
			Value(&config.URL),
	),
	)

	return groups
}

func (s *Configuration) FormCode(name string) {
	var authType string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(
					huh.NewOption("Basic Authentication", basicAuthForm.String()),
					huh.NewOption("Token/Service Authentication", tokenAuthForm.String()),
					huh.NewOption("Both", bothAuthForm.String()),
				).
				Value(&authType).
				Title("Choose your Auth Mechanism").
				Description("This will determine your Authentication type"),
		),
	).
		WithShowHelp(false).
		WithShowErrors(false).Run()
	if err != nil {
		log.Fatal("unable to get auth selection from user")
	}

	newConfig := &domain.GrafanaConfig{
		ConnectionSettings: &domain.ConnectionSettings{
			MatchingRules: make([]domain.RegexMatchesList, 0),
		},
	}
	newConfig.OrganizationName = "Main Org."
	err = huh.NewForm(buildFormGroups(authType, newConfig)...).Run()
	if err != nil {
		log.Fatalf("Could not set grafana config: %v", err)
	}

	var (
		connectionUser     string
		connectionPassword string
	)
	var folders string
	err = huh.NewForm(huh.NewGroup(
		huh.NewInput().Description("Grafana Folders to monitor (comma delimited list)").Value(&folders),
	),
		huh.NewGroup(
			huh.NewInput().Description("Grafana Connection Default User").Value(&connectionUser),
			huh.NewInput().Description("Grafana Connection Default User").EchoMode(huh.EchoModePassword).Value(&connectionPassword),
		),
	).Run()
	if err != nil {
		log.Fatalf("Unable to get folders and Connection Auth Settings")
	}
	defaultDs := domain.GrafanaConnection{
		"user":              connectionUser,
		"basicAuthPassword": connectionPassword,
	}
	// newConfig.
	if folders != "" {
		newConfig.MonitoredFolders = strings.Split(folders, ",")
	} else {
		newConfig.MonitoredFolders = []string{"General"}
	}
	securePath := domain.SecureSecretsResource
	location := filepath.Join(newConfig.OutputPath, string(securePath))
	err = os.MkdirAll(location, 0o750)
	if err != nil {
		log.Fatalf("unable to create default secret location.  location: %s, %v", location, err)
	}
	data, err := json.MarshalIndent(&defaultDs, "", "    ")
	if err != nil {
		log.Fatalf("unable to turn map into json representation.  location: %s, %v", location, err)
	}
	secretFileLocation := filepath.Join(location, "default.json")
	err = os.WriteFile(secretFileLocation, data, 0o600)
	if err != nil {
		log.Fatalf("unable to write secret default file.  location: %s, %v", secretFileLocation, err)
	}

	newConfig.ConnectionSettings.MatchingRules = []domain.RegexMatchesList{
		{
			Rules: []domain.MatchingRule{
				{
					Field: "name",
					Regex: ".*",
				},
			},
			SecureData: "default.json",
		},
	}

	contextMap := s.GetGDGConfig().GetContexts()
	contextMap[name] = newConfig
	s.GetGDGConfig().ContextName = name

	err = s.SaveToDisk(false)
	if err != nil {
		log.Fatal("could not save configuration.")
	}
	slog.Info("New configuration has been created", "newContext", name)
}

func (s *Configuration) NewContext(name string) {
	s.FormCode(name)
}
