// Package config provides interactive configuration tools for GDG, including
// the TUI-driven plugin rekey workflow.
package config

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
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
	"github.com/esnet/gdg/internal/tui"
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
		return "Contexts Selection"
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

	// Context selection (one or more contexts to migrate).
	contextNames []string

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
	phase      rekeyPhase
	startPhase rekeyPhase
	rs         *rekeyState
	screen     tui.Screen
	help       help.Model
	width      int
	height     int
	done       bool
	cancelled  bool
}

func newPluginRekeyModel(rs *rekeyState) pluginRekeyModel {
	m := pluginRekeyModel{
		phase:      phaseRekeyCurrentInfo,
		startPhase: phaseRekeyCurrentInfo,
		rs:         rs,
		width:      100,
		height:     30,
		help:       help.New(),
	}
	m.screen = m.buildScreen()
	return m
}

// ── tea.Model interface ───────────────────────────────────────────────────────

func (m pluginRekeyModel) Init() tea.Cmd {
	var cmd tea.Cmd
	_, cmd = m.screen.Init()
	return cmd
}

func (m pluginRekeyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.screen = m.screen.SetWidth(m.width)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.screen, cmd = m.screen.Update(msg)

	if m.screen.Submitted {
		m.applyPhase()
		m.phase = m.nextPhase()
		if m.phase == phaseRekeyDone {
			m.done = true
			return m, tea.Quit
		}
		m.screen = m.buildScreen()
		var initCmd tea.Cmd
		m.screen, initCmd = m.screen.Init()
		return m, initCmd
	}

	if m.screen.Cancelled {
		// On the opening info screen there is nowhere to go back to — cancel.
		if m.phase == m.startPhase {
			m.cancelled = true
			return m, tea.Quit
		}
		// Multi-field config loop: step back within the loop rather than
		// jumping all the way to the previous top-level phase.
		if m.phase == phaseRekeyPluginConfig && m.rs.configFieldIdx > 0 {
			m.rs.configFieldIdx--
			m.rs.configFieldValue = m.rs.pluginConfigValues[m.rs.pluginConfigFields[m.rs.configFieldIdx]]
		} else if m.phase == phaseRekeyContextSelect && m.rs.action == "switch" && len(m.rs.pluginConfigFields) > 0 {
			// Going back from context select when config fields exist: return to
			// the last config field entry.
			m.rs.configFieldIdx = len(m.rs.pluginConfigFields) - 1
			m.rs.configFieldValue = m.rs.pluginConfigValues[m.rs.pluginConfigFields[m.rs.configFieldIdx]]
			m.phase = phaseRekeyPluginConfig
		} else {
			m.phase = m.prevPhase()
		}
		m.screen = m.buildScreen()
		var initCmd tea.Cmd
		m.screen, initCmd = m.screen.Init()
		return m, initCmd
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

	screenView := m.screen.View()
	bodyPanel := lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 4).
		Render(screenView)

	helpView := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(m.width).
		Padding(0, 2).
		Render(m.help.View(tui.DefaultKeys))

	v.Content = lipgloss.JoinVertical(lipgloss.Left, header, stepInfo, bodyPanel, helpView)
	v.AltScreen = true
	return v
}

// ── Screen creation per phase ─────────────────────────────────────────────────

func (m *pluginRekeyModel) buildScreen() tui.Screen {
	w := m.width
	if w < 40 {
		w = 80
	}

	switch m.phase {

	case phaseRekeyCurrentInfo:
		desc := currentPluginDescription(m.rs.app)
		return tui.NewScreen(w,
			tui.NewNoteField("Current Cipher Plugin", desc),
		)

	case phaseRekeyAction:
		return tui.NewScreen(w,
			tui.NewSelectField(
				"What would you like to do?",
				"",
				[]tui.Option{
					tui.NewOption("Switch to a different cipher plugin", "switch"),
					tui.NewOption("Disable encryption (revert files to plaintext)", "disable"),
					tui.NewOption("Cancel — make no changes", "cancel"),
				},
				&m.rs.action,
			),
		)

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
			return tui.NewScreen(w,
				tui.NewNoteField("Registry Unavailable",
					m.rs.registryError+"\n\nPress enter to go back and choose a different action."),
			)
		}
		opts := make([]tui.Option, len(m.rs.availablePlugins))
		for i, p := range m.rs.availablePlugins {
			label := p.Name
			if p.Description != "" {
				label += " — " + p.Description
			}
			opts[i] = tui.NewOption(label, p.Name)
		}
		return tui.NewScreen(w,
			tui.NewSelectField(
				"Select Cipher Plugin",
				"Choose a plugin from the GDG plugin registry.",
				opts,
				&m.rs.pluginName,
			),
		)

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
			return tui.NewScreen(w,
				tui.NewNoteField("No Versions",
					fmt.Sprintf("No versions found for plugin %q.", m.rs.pluginName)),
			)
		}
		opts := make([]tui.Option, len(entry.Versions))
		for i, v := range entry.Versions {
			label := v.Version
			if i == len(entry.Versions)-1 {
				label += " (latest)"
			}
			opts[i] = tui.NewOption(label, v.Version)
		}
		// Pre-select the latest version.
		if m.rs.pluginVersion == "" {
			m.rs.pluginVersion = entry.Versions[len(entry.Versions)-1].Version
		}
		sf := tui.NewSelectField(
			fmt.Sprintf("Select Version for %q", m.rs.pluginName),
			"",
			opts,
			&m.rs.pluginVersion,
		)
		return tui.NewScreen(w, sf)

	case phaseRekeyPluginConfig:
		if m.rs.configFieldIdx >= len(m.rs.pluginConfigFields) {
			return tui.NewScreen(w)
		}
		field := m.rs.pluginConfigFields[m.rs.configFieldIdx]
		total := len(m.rs.pluginConfigFields)
		// Pre-populate with any already-entered value for this field.
		if existing, ok := m.rs.pluginConfigValues[field]; ok {
			m.rs.configFieldValue = existing
		}
		tf := tui.NewTextField(
			fmt.Sprintf("Config Field: %q (%d of %d)", field, m.rs.configFieldIdx+1, total),
			"Enter a value, or use env:VAR_NAME to read from an environment variable,\n"+
				"or file:/path/to/file to read from a file.\n"+
				"Raw values are stored in gdg.yml — use env: or file: for secrets.",
			&m.rs.configFieldValue,
		)
		return tui.NewScreen(w, tf)

	case phaseRekeyContextSelect:
		allContextNames := make([]string, 0, len(m.rs.app.GetContexts()))
		for name := range m.rs.app.GetContexts() {
			allContextNames = append(allContextNames, name)
		}
		sort.Strings(allContextNames)
		opts := make([]tui.Option, len(allContextNames))
		for i, n := range allContextNames {
			opts[i] = tui.NewOption(n, n)
		}
		msf := tui.NewMultiSelectField(
			"Select Contexts to Migrate",
			"Choose one or more Grafana contexts whose encrypted files will be re-keyed.",
			opts,
			&m.rs.contextNames,
		).WithValidate(func(v []string) error {
			if len(v) == 0 {
				return fmt.Errorf("please select at least one context")
			}
			return nil
		})
		// Seed pre-selection from current contextNames value (may be set by caller).
		if len(m.rs.contextNames) > 0 {
			msf = msf.WithSelected(m.rs.contextNames)
		}
		return tui.NewScreen(w, msf)

	case phaseRekeyBackupOptions:
		return tui.NewScreen(w,
			tui.NewConfirmField(
				"Create Backups Before Modifying Files?",
				"Recommended. Original files are mirrored under the backup directory.",
				&m.rs.doBackup,
			),
			tui.NewTextField(
				"Custom Backup Directory (optional)",
				"Leave blank to use an auto-generated timestamped temporary directory.\nOnly used when backups are enabled.",
				&m.rs.backupDir,
			),
		)

	case phaseRekeyCredentials:
		return tui.NewScreen(w,
			tui.NewConfirmField(
				"Also Migrate Grafana Login Credentials?",
				"Includes the per-context auth file holding the encrypted Grafana password and/or token.",
				&m.rs.includeGdgCredentials,
			),
		)

	case phaseRekeyDryRun:
		// Run the scan synchronously and populate rs.previews / rs.scanErrors.
		// This is intentionally done inside buildScreen so results are ready for the
		// subsequent phaseRekeyFileSelect screen.
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
				statusMark := tui.GlyphCheck
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
				_, fmtErr := fmt.Fprintf(&sb, "  %s %s\n", tui.GlyphBullet, e)
				fmtErrHandler(fmtErr)
			}
		}
		return tui.NewScreen(w,
			tui.NewNoteField("Scan Results", sb.String()),
		)

	case phaseRekeyFileSelect:
		if len(m.rs.previews) == 0 {
			return tui.NewScreen(w,
				tui.NewNoteField("No Files Found",
					"No files were found to process for the selected context."),
			)
		}
		opts := make([]tui.Option, len(m.rs.previews))
		for i, p := range m.rs.previews {
			label := fmt.Sprintf("[%s] %s", p.Category, filepath.Base(p.Path))
			if !p.DecodedOK {
				label += " ⚠ (decode error)"
			}
			opts[i] = tui.NewOption(label, p.Path)
		}
		msf := tui.NewMultiSelectField(
			"Select Files to Re-Encrypt",
			"Deselect files to skip them. Files marked ⚠ failed to decode with the current plugin.",
			opts,
			&m.rs.selectedPaths,
		)
		// Pre-select only files that decoded successfully.
		preselected := make([]string, 0, len(m.rs.previews))
		for _, p := range m.rs.previews {
			if p.DecodedOK {
				preselected = append(preselected, p.Path)
			}
		}
		msf = msf.WithSelected(preselected)
		return tui.NewScreen(w, msf)

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
				"%d file(s) across %d context(s) will be %s.\nBackup: %s",
				count, len(m.rs.contextNames), action, backupInfo,
			)
		}
		return tui.NewScreen(w,
			tui.NewConfirmField("Proceed with Re-key?", desc, &m.rs.confirmed),
		)

	default:
		return tui.NewScreen(w)
	}
}

// runDryScan builds a temporary Migrator for each selected context and merges
// the dry-run scan results so all affected files appear in a single list.
func (m *pluginRekeyModel) runDryScan() migration.RekeyReport {
	merged := migration.RekeyReport{}
	for _, ctxName := range m.rs.contextNames {
		grafanaConf, ok := m.rs.app.GetContexts()[ctxName]
		if !ok || grafanaConf == nil {
			continue
		}
		stor, err := storage.NewStorageFromConfig("local", nil, noop.NoOpEncoder{})
		if err != nil {
			slog.Warn("rekey dry-run: could not build local storage", "context", ctxName, "err", err)
			continue
		}
		scanner := migration.NewMigrator(m.rs.oldEncoder, noop.NoOpEncoder{}, grafanaConf, stor, resources.NewHelpers())
		report, _ := scanner.Rekey(migration.RekeyOptions{
			DryRun:                true,
			NoBackup:              true,
			IncludeGdgCredentials: m.rs.includeGdgCredentials,
		})
		merged.Previews = append(merged.Previews, report.Previews...)
		merged.Errors = append(merged.Errors, report.Errors...)
	}
	return merged
}

// ── Apply results from completed screen ───────────────────────────────────────

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
		// contextNames already bound via pointer

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

// prevPhase returns the phase that should be displayed when the user presses
// Esc (go back). It is a pure mapping with no side-effects; callers in Update
// are responsible for any extra state mutations needed before rebuilding the
// screen (e.g. adjusting configFieldIdx for the multi-field config loop).
func (m *pluginRekeyModel) prevPhase() rekeyPhase {
	switch m.phase {
	case phaseRekeyCurrentInfo:
		return phaseRekeyCurrentInfo // no previous — caller must cancel instead
	case phaseRekeyAction:
		return phaseRekeyCurrentInfo
	case phaseRekeyPluginSelect:
		return phaseRekeyAction
	case phaseRekeyPluginVersion:
		return phaseRekeyPluginSelect
	case phaseRekeyPluginConfig:
		// Single-step-back within the loop is handled in Update; reaching here
		// means configFieldIdx == 0, so go back to the version picker.
		return phaseRekeyPluginVersion
	case phaseRekeyContextSelect:
		if m.rs.action == "switch" {
			// Caller in Update handles the config-field case; fall back to version.
			return phaseRekeyPluginVersion
		}
		return phaseRekeyAction
	case phaseRekeyBackupOptions:
		return phaseRekeyContextSelect
	case phaseRekeyCredentials:
		return phaseRekeyBackupOptions
	case phaseRekeyDryRun:
		return phaseRekeyCredentials
	case phaseRekeyFileSelect:
		return phaseRekeyDryRun
	case phaseRekeyConfirm:
		if len(m.rs.previews) > 0 {
			return phaseRekeyFileSelect
		}
		return phaseRekeyDryRun
	default:
		return phaseRekeyCurrentInfo
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// currentPluginDescription returns a human-readable description of the active
// cipher plugin configuration, suitable for display in a NoteField.
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
		app:          app,
		regClient:    regClient,
		oldEncoder:   oldEncoder,
		doBackup:     true,
		contextNames: []string{app.GetContext()},
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

	opts := migration.RekeyOptions{
		NoBackup:              !final.rs.doBackup,
		BackupDir:             final.rs.backupDir,
		IncludeGdgCredentials: final.rs.includeGdgCredentials,
		AllowList:             final.rs.selectedPaths,
	}

	fmt.Printf("Re-keying %d context(s) — processing %d file(s) total...\n",
		len(final.rs.contextNames), len(final.rs.selectedPaths))

	// Migrate each selected context. The AllowList contains absolute paths from
	// the dry-run scan; paths for other contexts are naturally filtered out by
	// the migrator (they won't appear under a different context's output dir).
	for _, ctxName := range final.rs.contextNames {
		grafanaConf, ok := app.GetContexts()[ctxName]
		if !ok || grafanaConf == nil {
			slog.Warn("rekey: context not found, skipping", "context", ctxName)
			continue
		}
		storType, storData := app.GetCloudConfiguration(grafanaConf.Storage)
		stor, err := storage.NewStorageFromConfig(storType, storData, noop.NoOpEncoder{})
		if err != nil {
			return fmt.Errorf("build storage for context %q: %w", ctxName, err)
		}
		migrator := migration.NewMigrator(oldEncoder, newEncoder, grafanaConf, stor, resources.NewHelpers())

		report, rekeyErr := migrator.Rekey(opts)
		if rekeyErr != nil {
			return fmt.Errorf("re-key failed for context %q: %w", ctxName, rekeyErr)
		}
		fmt.Printf("\nContext %q:\n", ctxName)
		printRekeyReport(report)
	}

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

// RunRekeyWithPlugin is like RunRekey but skips the plugin-selection phases
// because the caller has already collected the new plugin configuration (e.g.
// from the context-builder TUI). The TUI starts at phaseRekeyContextSelect so
// the user is only asked about which files to migrate, backup preferences, and
// whether to include the auth file — not which plugin to use.
//
// oldPlugin is the PluginEntity that was active *before* the switch (may be
// nil if no plugin was previously configured or if it was disabled). A nil
// oldPlugin causes NoOpEncoder to be used, which is correct when existing
// files are stored as plaintext.
//
// initialContextNames, if non-empty, seeds the context multi-select with the
// provided context names pre-checked. Callers should pass the names of
// contexts that existed before the current operation — this way the newly
// created context (which has no files yet) is not pre-selected. When empty,
// the active context from app.GetContext() is pre-selected.
func RunRekeyWithPlugin(
	app *config_domain.GDGAppConfiguration,
	regClient *registry.Client,
	oldPlugin *config_domain.PluginEntity,
	newPlugin *config_domain.PluginEntity,
	initialContextNames []string,
) error {
	// Build the old encoder from the explicitly-provided pre-switch plugin.
	// When oldPlugin is nil (plugin was disabled or never configured), files
	// are plaintext and NoOpEncoder is the correct decoder.
	var oldEncoder outbound.CipherEncoder = noop.NoOpEncoder{}
	if oldPlugin != nil {
		enc, encErr := cipher.NewPluginCipherEncoder(oldPlugin, app.SecureConfig)
		if encErr != nil {
			return fmt.Errorf("loading old cipher plugin: %w", encErr)
		}
		oldEncoder = enc
	}

	// Seed the context multi-select: use the provided list (typically the
	// contexts that existed before the new one was added) or fall back to the
	// currently active context.
	preselected := initialContextNames
	if len(preselected) == 0 {
		preselected = []string{app.GetContext()}
	}

	rs := &rekeyState{
		app:             app,
		regClient:       regClient,
		oldEncoder:      oldEncoder,
		doBackup:        true,
		contextNames:    preselected,
		action:          "switch",
		newPluginEntity: newPlugin,
	}

	// Start directly at context selection — plugin has already been chosen.
	m := pluginRekeyModel{
		phase:      phaseRekeyContextSelect,
		startPhase: phaseRekeyContextSelect,
		rs:         rs,
		width:      100,
		height:     30,
		help:       help.New(),
	}
	m.screen = m.buildScreen()

	prog := tea.NewProgram(m)
	result, err := prog.Run()
	if err != nil {
		return fmt.Errorf("rekey TUI: %w", err)
	}

	final := result.(pluginRekeyModel)
	if final.cancelled || final.rs.action == "cancel" {
		fmt.Println("Re-key cancelled — no changes were made.")
		return nil
	}
	if !final.rs.confirmed {
		fmt.Println("Re-key not confirmed — no changes were made.")
		return nil
	}

	// Build the new encoder from the pre-chosen plugin.
	var newEncoder outbound.CipherEncoder = noop.NoOpEncoder{}
	if final.rs.newPluginEntity != nil {
		enc, encErr := cipher.NewPluginCipherEncoder(final.rs.newPluginEntity, app.SecureConfig)
		if encErr != nil {
			return fmt.Errorf("loading new cipher plugin: %w", encErr)
		}
		newEncoder = enc
	}

	opts := migration.RekeyOptions{
		NoBackup:              !final.rs.doBackup,
		BackupDir:             final.rs.backupDir,
		IncludeGdgCredentials: final.rs.includeGdgCredentials,
		AllowList:             final.rs.selectedPaths,
	}

	fmt.Printf("Re-keying %d context(s) — processing %d file(s) total...\n",
		len(final.rs.contextNames), len(final.rs.selectedPaths))

	for _, ctxName := range final.rs.contextNames {
		grafanaConf, ok := app.GetContexts()[ctxName]
		if !ok || grafanaConf == nil {
			slog.Warn("rekey: context not found, skipping", "context", ctxName)
			continue
		}
		storType, storData := app.GetCloudConfiguration(grafanaConf.Storage)
		stor, err := storage.NewStorageFromConfig(storType, storData, noop.NoOpEncoder{})
		if err != nil {
			return fmt.Errorf("build storage for context %q: %w", ctxName, err)
		}
		migrator := migration.NewMigrator(oldEncoder, newEncoder, grafanaConf, stor, resources.NewHelpers())

		report, rekeyErr := migrator.Rekey(opts)
		if rekeyErr != nil {
			return fmt.Errorf("re-key failed for context %q: %w", ctxName, rekeyErr)
		}
		fmt.Printf("\nContext %q:\n", ctxName)
		printRekeyReport(report)
	}
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
		fmt.Printf("    %s [contact_points] %s\n", tui.GlyphCheck, f)
	}
	for _, f := range r.SecureDataFiles {
		fmt.Printf("    %s [secure_data]    %s\n", tui.GlyphCheck, f)
	}
	if r.GdgCredentialsMigrated {
		fmt.Printf("    %s [auth]            (gdg credentials file)\n", tui.GlyphCheck)
	}
	if len(r.Errors) > 0 {
		fmt.Printf("  Errors (%d):\n", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Printf("    %s %s\n", tui.GlyphCross, e)
		}
	}
	fmt.Println("──────────────────────────────────────────────────────────")
}
