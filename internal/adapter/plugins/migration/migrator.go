// Package migration provides utilities for re-encrypting GDG's on-disk files
// when switching cipher plugins or disabling encryption entirely.
//
// The Migrator handles three categories of encrypted data:
//  1. Alerting contact point JSON files (read/written via ports.Storage)
//  2. SecureData connection credential files (YAML/JSON maps, read/written via os)
//  3. The per-context auth file holding the Grafana password/token SecureModel
//
// To remove encryption, pass noop.NoOpEncoder{} as NewEncoder.
// To add encryption from plaintext, pass noop.NoOpEncoder{} as OldEncoder.
package migration

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"gopkg.in/yaml.v3"
)

// contactsFile matches the constant used in the alerting_contactpoints.go adapter.
const contactsFile = "contacts"

// Migrator re-encrypts on-disk GDG files from one cipher encoder to another.
// Construct one per context that needs migration; call Rekey to execute.
type Migrator struct {
	// OldEncoder decrypts values as they are currently stored on disk.
	// Use noop.NoOpEncoder{} when data is currently in plaintext.
	OldEncoder ports.CipherEncoder

	// NewEncoder encrypts values for the new storage format.
	// Use noop.NoOpEncoder{} to revert all files to plaintext.
	NewEncoder ports.CipherEncoder

	// GrafanaConf is the active context configuration, used for path resolution.
	GrafanaConf *config_domain.GrafanaConfig

	// Storage is the storage backend; contact point migration is skipped for non-local backends.
	Storage ports.Storage
}

func NewMigrator(oldEncoder ports.CipherEncoder, newEncoder ports.CipherEncoder, grafanaConf *config_domain.GrafanaConfig, storage ports.Storage) *Migrator {
	return &Migrator{
		OldEncoder:  oldEncoder,
		NewEncoder:  newEncoder,
		GrafanaConf: grafanaConf,
		Storage:     storage,
	}
}

// RekeyOptions controls what gets migrated and how.
type RekeyOptions struct {
	// NoBackup disables the creation of backup copies before any file is modified.
	// Defaults to false (backups are created).
	NoBackup bool

	// BackupDir is the directory where backups are written.
	// If empty and NoBackup is false, a timestamped temp directory is created automatically.
	BackupDir string

	// IncludeGdgCredentials controls whether the per-context auth file
	// (holding the Grafana password/token) is also migrated.
	IncludeGdgCredentials bool

	// DryRun, when true, causes Rekey to scan files and populate RekeyReport.Previews
	// without writing any changes to disk. Backups are also skipped during a dry run.
	DryRun bool

	// AllowList restricts processing to the listed absolute file paths.
	// When empty, all discovered files are processed.
	AllowList []string
}

// FilePreview describes one on-disk file that will be processed by a Rekey operation.
// It is populated when RekeyOptions.DryRun is true.
type FilePreview struct {
	// Path is the absolute file path.
	Path string

	// Category identifies the type of file: "contact_points", "secure_data", or "auth".
	Category string

	// Keys lists the map keys whose values will be re-encrypted.
	// Empty for contact_points files (the entire blob is re-encoded as one unit).
	Keys []string

	// DecodedOK is true when OldEncoder successfully decoded every value in the file.
	DecodedOK bool

	// DecodeErr contains the first decode error message when DecodedOK is false.
	DecodeErr string
}

// RekeyReport summarises the results of a Rekey call.
type RekeyReport struct {
	// BackupDir is the directory where backups were written (empty when NoBackup is true).
	BackupDir string

	// ContactPointsFiles lists the contact point file paths that were successfully migrated.
	ContactPointsFiles []string

	// SecureDataFiles lists the SecureData credential file paths that were successfully migrated.
	SecureDataFiles []string

	// GdgCredentialsMigrated reports whether the auth file was successfully migrated.
	GdgCredentialsMigrated bool

	// Errors collects non-fatal errors encountered during migration.
	// When an error is added the affected file is skipped; other files continue.
	Errors []error

	// Previews holds per-file scan results populated during a DryRun.
	Previews []FilePreview
}

// Rekey iterates over all three encrypted data categories and re-encrypts each
// file from OldEncoder to NewEncoder.  It returns an error only for setup
// failures (e.g. cannot create the backup directory); per-file failures are
// accumulated in RekeyReport.Errors so the caller can decide how to proceed.
//
// When opts.DryRun is true no files are modified; instead RekeyReport.Previews
// is populated with one FilePreview per discovered file.
func (m *Migrator) Rekey(opts RekeyOptions) (RekeyReport, error) {
	report := RekeyReport{}

	// Dry runs never write — skip backup creation entirely.
	if opts.DryRun {
		opts.NoBackup = true
	}

	// Create the backup directory once, before any file is touched.
	if !opts.NoBackup {
		if opts.BackupDir == "" {
			timestamp := time.Now().Format("20060102-150405")
			cwd, err := os.Getwd()
			if err != nil {
				return report, fmt.Errorf("get working directory for backup: %w", err)
			}
			dir := filepath.Join(cwd, fmt.Sprintf("gdg-rekey-backup-%s", timestamp))
			if err := os.MkdirAll(dir, 0o750); err != nil {
				return report, fmt.Errorf("create backup directory: %w", err)
			}
			opts.BackupDir = dir
		} else {
			if err := os.MkdirAll(opts.BackupDir, 0o750); err != nil {
				return report, fmt.Errorf("create backup directory: %w", err)
			}
		}
		report.BackupDir = opts.BackupDir
	}

	allowSet := buildAllowSet(opts.AllowList)

	m.rekeyContactPoints(&report, opts, allowSet)
	m.rekeySecureDataFiles(&report, opts, allowSet)
	if opts.IncludeGdgCredentials {
		m.rekeyGdgCredentials(&report, opts, allowSet)
	}

	return report, nil
}

// buildAllowSet converts a slice of paths into a lookup map for O(1) membership
// tests. Returns nil when paths is empty, meaning "allow all files".
func buildAllowSet(paths []string) map[string]bool {
	if len(paths) == 0 {
		return nil
	}
	m := make(map[string]bool, len(paths))
	for _, p := range paths {
		m[p] = true
	}
	return m
}

// isAllowed returns true when allowSet is nil (no restriction) or contains path.
func isAllowed(path string, allowSet map[string]bool) bool {
	return allowSet == nil || allowSet[path]
}

// rekeyContactPoints migrates the alerting contact point file.
// Only local storage is supported; a warning is logged and the step is skipped for cloud backends.
func (m *Migrator) rekeyContactPoints(report *RekeyReport, opts RekeyOptions, allowSet map[string]bool) {
	if m.Storage.Name() != storage.LocalStorageType.String() {
		slog.Warn("contact point re-key only supports local storage; skipping",
			"storage_backend", m.Storage.Name())
		return
	}

	path := resources.BuildResourcePath(m.GrafanaConf, contactsFile, domain.AlertingResource, false, false)

	if !isAllowed(path, allowSet) {
		return
	}

	raw, err := m.Storage.ReadFile(path)
	if err != nil {
		// File may not exist if the user has never downloaded contact points — not an error.
		slog.Debug("contact points: file not found, skipping", "path", path)
		return
	}

	if opts.DryRun {
		preview := FilePreview{Path: path, Category: "contact_points"}
		_, decErr := m.OldEncoder.Decode(domain.AlertingResource, raw)
		if decErr != nil {
			preview.DecodedOK = false
			preview.DecodeErr = decErr.Error()
		} else {
			preview.DecodedOK = true
		}
		report.Previews = append(report.Previews, preview)
		return
	}

	if !opts.NoBackup {
		if err := backupFile(opts.BackupDir, path, raw); err != nil {
			report.Errors = append(report.Errors,
				fmt.Errorf("contact points: backup %s: %w", path, err))
			return
		}
	}

	plaintext, err := m.OldEncoder.Decode(domain.AlertingResource, raw)
	if err != nil {
		report.Errors = append(report.Errors,
			fmt.Errorf("contact points: decode %s: %w", path, err))
		return
	}

	encoded, err := m.NewEncoder.Encode(domain.AlertingResource, plaintext)
	if err != nil {
		report.Errors = append(report.Errors,
			fmt.Errorf("contact points: encode %s: %w", path, err))
		return
	}

	if err := m.Storage.WriteFile(path, encoded); err != nil {
		report.Errors = append(report.Errors,
			fmt.Errorf("contact points: write %s: %w", path, err))
		return
	}

	report.ContactPointsFiles = append(report.ContactPointsFiles, path)
}

// rekeySecureDataFiles migrates all SecureData credential files referenced by the
// active context's connection credential rules.
func (m *Migrator) rekeySecureDataFiles(report *RekeyReport, opts RekeyOptions, allowSet map[string]bool) {
	cs := m.GrafanaConf.GetConnectionSettings()
	if cs == nil {
		return
	}

	seen := make(map[string]bool) // deduplicate in case multiple rules point to the same file

	for _, rule := range cs.MatchingRules {
		if rule.SecureData == "" {
			continue
		}

		path := filepath.Join(m.GrafanaConf.SecureLocation(), rule.SecureData)
		if seen[path] {
			continue
		}
		seen[path] = true

		if !isAllowed(path, allowSet) {
			continue
		}

		raw, err := os.ReadFile(path) // #nosec G304
		if err != nil {
			report.Errors = append(report.Errors,
				fmt.Errorf("secure data: read %s: %w", path, err))
			continue
		}

		if opts.DryRun {
			values, _, parseErr := parseKeyValueFile(path, raw)
			preview := FilePreview{Path: path, Category: "secure_data"}
			if parseErr != nil {
				preview.DecodedOK = false
				preview.DecodeErr = parseErr.Error()
				report.Previews = append(report.Previews, preview)
				continue
			}
			keys := make([]string, 0, len(values))
			allOK := true
			var firstDecErr string
			for k, v := range values {
				keys = append(keys, k)
				if _, decErr := m.OldEncoder.DecodeValue(v); decErr != nil && allOK {
					allOK = false
					firstDecErr = fmt.Sprintf("key %q: %s", k, decErr)
				}
			}
			sort.Strings(keys)
			preview.Keys = keys
			preview.DecodedOK = allOK
			preview.DecodeErr = firstDecErr
			report.Previews = append(report.Previews, preview)
			continue
		}

		if !opts.NoBackup {
			if err := backupFile(opts.BackupDir, path, raw); err != nil {
				report.Errors = append(report.Errors,
					fmt.Errorf("secure data: backup %s: %w", path, err))
				continue
			}
		}

		values, ext, parseErr := parseKeyValueFile(path, raw)
		if parseErr != nil {
			report.Errors = append(report.Errors,
				fmt.Errorf("secure data: parse %s: %w", path, parseErr))
			continue
		}

		for k, v := range values {
			plaintext, decErr := m.OldEncoder.DecodeValue(v)
			if decErr != nil {
				slog.Warn("secure data: could not decode value, skipping key",
					"key", k, "file", path, "err", decErr)
				continue
			}
			newVal, encErr := m.NewEncoder.EncodeValue(plaintext)
			if encErr != nil {
				slog.Warn("secure data: could not re-encode value, keeping original",
					"key", k, "file", path, "err", encErr)
				continue
			}
			values[k] = newVal
		}

		out, marshalErr := marshalKeyValueFile(ext, values)
		if marshalErr != nil {
			report.Errors = append(report.Errors,
				fmt.Errorf("secure data: marshal %s: %w", path, marshalErr))
			continue
		}

		if err := os.WriteFile(path, out, 0o600); err != nil {
			report.Errors = append(report.Errors,
				fmt.Errorf("secure data: write %s: %w", path, err))
			continue
		}

		report.SecureDataFiles = append(report.SecureDataFiles, path)
	}
}

// rekeyGdgCredentials migrates the per-context auth file that holds the
// encrypted Grafana password and/or token.
func (m *Migrator) rekeyGdgCredentials(report *RekeyReport, opts RekeyOptions, allowSet map[string]bool) {
	authBase := m.GrafanaConf.GetAuthLocation()

	// Probe for the file with any supported extension.
	var (
		path string
		raw  []byte
		ext  string
	)
	for _, e := range []string{".yaml", ".yml", ".json"} {
		candidate := authBase + e
		data, err := os.ReadFile(candidate) // #nosec G304
		if err == nil {
			path = candidate
			raw = data
			ext = e
			break
		}
	}

	if path == "" {
		slog.Debug("gdg credentials: no auth file found, skipping", "base", authBase)
		return
	}

	if !isAllowed(path, allowSet) {
		return
	}

	var sm config_domain.SecureModel
	var parseErr error
	switch ext {
	case ".yml", ".yaml":
		parseErr = yaml.Unmarshal(raw, &sm)
	case ".json":
		parseErr = json.Unmarshal(raw, &sm)
	}
	if parseErr != nil {
		report.Errors = append(report.Errors,
			fmt.Errorf("gdg credentials: parse %s: %w", path, parseErr))
		return
	}

	if opts.DryRun {
		preview := FilePreview{Path: path, Category: "auth"}
		allOK := true
		var firstErr string
		if sm.Password != "" {
			preview.Keys = append(preview.Keys, "password")
			if _, decErr := m.OldEncoder.DecodeValue(sm.Password); decErr != nil {
				allOK = false
				firstErr = "password: " + decErr.Error()
			}
		}
		if sm.Token != "" {
			preview.Keys = append(preview.Keys, "token")
			if _, decErr := m.OldEncoder.DecodeValue(sm.Token); decErr != nil && allOK {
				allOK = false
				firstErr = "token: " + decErr.Error()
			}
		}
		preview.DecodedOK = allOK
		preview.DecodeErr = firstErr
		report.Previews = append(report.Previews, preview)
		return
	}

	if !opts.NoBackup {
		if err := backupFile(opts.BackupDir, path, raw); err != nil {
			report.Errors = append(report.Errors,
				fmt.Errorf("gdg credentials: backup %s: %w", path, err))
			return
		}
	}

	// Decode with old encoder then re-encode with new encoder.
	sm.UpdateSecureModel(func(ciphertext string) (string, error) {
		plaintext, err := m.OldEncoder.DecodeValue(ciphertext)
		if err != nil {
			return "", err
		}
		return m.NewEncoder.EncodeValue(plaintext)
	})

	var (
		out      []byte
		writeErr error
	)
	switch ext {
	case ".yml", ".yaml":
		out, writeErr = yaml.Marshal(sm)
	case ".json":
		out, writeErr = json.MarshalIndent(sm, "", "  ")
	}
	if writeErr != nil {
		report.Errors = append(report.Errors,
			fmt.Errorf("gdg credentials: marshal %s: %w", path, writeErr))
		return
	}

	if err := os.WriteFile(path, out, 0o600); err != nil {
		report.Errors = append(report.Errors,
			fmt.Errorf("gdg credentials: write %s: %w", path, err))
		return
	}

	report.GdgCredentialsMigrated = true
}

// backupFile copies srcContent to backupDir, recreating the original path structure
// under backupDir so that files with the same name from different directories
// do not overwrite each other.
func backupFile(backupDir, srcPath string, srcContent []byte) error {
	rel := srcPath
	if filepath.IsAbs(rel) {
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
	}
	dst := filepath.Join(backupDir, rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return fmt.Errorf("create backup directory structure: %w", err)
	}
	return os.WriteFile(dst, srcContent, 0o600)
}

// parseKeyValueFile deserialises a YAML or JSON file into a flat string map.
// It returns the map, the normalised file extension (".yaml", ".yml", or ".json"),
// and any parsing error.
func parseKeyValueFile(path string, raw []byte) (map[string]string, string, error) {
	ext := filepath.Ext(path)
	values := make(map[string]string)
	switch ext {
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(raw, values); err != nil {
			return nil, ext, err
		}
	case ".json":
		if err := json.Unmarshal(raw, &values); err != nil {
			return nil, ext, err
		}
	default:
		return nil, ext, fmt.Errorf("unsupported file extension %q", ext)
	}
	return values, ext, nil
}

// marshalKeyValueFile serialises a flat string map back to the original format.
func marshalKeyValueFile(ext string, values map[string]string) ([]byte, error) {
	switch ext {
	case ".yml", ".yaml":
		return yaml.Marshal(values)
	case ".json":
		return json.MarshalIndent(values, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported file extension %q", ext)
	}
}
