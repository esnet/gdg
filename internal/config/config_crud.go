package config

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/cipher"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
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

// ── Pure-logic helpers (no huh dependency — fully unit-testable) ──────────────

// regexMetaPattern matches the characters that clearly signal regex intent AND
// that encodeFolderName would destroy (e.g. '*' → '%2A', '(' → '%28').
// A plain dot '.' is intentionally excluded because it frequently appears in
// legitimate literal folder names (e.g. "v1.0 Dashboards").
var regexMetaPattern = regexp.MustCompile(`[*?\[\]()|^$\\]`)

// looksLikeRegex reports whether s contains characters that strongly suggest
// the user intends it as a regex pattern rather than a literal folder name.
// This is used to decide whether to prompt the user before encoding.
func looksLikeRegex(s string) bool {
	return regexMetaPattern.MatchString(s)
}

// encodeFolderName applies URL-encoding and regex-escaping to a single folder name.
// Spaces, slashes, and other special characters are encoded so that the value can
// safely be stored in the watched-folders list and later used as a regexp pattern.
func encodeFolderName(name string) string {
	return encode.EncodePath(encode.EncodeEscapeSpecialChars, strings.TrimSpace(name))
}

// testFolderRegexMatch URL-encodes rawValue (without regex-escaping) and then
// tests it against every pattern in folders.  Stored patterns are already
// regex-escaped (via encodeFolderName / EncodeEscapeSpecialChars), so the
// test value only needs the URL-encoding pass; applying QuoteMeta on top would
// produce a double-escaped string that would never match.
// Returns (anyMatch, slice-of-matching-patterns).
func testFolderRegexMatch(folders []string, rawValue string) (bool, []string) {
	// URL-encode only (no regexp.QuoteMeta) so that e.g. "Linux Data" becomes
	// "Linux+Data", which the stored pattern "Linux\+Data" correctly matches.
	encoded := encode.EncodePath(encode.Encode, strings.TrimSpace(rawValue))
	var matches []string
	for _, f := range folders {
		p, compErr := regexp.Compile(f)
		if compErr != nil {
			continue // bad pattern; silently skip
		}
		if p.MatchString(encoded) {
			matches = append(matches, f)
		}
	}
	return len(matches) > 0, matches
}

// validateFilter validates field + regex and, when both are well-formed, returns
// the corresponding MatchingRule. An error is returned if either value is blank
// or if regex does not compile.
func validateFilter(field, regex string, inclusive bool) (*config_domain.MatchingRule, error) {
	field = strings.TrimSpace(field)
	regex = strings.TrimSpace(regex)
	if field == "" || regex == "" {
		return nil, errors.New("field and regex are required")
	}
	if _, err := regexp.Compile(regex); err != nil {
		return nil, fmt.Errorf("invalid regex %q: %w", regex, err)
	}
	return &config_domain.MatchingRule{Field: field, Regex: regex, Inclusive: inclusive}, nil
}

// validateCredentialRule validates field + regex for use in a credential matching
// rule (inclusive is always false for credential rules).
func validateCredentialRule(field, regex string) (*config_domain.MatchingRule, error) {
	return validateFilter(field, regex, false)
}

// testRegexMatch compiles regex and tests whether value matches.
// Returns (matched, nil) on success, or (false, error) when the regex is invalid.
func testRegexMatch(regex, value string) (bool, error) {
	p, err := regexp.Compile(regex)
	if err != nil {
		return false, fmt.Errorf("invalid regex %q: %w", regex, err)
	}
	return p.MatchString(strings.TrimSpace(value)), nil
}

// validateSecureDataFile checks that a secure data filename is non-blank and
// is not the reserved name "default.yaml".
func validateSecureDataFile(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("filename is required (default.yaml is reserved)")
	}
	if s == "default.yaml" {
		return errors.New("default.yaml is reserved for default connection credentials")
	}
	return nil
}

// appendDefaultCredentialRule ensures a catch-all rule (field=name, regex=.*)
// pointing to default.yaml is always the last entry in the credential rules list.
// If an equivalent rule already exists the input is returned unchanged.
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

// validateGrafanaURL checks that s is a syntactically valid http/https URL with a
// non-empty host.  It is used as a huh.Input Validate callback so errors are shown
// inline in the TUI; no live reachability probe is performed.
func validateGrafanaURL(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("URL is required")
	}
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL must start with http:// or https://")
	}
	if u.Host == "" {
		return errors.New("URL must include a host (e.g. http://grafana.example.com)")
	}
	return nil
}

// summariseFilters returns a compact human-readable description of a filter list,
// used for display purposes between interactive prompts.
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

// ── TUI functions ─────────────────────────────────────────────────────────────

// CreateNewContext launches the 3-panel config builder TUI to interactively
// configure a new Grafana context. On successful completion it writes secure
// files, updates the context map, and saves the config to disk.
func CreateNewContext(app *config_domain.GDGAppConfiguration, name string) {
	var encoder ports.CipherEncoder
	if !app.PluginConfig.Disabled && app.PluginConfig.CipherPlugin != nil {
		var encErr error
		encoder, encErr = cipher.NewPluginCipherEncoder(app.PluginConfig.CipherPlugin, app.SecureConfig)
		if encErr != nil {
			log.Fatalf("Failed to load cipher plugin: %v", encErr)
		}
	} else {
		encoder = noop.NoOpEncoder{}
	}

	model := newConfigBuilderModel(app, name, encoder)
	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		log.Fatalf("TUI error: %v", err)
	}

	final := result.(configBuilderModel)
	if final.cancelled || !final.done {
		slog.Info("Configuration cancelled — no changes were made.")
		return
	}

	bs := final.bs
	newConfig := bs.config
	secure := bs.secure

	// ── Write default connection credentials ──────────────────────────────
	if bs.configureConnections {
		const passKey = "basicAuthPassword"
		defaultDs := config_domain.GrafanaConnection{
			"user":  bs.connectionUser,
			passKey: bs.connectionPassword,
		}

		securePath := resourceTypes.SecureSecretsResource
		location := filepath.Join(newConfig.OutputPath, string(securePath))
		if mkErr := os.MkdirAll(location, 0o750); mkErr != nil {
			log.Fatalf("unable to create default secret location.  location: %s, %v", location, mkErr)
		}

		secretFileLocation := filepath.Join(location, "default.yaml")
		if encoder != nil {
			newVal, encodeErr := encoder.EncodeValue(defaultDs.Password())
			if encodeErr == nil {
				defaultDs[passKey] = newVal
			}
		}

		if writeErr := writeSecureFileData(defaultDs, secretFileLocation); writeErr != nil {
			log.Fatalf("unable to write secret default file.  location: %s, %v", secretFileLocation, writeErr)
		}

		// Write credential rule files
		for _, cred := range bs.pendingCreds {
			ds := config_domain.GrafanaConnection{
				"user":  cred.user,
				passKey: cred.password,
			}
			if encoder != nil {
				newVal, encodeErr := encoder.EncodeValue(ds.Password())
				if encodeErr == nil {
					ds[passKey] = newVal
				}
			}
			credFilePath := filepath.Join(location, cred.secureData)
			if writeErr := writeSecureFileData(ds, credFilePath); writeErr != nil {
				log.Fatalf("unable to write credential file.  location: %s, %v", credFilePath, writeErr)
			}
			slog.Info("Credential file created", "file", credFilePath)
		}
	}

	// ── Write auth credentials file ───────────────────────────────────────
	authFileLocation := fmt.Sprintf("%s.yaml", newConfig.GetAuthLocation())
	secure.UpdateSecureModel(encoder.EncodeValue)

	if writeErr := writeSecureFileData(*secure, authFileLocation); writeErr != nil {
		log.Fatalf("unable to write secret auth file.  location: %s, %v", authFileLocation, writeErr)
	}

	// ── Write cloud storage config + credentials ──────────────────────────
	if bs.configureStorage && cloudProvider(bs.storageProvider) == providerCustom {
		if app.StorageEngine == nil {
			app.StorageEngine = make(map[string]map[string]string)
		}

		encodedSecret, encErr := encoder.EncodeValue(bs.storageSecretKey)
		if encErr != nil {
			log.Fatalf("failed to encode storage secret key: %v", encErr)
		}

		securePath := newConfig.SecureLocation()
		if mkErr := os.MkdirAll(securePath, 0o750); mkErr != nil {
			log.Fatalf("unable to create secure directory %s: %v", securePath, mkErr)
		}

		credFile := filepath.Join(
			securePath,
			fmt.Sprintf("%s_%s.yaml", config_domain.CloudAuthPrefix, bs.storageLabel),
		)
		creds := map[string]string{
			storage.AccessId:  bs.storageAccessID,
			storage.SecretKey: encodedSecret,
		}
		if writeErr := writeSecureFileData(creds, credFile); writeErr != nil {
			log.Fatalf("unable to write storage credentials file %s: %v", credFile, writeErr)
		}

		app.StorageEngine[bs.storageLabel] = map[string]string{
			cloudKindKey:       cloudKindCloud,
			storage.CloudType:  storage.Custom,
			storage.Endpoint:   bs.storageEndpoint,
			storage.BucketName: bs.storageBucket,
			storage.Region:     bs.storageRegion,
			storage.Prefix:     bs.storagePrefix,
			storage.InitBucket: fmt.Sprintf("%t", bs.storageInitBucket),
			sslEnabledKey:      fmt.Sprintf("%t", bs.storageSSL),
		}
	}

	// ── Update config and save ────────────────────────────────────────────
	contextMap := app.GetContexts()
	contextMap[name] = newConfig
	app.ContextName = name

	if saveErr := app.SaveToDisk(false); saveErr != nil {
		log.Fatal("could not save configuration.")
	}

	slog.Info("New configuration has been created", "newContext", name)
}

// ── Auth form helpers ─────────────────────────────────────────────────────────

// writeSecureFileData marshals an object to YAML and writes it to a file with 0600 permissions.
func writeSecureFileData[T any](object T, location string) error {
	data, err := yaml.Marshal(&object)
	if err != nil {
		log.Fatalf("unable to turn map into yaml representation.  location: %s, %v", location, err)
	}
	err = os.WriteFile(location, data, 0o600)
	return err
}

// ── Context management (no TUI) ───────────────────────────────────────────────

// DeleteContext removes a given context and its associated credential files.
// When skipConfirmation is true, all credential files (except default.yaml) are
// deleted without prompting. Otherwise the user is asked to confirm deletion of
// each file.
func DeleteContext(app *config_domain.GDGAppConfiguration, name string, skipConfirmation bool) {
	name = strings.ToLower(name) // ensure name is lower case
	contexts := app.GetContexts()
	ctx, ok := contexts[name]
	if !ok {
		log.Fatalf("Context not found, cannot delete context: %s", name)
		return
	}

	secureLoc := ctx.SecureLocation()

	// Collect credential files to delete (auth file + connection credential files).
	var filesToDelete []string

	authFile := filepath.Join(secureLoc, fmt.Sprintf("auth_%s.yaml", name))
	if _, statErr := os.Stat(authFile); statErr == nil {
		filesToDelete = append(filesToDelete, authFile)
	}

	if ctx.ConnectionSettings != nil {
		for _, rule := range ctx.ConnectionSettings.MatchingRules {
			if rule.SecureData == "" || rule.SecureData == "default.yaml" {
				continue
			}
			credFile := filepath.Join(secureLoc, rule.SecureData)
			if _, statErr := os.Stat(credFile); statErr == nil {
				filesToDelete = append(filesToDelete, credFile)
			}
		}
	}

	if len(filesToDelete) > 0 {
		if skipConfirmation {
			for _, f := range filesToDelete {
				if removeErr := os.Remove(f); removeErr != nil {
					slog.Warn("failed to remove credential file", "file", f)
				} else {
					slog.Info("Removed credential file", "file", f)
				}
			}
		} else {
			fmt.Println("\n  The following credential files are associated with this context:")
			for _, f := range filesToDelete {
				fmt.Printf("    - %s\n", f)
			}
			fmt.Println()

			var confirmDelete bool
			err := huh.NewForm(huh.NewGroup(
				huh.NewConfirm().
					Title("Delete these credential files?").
					Description("Select No to keep the files on disk.").
					Value(&confirmDelete),
			)).WithShowHelp(false).WithShowErrors(false).Run()
			if err == nil && confirmDelete {
				for _, f := range filesToDelete {
					if removeErr := os.Remove(f); removeErr != nil {
						slog.Warn("failed to remove credential file", "file", f)
					} else {
						slog.Info("Removed credential file", "file", f)
					}
				}
			} else {
				slog.Info("Credential files were kept on disk")
			}
		}
	}

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
