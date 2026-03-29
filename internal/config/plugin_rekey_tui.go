// Package config provides interactive configuration tools for GDG, including
// the TUI-driven plugin rekey workflow.
package config

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/ports/outbound"

	"github.com/esnet/gdg/internal/adapter/plugins/migration"
	"github.com/esnet/gdg/internal/adapter/plugins/registry"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/cipher"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
)

// ── Rekey TUI phases ──────────────────────────────────────────────────────────

type rekeyPhase int

const (
	phaseRekeyCurrentInfo   rekeyPhase = iota // Note: display current plugin config
	phaseRekeyAction                          // Select: switch / disable / cancel
	phaseRekeyPluginSelect                    // (switch only) Select plugin from registry
	phaseRekeyPluginVersion                   // (switch only) Select version
	phaseRekeyPluginConfig                    // (switch only, loops) Input each config field
	phaseRekeyContextSelect                   // Select which context to migrate
	phaseRekeyBackupOptions                   // Backup yes/no + optional backup directory
	phaseRekeyCredentials                     // Include per-context auth file?
	phaseRekeyDryRun                          // Scan files; results shown as a Note
	phaseRekeyFileSelect                      // MultiSelect: choose which files to process
	phaseRekeyConfirm                         // Final confirmation before writing
	phaseRekeyDone                            // Sentinel — triggers program exit
)

func (p rekeyPhase) sectionName() string {
	switch p {
	case phaseRekeyCurrentInfo:
		return "Current Configuration"
	case phaseRekeyAction:
		return "Choose Action"
	case phaseRekeyPluginSelect, phaseRekeyPluginVersion, phaseRekeyPluginConfig:
		return "Plugin Selection"
	case phaseRekeyContextSelect:
		return "Context Selection"
	case phaseRekeyBackupOptions:
		return "Backup Options"
	case phaseRekeyCredentials:
		return "Credential Migration"
	case phaseRekeyDryRun:
		return "File Scan"
	case phaseRekeyFileSelect:
		return "Select Files"
	case phaseRekeyConfirm:
		return "Confirm"
	default:
		return ""
	}
}

// ── Rekey state ───────────────────────────────────────────────────────────────

type rekeyState struct {
	app        *config_domain.GDGAppConfiguration
	regClient  *registry.Client
	oldEncoder outbound.CipherEncoder

	// Action chosen by the user.
	action string // "switch" | "disable" | "cancel"

	// Plugin selection (action == "switch").
	availablePlugins     []domain.PluginRegistryEntry
	registryError        string
	pluginName           string
	pluginVersion        string
	resolvedEntry        *domain.PluginRegistryEntry
	resolvedVersionEntry *domain.PluginVersionEntry
	pluginConfigFields   []string
	configFieldIdx       int
	configFieldValue     string
	pluginConfigValues   map[string]string
	newPluginEntity      *config_domain.PluginEntity

	// Context selection.
	contextName string

	// Backup options.
	doBackup  bool
	backupDir string

	// Credential migration.
	includeGdgCredentials bool

	// Dry-run scan results.
	previews   []migration.FilePreview
	scanErrors []error

	// File selection — paths chosen by the user.
	selectedPaths []string

	// Final confirmation.
	confirmed bool
}

// buildNewPluginEntity assembles a PluginEntity from the collected config values
// and the resolved registry entry. Must be called after resolvedEntry and
// resolvedVersionEntry are set and all config field values have been collected.
func (rs *rekeyState) buildNewPluginEntity() {
	if rs.resolvedEntry == nil || rs.resolvedVersionEntry == nil {
		return
	}
	wasmURL := rs.resolvedEntry.ResolveURL(rs.resolvedVersionEntry.Version)
	cfg := make(map[string]string, len(rs.pluginConfigValues))
	for k, v := range rs.pluginConfigValues {
		cfg[k] = v
	}
	rs.newPluginEntity = &config_domain.PluginEntity{
		Url:          wasmURL,
		PluginConfig: cfg,
	}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type pluginRekeyModel struct {
	phase     rekeyPhase
	rs        *rekeyState
	form      *huh.Form
	width     int
	height    int
	done      bool
	cancelled bool
}

func newPluginRekeyModel(rs *rekeyState) pluginRekeyModel {
	m := pluginRekeyModel{
		phase:  phaseRekeyCurrentInfo,
		rs:     rs,
		width:  100,
		height: 30,
	}
	m.form = m.buildForm()
	return m
}

// ── tea.Model interface ───────────────────────────────────────────────────────

func (m pluginRekeyModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m pluginRekeyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.form.WithWidth(m.width - 8)
		m.form.WithHeight(m.bodyHeight())
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
	}

	newModel, cmd := m.form.Update(msg)
	m.form = newModel.(*huh.Form)

	if m.form.State == huh.StateCompleted {
		m.applyPhase()
		m.phase = m.nextPhase()
		if m.phase == phaseRekeyDone {
			m.done = true
			return m, tea.Quit
		}
		m.form = m.buildForm()
		m.form.WithWidth(m.width - 8)
		m.form.WithHeight(m.bodyHeight())
		return m, m.form.Init()
	}

	if m.form.State == huh.StateAborted {
		m.cancelled = true
		return m, tea.Quit
	}

	return m, cmd
}

func (m pluginRekeyModel) View() tea.View {
	if m.done || m.cancelled {
		return tea.NewView("")
	}
	v := tea.NewView("")

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("202")).
		Align(lipgloss.Center).
		Width(m.width).
		Padding(0, 1)
	header := headerStyle.Render("GDG Plugin Re-key")

	stepStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Foreground(lipgloss.Color("245"))
	stepInfo := stepStyle.Render(m.phase.sectionName())

	formView := m.form.View()
	formPanel := lipgloss.NewStyle().
		Width(m.width).
		Height(m.bodyHeight()).
		Padding(1, 4).
		Render(formView)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(m.width).
		Padding(0, 2).
		Render("ctrl+c: cancel  •  enter: submit  •  ↑/↓: navigate")

	v.Content = lipgloss.JoinVertical(lipgloss.Left, header, stepInfo, formPanel, footer)
	v.AltScreen = true
	return v
}

// ── Layout helpers ────────────────────────────────────────────────────────────

func (m pluginRekeyModel) bodyHeight() int {
	h := m.height - tuiHeaderHeight - tuiFooterHeight
	if h < tuiMinBodyH {
		h = tuiMinBodyH
	}
	return h
}

// ── Form creation per phase ───────────────────────────────────────────────────

func (m *pluginRekeyModel) buildForm() *huh.Form {
	switch m.phase {

	case phaseRekeyCurrentInfo:
		desc := currentPluginDescription(m.rs.app)
		return huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Current Cipher Plugin").
					Description(desc).
					Next(true).
					NextLabel("Continue →"),
			),
		).WithShowHelp(false)

	case phaseRekeyAction:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to do?").
					Options(
						huh.NewOption("Switch to a different cipher plugin", "switch"),
						huh.NewOption("Disable encryption (revert files to plaintext)", "disable"),
						huh.NewOption("Cancel — make no changes", "cancel"),
					).
					Value(&m.rs.action),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseRekeyPluginSelect:
		// Lazy-fetch the plugin registry on first visit.
		if m.rs.regClient == nil && m.rs.registryError == "" {
			m.rs.registryError = "No registry client configured. Cannot select a plugin without registry access."
		}
		if m.rs.availablePlugins == nil && m.rs.registryError == "" {
			plugins, err := m.rs.regClient.CipherPlugins()
			if err != nil {
				m.rs.registryError = fmt.Sprintf("Could not load plugin registry: %s", err)
			} else if len(plugins) == 0 {
				m.rs.registryError = "No cipher plugins found in the registry."
			} else {
				m.rs.availablePlugins = plugins
			}
		}
		if m.rs.registryError != "" {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewNote().
						Title("Registry Unavailable").
						Description(m.rs.registryError + "\n\nPress enter to go back and choose a different action.").
						Next(true).
						NextLabel("Go Back"),
				),
			).WithShowHelp(false)
		}
		opts := make([]huh.Option[string], len(m.rs.availablePlugins))
		for i, p := range m.rs.availablePlugins {
			label := p.Name
			if p.Description != "" {
				label += " — " + p.Description
			}
			opts[i] = huh.NewOption(label, p.Name)
		}
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Cipher Plugin").
					Description("Choose a plugin from the GDG plugin registry.").
					Options(opts...).
					Value(&m.rs.pluginName),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseRekeyPluginVersion:
		// Find the selected plugin entry.
		var entry *domain.PluginRegistryEntry
		for i := range m.rs.availablePlugins {
			if m.rs.availablePlugins[i].Name == m.rs.pluginName {
				entry = &m.rs.availablePlugins[i]
				break
			}
		}
		if entry == nil || len(entry.Versions) == 0 {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewNote().
						Title("No Versions").
						Description(fmt.Sprintf("No versions found for plugin %q.", m.rs.pluginName)).
						Next(true).
						NextLabel("Go Back"),
				),
			).WithShowHelp(false)
		}
		opts := make([]huh.Option[string], len(entry.Versions))
		for i, v := range entry.Versions {
			label := v.Version
			if i == len(entry.Versions)-1 {
				label += " (latest)"
			}
			opts[i] = huh.NewOption(label, v.Version)
		}
		// Pre-select the latest version.
		if m.rs.pluginVersion == "" {
			m.rs.pluginVersion = entry.Versions[len(entry.Versions)-1].Version
		}
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("Select Version for %q", m.rs.pluginName)).
					Options(opts...).
					Value(&m.rs.pluginVersion),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseRekeyPluginConfig:
		if m.rs.configFieldIdx >= len(m.rs.pluginConfigFields) {
			return huh.NewForm(huh.NewGroup())
		}
		field := m.rs.pluginConfigFields[m.rs.configFieldIdx]
		total := len(m.rs.pluginConfigFields)
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Config Field: %q (%d of %d)", field, m.rs.configFieldIdx+1, total)).
					Description(
						"Enter a value, or use env:VAR_NAME to read from an environment variable,\n" +
							"or file:/path/to/file to read from a file.\n" +
							"Raw values are stored in gdg.yml — use env: or file: for secrets.",
					).
					Value(&m.rs.configFieldValue),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseRekeyContextSelect:
		contextNames := make([]string, 0, len(m.rs.app.GetContexts()))
		for name := range m.rs.app.GetContexts() {
			contextNames = append(contextNames, name)
		}
		sort.Strings(contextNames)
		opts := make([]huh.Option[string], len(contextNames))
		for i, n := range contextNames {
			opts[i] = huh.NewOption(n, n)
		}
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Context to Migrate").
					Description("Choose which Grafana context's encrypted files will be re-keyed.").
					Options(opts...).
					Value(&m.rs.contextName),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseRekeyBackupOptions:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Create Backups Before Modifying Files?").
					Description("Recommended. Original files are mirrored under the backup directory.").
					Value(&m.rs.doBackup),
				huh.NewInput().
					Title("Custom Backup Directory (optional)").
					Description("Leave blank to use an auto-generated timestamped temporary directory.\nOnly used when backups are enabled.").
					Value(&m.rs.backupDir),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseRekeyCredentials:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Also Migrate Grafana Login Credentials?").
					Description("Includes the per-context auth file holding the encrypted Grafana password and/or token.").
					Value(&m.rs.includeGdgCredentials),
			),
		).WithShowHelp(false)

	case phaseRekeyDryRun:
		// Run the scan synchronously and populate rs.previews / rs.scanErrors.
		// This is intentionally done inside buildForm so results are ready for the
		// subsequent phaseRekeyFileSelect form.
		scanReport := m.runDryScan()
		m.rs.previews = scanReport.Previews
		m.rs.scanErrors = scanReport.Errors
		fmtErrHandler := func(err error) {
			if err != nil {
				slog.Warn("Failed to update string builder", "err", err)
			}
		}

		// Build the summary text.
		var sb strings.Builder
		if len(scanReport.Previews) == 0 {
			sb.WriteString("No encrypted files found for the selected context.")
		} else {
			_, fmtErr := fmt.Fprintf(&sb, "Found %d file(s) that would be processed:\n\n", len(scanReport.Previews))
			fmtErrHandler(fmtErr)
			for _, p := range scanReport.Previews {
				statusMark := "✓"
				if !p.DecodedOK {
					statusMark = "⚠"
				}
				label := filepath.Base(p.Path)
				_, fmtErr = fmt.Fprintf(&sb, "  %s [%s] %s\n", statusMark, p.Category, label)
				fmtErrHandler(fmtErr)
				_, fmtErr = fmt.Fprintf(&sb, "      %s\n", p.Path)
				fmtErrHandler(fmtErr)
				if !p.DecodedOK {
					_, fmtErr = fmt.Fprintf(&sb, "      Error: %s\n", p.DecodeErr)
					fmtErrHandler(fmtErr)
				}
				sb.WriteString("\n")
			}
		}
		if len(scanReport.Errors) > 0 {
			sb.WriteString("Scan errors:\n")
			for _, e := range scanReport.Errors {
				_, fmtErr := fmt.Fprintf(&sb, "  • %s\n", e)
				fmtErrHandler(fmtErr)
			}
		}
		return huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Scan Results").
					Description(sb.String()).
					Next(true).
					NextLabel("Continue →"),
			),
		).WithShowHelp(false)

	case phaseRekeyFileSelect:
		if len(m.rs.previews) == 0 {
			return huh.NewForm(
				huh.NewGroup(
					huh.NewNote().
						Title("No Files Found").
						Description("No files were found to process for the selected context.").
						Next(true).
						NextLabel("Continue →"),
				),
			).WithShowHelp(false)
		}
		// Reset selectedPaths so huh drives selection state from the option-level
		// Selected flags, not from any pre-populated slice.
		m.rs.selectedPaths = nil
		opts := make([]huh.Option[string], len(m.rs.previews))
		for i, p := range m.rs.previews {
			label := fmt.Sprintf("[%s] %s", p.Category, filepath.Base(p.Path))
			if !p.DecodedOK {
				label += " ⚠ (decode error)"
			}
			// Pre-select only files that decoded successfully.
			opts[i] = huh.NewOption(label, p.Path).Selected(p.DecodedOK)
		}
		return huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select Files to Re-Encrypt").
					Description("Deselect files to skip them. Files marked ⚠ failed to decode with the current plugin.").
					Options(opts...).
					Value(&m.rs.selectedPaths),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseRekeyConfirm:
		count := len(m.rs.selectedPaths)
		var desc string
		if count == 0 {
			desc = "No files are selected — no changes will be made."
		} else {
			backupInfo := "disabled"
			if m.rs.doBackup {
				backupInfo = m.rs.backupDir
				if backupInfo == "" {
					backupInfo = "auto-generated temp directory"
				}
			}
			action := "re-encrypted with the new plugin"
			if m.rs.action == "disable" {
				action = "reverted to plaintext"
			}
			desc = fmt.Sprintf(
				"%d file(s) will be %s.\nBackup: %s",
				count, action, backupInfo,
			)
		}
		return huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Proceed with Re-key?").
					Description(desc).
					Value(&m.rs.confirmed),
			),
		).WithShowHelp(false)

	default:
		return huh.NewForm(huh.NewGroup())
	}
}

// runDryScan builds a temporary Migrator against the selected context and runs a
// dry-run scan to discover files and check decodability.
func (m *pluginRekeyModel) runDryScan() migration.RekeyReport {
	grafanaConf, ok := m.rs.app.GetContexts()[m.rs.contextName]
	if !ok || grafanaConf == nil {
		return migration.RekeyReport{}
	}
	stor, err := storage.NewStorageFromConfig("local", nil, noop.NoOpEncoder{})
	if err != nil {
		slog.Warn("rekey dry-run: could not build local storage", "err", err)
		return migration.RekeyReport{}
	}
	scanner := migration.NewMigrator(m.rs.oldEncoder, noop.NoOpEncoder{}, grafanaConf, stor, resources.NewHelpers())
	report, _ := scanner.Rekey(migration.RekeyOptions{
		DryRun:                true,
		NoBackup:              true,
		IncludeGdgCredentials: m.rs.includeGdgCredentials,
	})
	return report
}

// ── Apply results from completed form ─────────────────────────────────────────

func (m *pluginRekeyModel) applyPhase() {
	switch m.phase {
	case phaseRekeyCurrentInfo:
		// nothing — just informational

	case phaseRekeyAction:
		// action already bound via pointer

	case phaseRekeyPluginSelect:
		// pluginName already bound; reset version/fields for fresh selection
		m.rs.pluginVersion = ""
		m.rs.resolvedEntry = nil
		m.rs.resolvedVersionEntry = nil
		m.rs.pluginConfigFields = nil
		m.rs.pluginConfigValues = make(map[string]string)
		m.rs.configFieldIdx = 0
		m.rs.configFieldValue = ""
		m.rs.newPluginEntity = nil

	case phaseRekeyPluginVersion:
		// Resolve the registry entry and version.
		for i := range m.rs.availablePlugins {
			if m.rs.availablePlugins[i].Name == m.rs.pluginName {
				m.rs.resolvedEntry = &m.rs.availablePlugins[i]
				break
			}
		}
		if m.rs.resolvedEntry != nil {
			m.rs.resolvedVersionEntry = m.rs.resolvedEntry.FindVersion(m.rs.pluginVersion)
			if m.rs.resolvedVersionEntry == nil {
				m.rs.resolvedVersionEntry = m.rs.resolvedEntry.LatestVersion()
			}
			if m.rs.resolvedVersionEntry != nil {
				m.rs.pluginConfigFields = m.rs.resolvedVersionEntry.ConfigFields
			}
		}
		m.rs.pluginConfigValues = make(map[string]string)
		m.rs.configFieldIdx = 0
		m.rs.configFieldValue = ""

	case phaseRekeyPluginConfig:
		if m.rs.configFieldIdx < len(m.rs.pluginConfigFields) {
			field := m.rs.pluginConfigFields[m.rs.configFieldIdx]
			m.rs.pluginConfigValues[field] = m.rs.configFieldValue
			m.rs.configFieldValue = ""
			m.rs.configFieldIdx++
		}
		// If all fields have been collected, build the entity.
		if m.rs.configFieldIdx >= len(m.rs.pluginConfigFields) {
			m.rs.buildNewPluginEntity()
		}

	case phaseRekeyContextSelect:
		// contextName already bound

	case phaseRekeyBackupOptions:
		// doBackup and backupDir already bound

	case phaseRekeyCredentials:
		// includeGdgCredentials already bound

	case phaseRekeyDryRun:
		// Pre-select all decodable files for the file selection phase.
		selected := make([]string, 0, len(m.rs.previews))
		for _, p := range m.rs.previews {
			if p.DecodedOK {
				selected = append(selected, p.Path)
			}
		}
		m.rs.selectedPaths = selected

	case phaseRekeyFileSelect:
		// selectedPaths already bound via pointer

	case phaseRekeyConfirm:
		// confirmed already bound
	}
}

// ── Phase transitions ─────────────────────────────────────────────────────────

func (m *pluginRekeyModel) nextPhase() rekeyPhase {
	switch m.phase {
	case phaseRekeyCurrentInfo:
		return phaseRekeyAction

	case phaseRekeyAction:
		switch m.rs.action {
		case "cancel":
			return phaseRekeyDone
		case "switch":
			return phaseRekeyPluginSelect
		default: // "disable"
			return phaseRekeyContextSelect
		}

	case phaseRekeyPluginSelect:
		if m.rs.registryError != "" {
			// Registry unavailable — bounce back to action selection.
			m.rs.action = ""
			return phaseRekeyAction
		}
		return phaseRekeyPluginVersion

	case phaseRekeyPluginVersion:
		if len(m.rs.pluginConfigFields) == 0 {
			// No config fields — entity is already built; skip to context selection.
			m.rs.buildNewPluginEntity()
			return phaseRekeyContextSelect
		}
		return phaseRekeyPluginConfig

	case phaseRekeyPluginConfig:
		// Loop until all fields have been collected.
		if m.rs.configFieldIdx < len(m.rs.pluginConfigFields) {
			return phaseRekeyPluginConfig
		}
		return phaseRekeyContextSelect

	case phaseRekeyContextSelect:
		return phaseRekeyBackupOptions

	case phaseRekeyBackupOptions:
		return phaseRekeyCredentials

	case phaseRekeyCredentials:
		return phaseRekeyDryRun

	case phaseRekeyDryRun:
		if len(m.rs.previews) == 0 {
			// Nothing to migrate — jump straight to confirm so user can exit cleanly.
			return phaseRekeyConfirm
		}
		return phaseRekeyFileSelect

	case phaseRekeyFileSelect:
		return phaseRekeyConfirm

	case phaseRekeyConfirm:
		return phaseRekeyDone

	default:
		return phaseRekeyDone
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// currentPluginDescription returns a human-readable description of the active
// cipher plugin configuration, suitable for display in a huh.Note.
func currentPluginDescription(app *config_domain.GDGAppConfiguration) string {
	if app.PluginConfig.Disabled {
		return "Cipher plugin is currently DISABLED.\n\nAll files are stored as plaintext."
	}
	pe := app.PluginConfig.CipherPlugin
	if pe == nil {
		return "No cipher plugin is currently configured.\n\nFiles are stored without encryption."
	}
	var source string
	switch {
	case pe.FilePath != "":
		source = pe.FilePath + " (local WASM file)"
	case pe.Url != "":
		source = pe.Url
	default:
		source = "(unknown — no URL or file path set)"
	}
	fields := make([]string, 0, len(pe.PluginConfig))
	for k := range pe.PluginConfig {
		fields = append(fields, k)
	}
	sort.Strings(fields)
	configDesc := "none"
	if len(fields) > 0 {
		configDesc = strings.Join(fields, ", ")
	}
	return fmt.Sprintf(
		"Active cipher plugin:\n  Source:        %s\n  Config fields: %s",
		source, configDesc,
	)
}

// ── Entry point ───────────────────────────────────────────────────────────────

// RunRekey launches the interactive plugin re-key TUI for the provided app
// configuration. It builds the current (old) encoder, collects user choices,
// runs a dry-run scan for file preview, then executes migration and optionally
// updates gdg.yml.
//
// regClient is used to fetch available cipher plugins from the registry. Pass
// nil to skip plugin selection (only "disable" will be available).
func RunRekey(app *config_domain.GDGAppConfiguration, regClient *registry.Client) error {
	// Build the old encoder from the current plugin config.
	var oldEncoder outbound.CipherEncoder = noop.NoOpEncoder{}
	if !app.PluginConfig.Disabled && app.PluginConfig.CipherPlugin != nil {
		enc, encErr := cipher.NewPluginCipherEncoder(app.PluginConfig.CipherPlugin, app.SecureConfig)
		if encErr != nil {
			return fmt.Errorf("loading current cipher plugin: %w", encErr)
		}
		oldEncoder = enc
	}

	rs := &rekeyState{
		app:         app,
		regClient:   regClient,
		oldEncoder:  oldEncoder,
		doBackup:    true,
		contextName: app.GetContext(),
	}

	m := newPluginRekeyModel(rs)
	prog := tea.NewProgram(m)
	result, err := prog.Run()
	if err != nil {
		return fmt.Errorf("rekey TUI: %w", err)
	}

	final := result.(pluginRekeyModel)
	if final.cancelled || final.rs.action == "cancel" || final.rs.action == "" {
		fmt.Println("Re-key cancelled — no changes were made.")
		return nil
	}
	if !final.rs.confirmed {
		fmt.Println("Re-key not confirmed — no changes were made.")
		return nil
	}

	// Build the new encoder.
	var newEncoder outbound.CipherEncoder = noop.NoOpEncoder{}
	if final.rs.action == "switch" && final.rs.newPluginEntity != nil {
		enc, encErr := cipher.NewPluginCipherEncoder(final.rs.newPluginEntity, app.SecureConfig)
		if encErr != nil {
			return fmt.Errorf("loading new cipher plugin: %w", encErr)
		}
		newEncoder = enc
	}

	// Build migrator.
	grafanaConf, ok := app.GetContexts()[final.rs.contextName]
	if !ok || grafanaConf == nil {
		return fmt.Errorf("context %q not found in config", final.rs.contextName)
	}
	storType, storData := app.GetCloudConfiguration(grafanaConf.Storage)
	stor, err := storage.NewStorageFromConfig(storType, storData, noop.NoOpEncoder{})
	if err != nil {
		return fmt.Errorf("build storage for context %q: %w", final.rs.contextName, err)
	}
	migrator := migration.NewMigrator(oldEncoder, newEncoder, grafanaConf, stor, resources.NewHelpers())

	opts := migration.RekeyOptions{
		NoBackup:              !final.rs.doBackup,
		BackupDir:             final.rs.backupDir,
		IncludeGdgCredentials: final.rs.includeGdgCredentials,
		AllowList:             final.rs.selectedPaths,
	}

	fmt.Printf("Re-keying context %q — processing %d file(s)...\n",
		final.rs.contextName, len(final.rs.selectedPaths))

	report, rekeyErr := migrator.Rekey(opts)
	if rekeyErr != nil {
		return fmt.Errorf("re-key failed: %w", rekeyErr)
	}

	// Print report.
	printRekeyReport(report)

	// Update gdg.yml.
	switch final.rs.action {
	case "switch":
		app.PluginConfig.Disabled = false
		app.PluginConfig.CipherPlugin = final.rs.newPluginEntity
	case "disable":
		app.PluginConfig.Disabled = true
	}
	if saveErr := app.SaveToDisk(false); saveErr != nil {
		return fmt.Errorf("saving updated config: %w", saveErr)
	}
	fmt.Println("Configuration saved to gdg.yml.")
	return nil
}

// printRekeyReport writes a human-readable summary of a RekeyReport to stdout.
func printRekeyReport(r migration.RekeyReport) {
	fmt.Println("\n── Re-key Report ─────────────────────────────────────────")
	if r.BackupDir != "" {
		fmt.Printf("  Backup location : %s\n", r.BackupDir)
	}
	total := len(r.ContactPointsFiles) + len(r.SecureDataFiles)
	if r.GdgCredentialsMigrated {
		total++
	}
	fmt.Printf("  Files migrated  : %d\n", total)
	for _, f := range r.ContactPointsFiles {
		fmt.Printf("    ✓ [contact_points] %s\n", f)
	}
	for _, f := range r.SecureDataFiles {
		fmt.Printf("    ✓ [secure_data]    %s\n", f)
	}
	if r.GdgCredentialsMigrated {
		fmt.Println("    ✓ [auth]            (gdg credentials file)")
	}
	if len(r.Errors) > 0 {
		fmt.Printf("  Errors (%d):\n", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Printf("    ✗ %s\n", e)
		}
	}
	fmt.Println("──────────────────────────────────────────────────────────")
}
