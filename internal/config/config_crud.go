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
	"github.com/esnet/gdg/internal/adapter/plugins/registry"
	"github.com/esnet/gdg/internal/tui"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/cipher"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config/config_domain"
	resourceTypes "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
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
// registryClient is used to populate the optional cipher-plugin phases; pass a
// default-constructed client when no override is needed.
func CreateNewContext(app *config_domain.GDGAppConfiguration, name string, registryClient *registry.Client) {
	// Normalise to lowercase — consistent with DeleteContext, SetContext, and
	// Viper's own key normalisation on reload.
	name = strings.ToLower(name)

	// Snapshot the plugin state BEFORE any modifications.
	// pluginWasDisabled records whether encryption was disabled at entry; when
	// true the on-disk files are plaintext and the old cipher must NOT be used
	// as the decoder — we use NoOpEncoder instead.
	oldPlugin := app.PluginConfig.CipherPlugin
	pluginWasDisabled := app.PluginConfig.Disabled

	// Capture the set of existing context names BEFORE adding the new one.
	// These are the contexts that already have on-disk files and are therefore
	// the candidates for migration in the rekey TUI.
	var existingContextNames []string
	for ctxName := range app.GetContexts() {
		existingContextNames = append(existingContextNames, ctxName)
	}

	var encoder outbound.CipherEncoder
	if !app.PluginConfig.Disabled && app.PluginConfig.CipherPlugin != nil {
		var encErr error
		encoder, encErr = cipher.NewPluginCipherEncoder(app.PluginConfig.CipherPlugin, app.SecureConfig)
		if encErr != nil {
			log.Fatalf("Failed to load cipher plugin: %v", encErr)
		}
	} else {
		encoder = noop.NoOpEncoder{}
	}

	model := newConfigBuilderModel(app, name, encoder, registryClient)
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

	// ── Apply cipher plugin chosen in TUI ────────────────────────────────
	if bs.pluginResult != nil {
		newEnc, encErr := cipher.NewPluginCipherEncoder(bs.pluginResult, app.SecureConfig)
		if encErr != nil {
			log.Fatalf("Failed to initialise cipher plugin from TUI selection: %v", encErr)
		}
		encoder = newEnc
		app.PluginConfig.CipherPlugin = bs.pluginResult
		app.PluginConfig.Disabled = false
	}

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
	if contextMap == nil {
		contextMap = make(map[string]*config_domain.GrafanaConfig)
		app.Contexts = contextMap
	}
	contextMap[name] = newConfig
	app.ContextName = name

	if saveErr := app.SaveToDisk(false); saveErr != nil {
		log.Fatal("could not save configuration.")
	}

	slog.Info("New configuration has been created", "newContext", name)

	// ── Offer rekey when the user configured a new cipher plugin ─────────────
	// Launch the rekey TUI whenever the user picks a cipher plugin so they can
	// migrate any existing on-disk files to the new encoding.
	//
	// Key subtlety around the "old" encoder:
	//   • If encryption was DISABLED at entry (pluginWasDisabled == true), all
	//     existing files are stored as plaintext. We pass effectiveOldPlugin=nil
	//     so RunRekeyWithPlugin uses NoOpEncoder as the decoder — attempting to
	//     load and use the configured-but-disabled WASM would fail or produce
	//     wrong results.
	//   • If a cipher was ACTIVE (pluginWasDisabled == false && oldPlugin != nil),
	//     files are encrypted with it; pass it so the decoder is correct.
	//   • If no cipher was ever configured (oldPlugin == nil), files are also
	//     plaintext, so effectiveOldPlugin stays nil.
	if bs.pluginResult != nil && len(existingContextNames) > 0 {
		slog.Info("Cipher plugin configured — launching rekey TUI to migrate existing files")
		var effectiveOldPlugin *config_domain.PluginEntity
		if !pluginWasDisabled && oldPlugin != nil {
			effectiveOldPlugin = oldPlugin
		}
		if rekeyErr := RunRekeyWithPlugin(app, registryClient, effectiveOldPlugin, bs.pluginResult, existingContextNames); rekeyErr != nil {
			slog.Warn("Rekey aborted or encountered errors", "err", rekeyErr)
		}
	}
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
// each file. If the context references a storage engine that is no longer used
// by any other context, the user is also prompted to delete that engine and its
// credentials file.
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

			confirmDelete := tui.RunConfirm(
				"Delete these credential files?",
				"Select No to keep the files on disk.",
			)
			if confirmDelete {
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

	// ── Orphaned storage engine check ────────────────────────────────────────
	// If this context owns a storage engine that no other context references,
	// offer to remove it from gdg.yml and delete its credentials file.
	if storageName := ctx.Storage; storageName != "" {
		inUse := false
		for ctxName, otherCtx := range contexts {
			if ctxName == name {
				continue
			}
			if otherCtx.Storage == storageName {
				inUse = true
				break
			}
		}

		if !inUse {
			// Credentials file written alongside the context (CreateNewContext path).
			s3CredFile := filepath.Join(secureLoc, fmt.Sprintf("%s_%s.yaml", config_domain.CloudAuthPrefix, storageName))

			doDeleteEngine := skipConfirmation
			if !skipConfirmation {
				fmt.Printf("\n  Storage engine %q is only used by context %q.\n", storageName, name)
				if _, statErr := os.Stat(s3CredFile); statErr == nil {
					fmt.Printf("  Credentials file: %s\n", s3CredFile)
				}
				fmt.Println()

				doDeleteEngine = tui.RunConfirm(
					fmt.Sprintf("Delete storage engine %q and its credentials?", storageName),
					"No other contexts are using this storage engine. Select No to keep it.",
				)
			}

			if doDeleteEngine {
				// Case-insensitive scan: Viper normalises map keys to lowercase on
				// load, so ctx.Storage (a struct-field value, not lowercased) may
				// differ in case from the actual key stored in app.StorageEngine.
				engineKey := ""
				for k := range app.StorageEngine {
					if strings.EqualFold(k, storageName) {
						engineKey = k
						break
					}
				}
				if engineKey != "" {
					delete(app.StorageEngine, engineKey)
					// Save immediately after the storage engine deletion — mirrors the
					// working DeleteS3Config pattern. A single deferred save at the end
					// is unreliable when the context deletion also modifies app.Contexts
					// in the same yaml.Marshal call.
					if saveErr := app.SaveToDisk(false); saveErr != nil {
						log.Fatal("failed to save configuration after storage engine deletion")
					}
					slog.Info("storage engine removed from config", "engine", engineKey)
				} else {
					slog.Warn("storage engine not found in config — skipping map deletion", "engine", storageName)
				}
				// Remove the credentials file if present.
				if _, statErr := os.Stat(s3CredFile); statErr == nil {
					if removeErr := os.Remove(s3CredFile); removeErr != nil {
						slog.Warn("failed to remove storage credentials file", "file", s3CredFile)
					} else {
						slog.Info("storage credentials file removed", "file", s3CredFile)
					}
				}
			} else {
				slog.Info("storage engine kept", "engine", storageName)
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
