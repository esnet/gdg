package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/internal/tools"
	"github.com/esnet/gdg/internal/tools/encode"
	resourceTypes "github.com/esnet/gdg/pkg/config/domain"
	"github.com/esnet/gdg/pkg/plugins/secure"
	"github.com/esnet/gdg/pkg/plugins/secure/contract"
	"gopkg.in/yaml.v3"
)

type formSelection string

func (s formSelection) String() string {
	return string(s)
}

const (
	basicAuthForm formSelection = "basicAuth"
	tokenAuthForm formSelection = "tokenAuth"
	bothAuthForm  formSelection = "bothAuth"
)

// CreateNewContext prompts the user to configure a new Grafana context with authentication, folders,
// and default connection settings. It builds the configuration, writes secure files, updates
// the internal context map, saves the config to disk, and logs completion.
func CreateNewContext(app *domain.GDGAppConfiguration, name string) {
	var encoder contract.CipherEncoder
	if !app.PluginConfig.Disabled && app.PluginConfig.CipherPlugin != nil {
		encoder = secure.NewPluginCipherEncoder(app.PluginConfig.CipherPlugin, app.SecureConfig)
	} else {
		encoder = secure.NoOpEncoder{}
	}
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

	newConfig := domain.NewGrafanaConfig(name)
	newConfig.ConnectionSettings = &domain.ConnectionSettings{
		MatchingRules: make([]*domain.RegexMatchesList, 0),
	}
	newConfig.OrganizationName = "Main Org."
	secure := domain.SecureModel{}
	err = huh.NewForm(buildFormGroups(authType, newConfig, &secure)...).Run()
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

	const passKey = "basicAuthPassword"
	defaultDs := domain.GrafanaConnection{
		"user":  connectionUser,
		passKey: connectionPassword,
	}
	// newConfig.
	if folders != "" {
		newConfig.MonitoredFolders = strings.Split(folders, ",")
		for ndx, item := range newConfig.MonitoredFolders {
			newVal := encode.EncodePath(encode.EncodeEscapeSpecialChars, item)
			newConfig.MonitoredFolders[ndx] = newVal
		}
	} else {
		newConfig.MonitoredFolders = []string{"General"}
	}
	securePath := resourceTypes.SecureSecretsResource
	location := filepath.Join(newConfig.OutputPath, string(securePath))
	err = os.MkdirAll(location, 0o750)
	if err != nil {
		log.Fatalf("unable to create default secret location.  location: %s, %v", location, err)
	}

	secretFileLocation := filepath.Join(location, "default.yaml")
	if encoder != nil {
		newVal, encodeErr := encoder.EncodeValue(defaultDs.Password())
		if encodeErr == nil {
			defaultDs[passKey] = newVal
		}
	}

	err = writeSecureFileData(defaultDs, secretFileLocation)
	if err != nil {
		log.Fatalf("unable to write secret default file.  location: %s, %v", secretFileLocation, err)
	}
	newConfig.ConnectionSettings.MatchingRules = []*domain.RegexMatchesList{
		{
			Rules: []domain.MatchingRule{
				{
					Field: "name",
					Regex: ".*",
				},
			},
			SecureData: "default.yaml",
		},
	}

	// Auth location
	secretFileLocation = fmt.Sprintf("%s.yaml", newConfig.GetAuthLocation())
	secure.UpdateSecureModel(encoder.EncodeValue)

	err = writeSecureFileData(secure, secretFileLocation)
	if err != nil {
		log.Fatalf("unable to write secret auth file.  location: %s, %v", secretFileLocation, err)
	}

	contextMap := app.GetContexts()
	contextMap[name] = newConfig
	app.ContextName = name

	err = app.SaveToDisk(false)
	if err != nil {
		log.Fatal("could not save configuration.")
	}
	slog.Info("New configuration has been created", "newContext", name)
}

// writeSecureFileData marshals an object to JSON and writes it to a file with 0600 permissions.
func writeSecureFileData[T any](object T, location string) error {
	data, err := yaml.Marshal(&object)
	if err != nil {
		log.Fatalf("unable to turn map into yaml representation.  location: %s, %v", location, err)
	}
	err = os.WriteFile(location, data, 0o600)
	return err

}

// buildFormGroups creates form groups for Grafana authentication and configuration.
// It returns a slice of *huh.Group based on authType, including username/password,
// token, output path, and URL inputs.
func buildFormGroups(authType string, config *domain.GrafanaConfig, secureModel *domain.SecureModel) []*huh.Group {
	groups := make([]*huh.Group, 0)
	basicGrps := huh.NewGroup(
		huh.NewInput().
			Value(&config.UserName).
			Title("Grafana Username").Description("Grafana Username"),
		huh.NewInput().
			Value(&secureModel.Password).
			Title("Grafana Password").
			Description("Grafana Username").
			EchoMode(huh.EchoModePassword),
	)
	tokenGrps := huh.NewGroup(
		huh.NewInput().
			Value(&secureModel.Token).
			Title("Grafana Token").
			Description("Grafana Token").
			EchoMode(huh.EchoModePassword),
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

// DeleteContext remove a given context
func DeleteContext(app *domain.GDGAppConfiguration, name string) {
	name = strings.ToLower(name) // ensure name is lower case
	contexts := app.GetContexts()
	ctx, ok := contexts[name]
	if !ok {
		log.Fatalf("Context not found, cannot delete context: %s", name)
		return
	}
	secureLoc := ctx.SecureLocation()
	fileName := filepath.Join(secureLoc, fmt.Sprintf("auth_%s.yaml", name))
	delete(contexts, name)
	if len(contexts) != 0 {
		for key := range contexts {
			app.ContextName = key
			break
		}
	}

	err := app.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	if _, statErr := os.Stat(fileName); statErr != nil {
		slog.Warn("auth file does not exists")
	} else {
		errRemove := os.Remove(fileName)
		if errRemove != nil {
			slog.Warn("failed to remove auth file", "file", fileName)
		}
	}

	slog.Info("Deleted context and set new context to", "deletedContext", name, "newActiveContext", app.ContextName)
}

// CopyContext Makes a copy of the specified context and write to disk
func CopyContext(app *domain.GDGAppConfiguration, src, dest string) {
	// Validate context
	contexts := app.GetContexts()
	if len(contexts) == 0 {
		log.Fatal("Cannot set context.  No valid configuration found in gdg.yml")
	}
	cfg, ok := contexts[src]
	if !ok {
		log.Fatalf("Cannot find context to: '%s'.  No valid configuration found in gdg.yml", src)
	}
	newCopy, err := tools.DeepCopy(*cfg)
	if err != nil {
		log.Fatal("unable to make a copy of contexts")
	}
	contexts[dest] = newCopy
	app.ContextName = dest
	err = app.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	slog.Info("Copied context to destination, please check your config to confirm", "sourceContext", src, "destinationContext", dest)
}

// ClearContexts resets all contexts to a single default example context and saves the config.```
func ClearContexts(app *domain.GDGAppConfiguration) {
	newContext := make(map[string]*domain.GrafanaConfig)
	newContext["example"] = domain.NewGrafanaConfig("example")
	app.Contexts = newContext
	app.ContextName = "example"
	err := app.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}

	slog.Info("All contexts were cleared")
}
