package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/cipher"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/config/config_domain"
	resourceTypes "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/encode"
	"github.com/esnet/gdg/pkg/tools"
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

// CreateNewContext prompts the user to configure a new Grafana context with authentication,
// watched folders, connection filters, and credential rules. It builds the configuration,
// writes secure files, updates the internal context map, saves the config to disk, and logs completion.
func CreateNewContext(app *config_domain.GDGAppConfiguration, name string) {
	var encoder ports.CipherEncoder
	if !app.PluginConfig.Disabled && app.PluginConfig.CipherPlugin != nil {
		encoder = cipher.NewPluginCipherEncoder(app.PluginConfig.CipherPlugin, app.SecureConfig)
	} else {
		encoder = noop.NoOpEncoder{}
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

	newConfig := config_domain.NewGrafanaConfig(config_domain.WithContextName(name))
	newConfig.ConnectionSettings = &config_domain.ConnectionSettings{
		MatchingRules: make([]*config_domain.RegexMatchesList, 0),
	}
	newConfig.OrganizationName = "Main Org."
	secure := config_domain.SecureModel{}
	err = huh.NewForm(buildFormGroups(authType, newConfig, &secure)...).Run()
	if err != nil {
		log.Fatalf("Could not set grafana config: %v", err)
	}

	// Watched folders — seamless one-at-a-time workflow with encoding and regex testing
	newConfig.MonitoredFolders = configureWatchedFolders()

	// Connection filters
	newConfig.ConnectionSettings.FilterRules = configureConnectionFilters()

	// Default connection credentials
	var (
		connectionUser     string
		connectionPassword string
	)
	err = huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Connection Default User").Description("Default user for Grafana data source connections").Value(&connectionUser),
		huh.NewInput().Title("Connection Default Password").Description("Default password for Grafana data source connections").EchoMode(huh.EchoModePassword).Value(&connectionPassword),
	)).Run()
	if err != nil {
		log.Fatalf("Unable to get Connection Auth Settings")
	}

	const passKey = "basicAuthPassword"
	defaultDs := config_domain.GrafanaConnection{
		"user":  connectionUser,
		passKey: connectionPassword,
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

	// Credential rules — loop with default catch-all appended
	newConfig.ConnectionSettings.MatchingRules = configureCredentialRules()

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

// configureWatchedFolders interactively prompts the user to build the watched-folder list
// one entry at a time. Each entry is URL- and regex-encoded automatically, and the user
// may test any entered pattern against a sample value before finishing.
func configureWatchedFolders() []string {
	var folders []string

	for {
		currentList := "none"
		if len(folders) > 0 {
			currentList = strings.Join(folders, ", ")
		}

		type folderAction string
		const (
			actionAdd  folderAction = "add"
			actionTest folderAction = "test"
			actionDone folderAction = "done"
		)

		var action string
		opts := []huh.Option[string]{
			huh.NewOption("Add a folder", string(actionAdd)),
		}
		if len(folders) > 0 {
			opts = append(opts,
				huh.NewOption("Test a folder name against current list", string(actionTest)),
				huh.NewOption("Done", string(actionDone)),
			)
		}

		err := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Watched Folders").
				Description(fmt.Sprintf("Folders currently in list: [%s]", currentList)).
				Options(opts...).
				Value(&action),
		)).WithShowHelp(false).WithShowErrors(false).Run()
		if err != nil {
			// Treat cancelled form (Ctrl+C / Esc) as "done"
			break
		}

		fa := folderAction(action)
		if fa == actionDone {
			break
		}

		switch fa {
		case actionAdd:
			var folderName string
			err = huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Folder Name").
					Description("Enter the Grafana folder name. Spaces and special characters will be encoded automatically.").
					Value(&folderName),
			)).WithShowHelp(false).WithShowErrors(false).Run()
			if err != nil || strings.TrimSpace(folderName) == "" {
				continue
			}
			encoded := encode.EncodePath(encode.EncodeEscapeSpecialChars, strings.TrimSpace(folderName))
			if encoded != strings.TrimSpace(folderName) {
				fmt.Printf("\n  ✎  '%s'  →  encoded as  '%s'\n\n", folderName, encoded)
			}
			folders = append(folders, encoded)

		case actionTest:
			var testValue string
			err = huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Test Folder Name").
					Description(fmt.Sprintf("Enter a folder name to test against the current list: [%s]", currentList)).
					Value(&testValue),
			)).WithShowHelp(false).WithShowErrors(false).Run()
			if err != nil || strings.TrimSpace(testValue) == "" {
				continue
			}
			encodedTest := encode.EncodePath(encode.EncodeEscapeSpecialChars, strings.TrimSpace(testValue))
			fmt.Printf("\n  Testing '%s' (encoded: '%s') against current folders:\n", testValue, encodedTest)
			matched := false
			for _, f := range folders {
				p, compErr := regexp.Compile(f)
				if compErr != nil {
					fmt.Printf("  ⚠  Pattern '%s' is not a valid regex: %v\n", f, compErr)
					continue
				}
				if p.MatchString(encodedTest) {
					fmt.Printf("  ✓  Matches pattern '%s'\n", f)
					matched = true
				}
			}
			if !matched {
				fmt.Printf("  ✗  No match found in current folder list\n")
			}
			fmt.Println()
		}
	}

	if len(folders) == 0 {
		folders = []string{"General"}
		fmt.Println("  No folders specified — defaulting to 'General'")
	}
	return folders
}

// configureConnectionFilters interactively builds the connections.filters slice.
// Each filter specifies a field, a regex, and whether it is inclusive (allowlist) or
// exclusive (denylist, the default). The user may test each regex before confirming it.
func configureConnectionFilters() []config_domain.MatchingRule {
	var filters []config_domain.MatchingRule

	var addFilters bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Configure Connection Filters?").
			Description("Filters control which data sources are included or excluded during backup/restore.\nExample: exclude all DEV connections, or include only elasticsearch sources.").
			Value(&addFilters),
	)).WithShowHelp(false).WithShowErrors(false).Run()
	if err != nil || !addFilters {
		return filters
	}

	for {
		var field, regex string
		var inclusive bool

		err = huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Filter Field").
				Description("JSON field to match on (e.g. 'name', 'type', 'url')").
				Value(&field),
			huh.NewInput().
				Title("Filter Regex").
				Description("Regular expression to match against the field value").
				Value(&regex),
			huh.NewConfirm().
				Title("Inclusive filter?").
				Description("Inclusive = keep only matches (allowlist). Exclusive = drop matches (denylist, default).").
				Value(&inclusive),
		)).WithShowHelp(false).WithShowErrors(false).Run()
		if err != nil {
			break
		}

		field = strings.TrimSpace(field)
		regex = strings.TrimSpace(regex)
		if field == "" || regex == "" {
			fmt.Println("  ⚠  Field and regex are required — skipping.")
		} else if _, compErr := regexp.Compile(regex); compErr != nil {
			fmt.Printf("  ⚠  Invalid regex '%s': %v — skipping.\n", regex, compErr)
		} else {
			filters = append(filters, config_domain.MatchingRule{
				Field:     field,
				Regex:     regex,
				Inclusive: inclusive,
			})
			filterType := "exclusive (denylist)"
			if inclusive {
				filterType = "inclusive (allowlist)"
			}
			fmt.Printf("\n  ✓  Added filter: field='%s'  regex='%s'  type=%s\n", field, regex, filterType)

			// Offer inline regex test
			if shouldTestRegex(fmt.Sprintf("field='%s' regex='%s'", field, regex)) {
				runRegexTest(regex)
			}
		}

		fmt.Printf("\n  Filters so far (%d): %s\n\n", len(filters), summariseFilters(filters))

		var addMore bool
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Add another filter?").
				Value(&addMore),
		)).WithShowHelp(false).WithShowErrors(false).Run()
		if err != nil || !addMore {
			break
		}
	}

	return filters
}

// configureCredentialRules interactively builds the connections.credential_rules slice.
// Each rule groups one or more field/regex matchers with a secure-data filename.
// A default catch-all rule (field=name, regex=.*) is always appended at the end.
func configureCredentialRules() []*config_domain.RegexMatchesList {
	var rules []*config_domain.RegexMatchesList

	var addRules bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Configure Credential Rules?").
			Description("Credential rules map data sources to their credentials file.\nA default catch-all rule (name=.*) will always be appended at the end.").
			Value(&addRules),
	)).WithShowHelp(false).WithShowErrors(false).Run()
	if err != nil || !addRules {
		return appendDefaultCredentialRule(rules)
	}

	for {
		// ── Secure data file ──────────────────────────────────────────────
		var secureData string
		err = huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Secure Data File").
				Description("Credentials filename stored in the secure directory (e.g. 'default.yaml', 'prod.yaml').\nLeave blank to use 'default.yaml'.").
				Value(&secureData),
		)).WithShowHelp(false).WithShowErrors(false).Run()
		if err != nil {
			break
		}
		secureData = strings.TrimSpace(secureData)
		if secureData == "" {
			secureData = "default.yaml"
		}

		// ── Matching rules for this credential entry ──────────────────────
		var matchingRules []config_domain.MatchingRule
		for {
			var field, regexVal string
			err = huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Matching Field").
					Description(fmt.Sprintf("Field to match (e.g. 'name', 'url', 'type'). Secure file: '%s'", secureData)).
					Value(&field),
				huh.NewInput().
					Title("Matching Regex").
					Description("Regular expression to match against the field value").
					Value(&regexVal),
			)).WithShowHelp(false).WithShowErrors(false).Run()
			if err != nil {
				break
			}

			field = strings.TrimSpace(field)
			regexVal = strings.TrimSpace(regexVal)
			if field == "" || regexVal == "" {
				fmt.Println("  ⚠  Field and regex are required — skipping.")
			} else if _, compErr := regexp.Compile(regexVal); compErr != nil {
				fmt.Printf("  ⚠  Invalid regex '%s': %v — skipping.\n", regexVal, compErr)
			} else {
				matchingRules = append(matchingRules, config_domain.MatchingRule{
					Field: field,
					Regex: regexVal,
				})
				fmt.Printf("\n  ✓  Added matching rule: field='%s'  regex='%s'\n", field, regexVal)

				// Offer inline regex test
				if shouldTestRegex(fmt.Sprintf("field='%s' regex='%s'", field, regexVal)) {
					runRegexTest(regexVal)
				}
			}

			var addMoreRules bool
			err = huh.NewForm(huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Add another matching rule to this credential entry? (%d so far)", len(matchingRules))).
					Value(&addMoreRules),
			)).WithShowHelp(false).WithShowErrors(false).Run()
			if err != nil || !addMoreRules {
				break
			}
		}

		if len(matchingRules) > 0 {
			rules = append(rules, &config_domain.RegexMatchesList{
				Rules:      matchingRules,
				SecureData: secureData,
			})
			fmt.Printf("\n  ✓  Credential rule added: %d matcher(s) → '%s'\n\n", len(matchingRules), secureData)
		}

		var addMoreCredRules bool
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Add another credential rule? (%d rule(s) configured so far)", len(rules))).
				Value(&addMoreCredRules),
		)).WithShowHelp(false).WithShowErrors(false).Run()
		if err != nil || !addMoreCredRules {
			break
		}
	}

	return appendDefaultCredentialRule(rules)
}

// appendDefaultCredentialRule ensures a catch-all rule (field=name, regex=.*) pointing to
// default.yaml is always the last entry in the credential rules list. If one already exists
// it is a no-op.
func appendDefaultCredentialRule(rules []*config_domain.RegexMatchesList) []*config_domain.RegexMatchesList {
	for _, rule := range rules {
		for _, r := range rule.Rules {
			if r.Field == "name" && r.Regex == ".*" {
				return rules // default already present
			}
		}
	}
	return append(rules, &config_domain.RegexMatchesList{
		Rules: []config_domain.MatchingRule{
			{Field: "name", Regex: ".*"},
		},
		SecureData: "default.yaml",
	})
}

// shouldTestRegex asks whether the user wants to test a given regex inline.
func shouldTestRegex(label string) bool {
	var want bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(fmt.Sprintf("Test regex for %s?", label)).
			Value(&want),
	)).WithShowHelp(false).WithShowErrors(false).Run()
	return err == nil && want
}

// runRegexTest prompts for a test value and prints whether it matches the supplied regex.
func runRegexTest(regex string) {
	var testValue string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Test Value").
			Description(fmt.Sprintf("Enter a value to test against regex '%s'", regex)).
			Value(&testValue),
	)).WithShowHelp(false).WithShowErrors(false).Run()
	if err != nil || strings.TrimSpace(testValue) == "" {
		return
	}
	p, compErr := regexp.Compile(regex)
	if compErr != nil {
		fmt.Printf("  ⚠  Invalid regex '%s': %v\n\n", regex, compErr)
		return
	}
	if p.MatchString(strings.TrimSpace(testValue)) {
		fmt.Printf("  ✓  '%s' MATCHES regex '%s'\n\n", testValue, regex)
	} else {
		fmt.Printf("  ✗  '%s' does NOT match regex '%s'\n\n", testValue, regex)
	}
}

// summariseFilters returns a short human-readable summary of the current filter list.
func summariseFilters(filters []config_domain.MatchingRule) string {
	if len(filters) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(filters))
	for _, f := range filters {
		kind := "excl"
		if f.Inclusive {
			kind = "incl"
		}
		parts = append(parts, fmt.Sprintf("%s/%s(%s)", f.Field, f.Regex, kind))
	}
	return strings.Join(parts, ", ")
}

// writeSecureFileData marshals an object to YAML and writes it to a file with 0600 permissions.
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
func buildFormGroups(authType string, config *config_domain.GrafanaConfig, secureModel *config_domain.SecureModel) []*huh.Group {
	groups := make([]*huh.Group, 0)
	basicGrps := huh.NewGroup(
		huh.NewInput().
			Value(&config.UserName).
			Title("Grafana Username").Description("Grafana Username"),
		huh.NewInput().
			Value(&secureModel.Password).
			Title("Grafana Password").
			Description("Grafana Password").
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
func DeleteContext(app *config_domain.GDGAppConfiguration, name string) {
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
func CopyContext(app *config_domain.GDGAppConfiguration, src, dest string) {
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

// ClearContexts resets all contexts to a single default example context and saves the config.
func ClearContexts(app *config_domain.GDGAppConfiguration) {
	newContext := make(map[string]*config_domain.GrafanaConfig)
	newContext["example"] = config_domain.NewGrafanaConfig(config_domain.WithContextName("example"))
	app.Contexts = newContext
	app.ContextName = "example"
	err := app.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}

	slog.Info("All contexts were cleared")
}
