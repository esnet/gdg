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
	"strings"
	"time"

	grafanaapi "github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/adapter/storage"
	config_domain "github.com/esnet/gdg/internal/config/config_domain"
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
}

// Rekey iterates over all three encrypted data categories and re-encrypts each
// file from OldEncoder to NewEncoder.  It returns an error only for setup
// failures (e.g. cannot create the backup directory); per-file failures are
// accumulated in RekeyReport.Errors so the caller can decide how to proceed.
func (m *Migrator) Rekey(opts RekeyOptions) (RekeyReport, error) {
	report := RekeyReport{}

	// Create the backup directory once, before any file is touched.
	if !opts.NoBackup {
		if opts.BackupDir == "" {
			timestamp := time.Now().Format("20060102-150405")
			dir, err := os.MkdirTemp("", fmt.Sprintf("gdg-rekey-backup-%s-*", timestamp))
			if err != nil {
				return report, fmt.Errorf("create backup directory: %w", err)
			}
			opts.BackupDir = dir
		}
		report.BackupDir = opts.BackupDir
	}

	m.rekeyContactPoints(&report, opts)
	m.rekeySecureDataFiles(&report, opts)
	if opts.IncludeGdgCredentials {
		m.rekeyGdgCredentials(&report, opts)
	}

	return report, nil
}

// rekeyContactPoints migrates the alerting contact point file.
// Only local storage is supported; a warning is logged and the step is skipped for cloud backends.
func (m *Migrator) rekeyContactPoints(report *RekeyReport, opts RekeyOptions) {
	if m.Storage.Name() != storage.LocalStorageType.String() {
		slog.Warn("contact point re-key only supports local storage; skipping",
			"storage_backend", m.Storage.Name())
		return
	}

	path := grafanaapi.BuildResourcePath(m.GrafanaConf, contactsFile, domain.AlertingResource, false, false)

	raw, err := m.Storage.ReadFile(path)
	if err != nil {
		// File may not exist if the user has never downloaded contact points — not an error.
		slog.Debug("contact points: file not found, skipping", "path", path)
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
func (m *Migrator) rekeySecureDataFiles(report *RekeyReport, opts RekeyOptions) {
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

		raw, err := os.ReadFile(path) // #nosec G304
		if err != nil {
			report.Errors = append(report.Errors,
				fmt.Errorf("secure data: read %s: %w", path, err))
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
func (m *Migrator) rekeyGdgCredentials(report *RekeyReport, opts RekeyOptions) {
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

	if !opts.NoBackup {
		if err := backupFile(opts.BackupDir, path, raw); err != nil {
			report.Errors = append(report.Errors,
				fmt.Errorf("gdg credentials: backup %s: %w", path, err))
			return
		}
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
