package config

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/esnet/gdg/internal/adapter/plugins/registry"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/internal/tui"
	"gopkg.in/yaml.v3"
)

// ── Builder TUI phases ────────────────────────────────────────────────────────

type builderPhase int

const (
	phaseAuthType builderPhase = iota
	phaseBasicCreds
	phaseTokenCreds
	phaseServerSettings
	phaseUserSettings // "Generate random passwords for imported users?"
	phaseUserLengths  // (if yes) Min / Max password length
	phaseFolderScope
	phaseFolderMenu       // select: Add / Test / Done
	phaseFolderAdd        // input folder name
	phaseFolderTest       // input test value
	phaseFolderTestResult // show results + test another?
	phaseConnectionToggle
	phaseFilterToggle
	phaseFilterInput
	phaseDefaultCreds
	phaseCredRuleToggle
	phaseCredRuleFile    // input secure data filename
	phaseCredRuleMatcher // input field + regex + add another matcher?
	phaseCredRuleCreds   // input user + password + add another rule?
	phaseStorageToggle
	phaseStorageProvider
	phaseStorageProviderInfo // shows docs for managed providers (S3/GCS/Azure)
	phaseStorageCustomConfig
	phaseStorageCustomCreds
	phaseStorageCustomOptions
	phaseStorageAssign
	phasePluginToggle  // "Configure a cipher plugin?"
	phasePluginSelect  // select plugin name from registry
	phasePluginVersion // select plugin version
	phasePluginConfig  // input per config_field (loops)
	phaseDone
)

func (p builderPhase) sectionName() string {
	switch p {
	case phaseAuthType, phaseBasicCreds, phaseTokenCreds:
		return "Authentication"
	case phaseServerSettings:
		return "Server Settings"
	case phaseUserSettings, phaseUserLengths:
		return "User Settings"
	case phaseFolderScope, phaseFolderMenu, phaseFolderAdd, phaseFolderTest, phaseFolderTestResult:
		return "Watched Folders"
	case phaseConnectionToggle, phaseFilterToggle, phaseFilterInput,
		phaseDefaultCreds, phaseCredRuleToggle, phaseCredRuleFile,
		phaseCredRuleMatcher, phaseCredRuleCreds:
		return "Connection Settings"
	case phaseStorageToggle, phaseStorageProvider, phaseStorageProviderInfo,
		phaseStorageCustomConfig, phaseStorageCustomCreds, phaseStorageCustomOptions,
		phaseStorageAssign:
		return "Cloud Storage"
	case phasePluginToggle, phasePluginSelect, phasePluginVersion, phasePluginConfig:
		return "Cipher Plugin"
	default:
		return ""
	}
}

// ── Pending credential file (deferred write) ─────────────────────────────────

type pendingCredFile struct {
	secureData string
	user       string
	password   string
}

// ── Builder state ─────────────────────────────────────────────────────────────

type builderState struct {
	contextName string
	app         *config_domain.GDGAppConfiguration
	encoder     outbound.CipherEncoder

	// Config being assembled
	config *config_domain.GrafanaConfig
	secure *config_domain.SecureModel

	// Auth
	authType string
	userName string
	password string
	token    string

	// Server
	url        string
	outputPath string

	// UserSettings
	userRandomPassword bool
	userMinLength      string // stored as string for tui.TextField; parsed to int in applyPhase
	userMaxLength      string // stored as string for tui.TextField; parsed to int in applyPhase

	// Folders
	folderScope      string // "all" or "allowlist"
	folderAction     string // "add", "test", "done"
	folderName       string
	folders          []string
	folderTestValue  string
	folderTestResult string // rendered result text from last test
	folderTestMore   bool

	// Connections toggle
	configureConnections bool

	// Filters
	addFilters      bool
	filterField     string
	filterRegex     string
	filterInclusive bool
	addMoreFilters  bool
	filters         []config_domain.MatchingRule

	// Default connection credentials
	connectionUser     string
	connectionPassword string

	// Credential rules
	addCredRules       bool
	credSecureData     string
	credField          string
	credRegex          string
	credAddMoreMatcher bool
	credUser           string
	credPassword       string
	addMoreCredRules   bool
	credCurrentRules   []config_domain.MatchingRule // matchers for current rule being built
	credRules          []*config_domain.RegexMatchesList
	pendingCreds       []pendingCredFile

	// Storage
	configureStorage  bool
	storageProvider   string
	storageLabel      string
	storageEndpoint   string
	storageBucket     string
	storageRegion     string
	storageAccessID   string
	storageSecretKey  string
	storagePrefix     string
	storageInitBucket bool
	storageSSL        bool
	storageAssign     bool

	// Plugin
	configurePlugin          bool
	pluginName               string
	pluginVersion            string
	pluginConfigValues       map[string]string
	pluginConfigFields       []string
	pluginConfigIdx          int
	pluginConfigCurrentValue string // scratch field for current config-field input
	pluginResolvedURL        string
	pluginLoadErr            bool
	pluginResult             *config_domain.PluginEntity
	registryClient           *registry.Client
}

func (s *builderState) renderPreview() string {
	type pluginsPreview struct {
		CipherPlugin *config_domain.PluginEntity `yaml:"cipher_plugin,omitempty"`
	}
	type previewConfig struct {
		ContextName string                                  `yaml:"context_name"`
		Contexts    map[string]*config_domain.GrafanaConfig `yaml:"contexts"`
		Plugins     *pluginsPreview                         `yaml:"plugins,omitempty"`
	}
	p := previewConfig{
		ContextName: s.contextName,
		Contexts:    map[string]*config_domain.GrafanaConfig{s.contextName: s.config},
	}
	if s.pluginResult != nil {
		p.Plugins = &pluginsPreview{CipherPlugin: s.pluginResult}
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return "# error rendering preview"
	}
	return string(data)
}

// ── Layout constants ──────────────────────────────────────────────────────────

const (
	tuiHeaderHeight = 3
	tuiFooterHeight = 2
	tuiMinBodyH     = 10
)

// ── Builder model ─────────────────────────────────────────────────────────────

type configBuilderModel struct {
	phase      builderPhase
	startPhase builderPhase // Esc on this phase cancels the wizard
	bs         *builderState
	screen     tui.Screen
	help       help.Model
	width      int
	height     int
	done       bool
	cancelled  bool
}

func newConfigBuilderModel(
	app *config_domain.GDGAppConfiguration,
	name string,
	encoder outbound.CipherEncoder,
	regClient *registry.Client,
) configBuilderModel {
	bs := &builderState{
		contextName:    name,
		app:            app,
		encoder:        encoder,
		config:         config_domain.NewGrafanaConfig(config_domain.WithContextName(name)),
		secure:         &config_domain.SecureModel{},
		registryClient: regClient,
	}
	bs.config.OrganizationName = "Main Org."
	bs.config.ConnectionSettings = &config_domain.ConnectionSettings{
		MatchingRules: make([]*config_domain.RegexMatchesList, 0),
	}

	startPhase := phasePluginToggle
	if !app.PluginConfig.Disabled && app.PluginConfig.CipherPlugin != nil {
		startPhase = phaseAuthType
	}

	h := help.New()
	h.ShowAll = false

	m := configBuilderModel{
		phase:      startPhase,
		startPhase: startPhase,
		bs:         bs,
		help:       h,
		width:      100,
		height:     30,
	}
	m.screen = m.buildScreen()
	var initCmd tea.Cmd
	m.screen, initCmd = m.screen.Init()
	_ = initCmd // stored and returned from Init()
	return m
}

// ── tea.Model interface ───────────────────────────────────────────────────────

func (m configBuilderModel) Init() tea.Cmd {
	_, cmd := m.screen.Init()
	return cmd
}

func (m configBuilderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.SetWidth(m.leftWidth())
		m.screen = m.screen.SetWidth(m.leftWidth())
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
		if m.phase == phaseDone {
			m.done = true
			return m, tea.Quit
		}
		m.screen = m.buildScreen()
		var initCmd tea.Cmd
		m.screen, initCmd = m.screen.Init()
		return m, initCmd
	}

	if m.screen.Cancelled {
		// Esc on the opening phase cancels the wizard entirely.
		if m.phase == m.startPhase {
			m.cancelled = true
			return m, tea.Quit
		}
		// Plugin config loop: step back one field at a time.
		if m.phase == phasePluginConfig && m.bs.pluginConfigIdx > 0 {
			m.bs.pluginConfigIdx--
			m.bs.pluginConfigCurrentValue = m.bs.pluginConfigValues[m.bs.pluginConfigFields[m.bs.pluginConfigIdx]]
		} else {
			prev := m.prevPhase()
			// Navigating back into the plugin-config loop from phaseAuthType:
			// restore position to the last config field.
			if prev == phasePluginConfig && len(m.bs.pluginConfigFields) > 0 {
				m.bs.pluginConfigIdx = len(m.bs.pluginConfigFields) - 1
				m.bs.pluginConfigCurrentValue = m.bs.pluginConfigValues[m.bs.pluginConfigFields[m.bs.pluginConfigIdx]]
			}
			m.phase = prev
		}
		m.screen = m.buildScreen()
		var initCmd tea.Cmd
		m.screen, initCmd = m.screen.Init()
		return m, initCmd
	}

	return m, cmd
}

func (m configBuilderModel) View() tea.View {
	if m.done || m.cancelled {
		return tea.NewView("")
	}
	v := tea.NewView("")

	// ── Header ──
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("63")).
		Align(lipgloss.Center).
		Width(m.width).
		Padding(0, 1)
	header := headerStyle.Render("GDG Config Builder")

	stepStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Foreground(lipgloss.Color("245"))
	stepInfo := stepStyle.Render(m.phase.sectionName())

	// ── Body: left (form) + right (preview) ──
	bodyH := m.bodyHeight()
	leftW := m.leftWidth()
	rightW := m.rightWidth()

	formContent := m.screen.View()
	leftPanel := lipgloss.NewStyle().
		Width(leftW).
		Height(bodyH).
		Padding(1, 2).
		Render(formContent)

	previewTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("215")).
		Render("Preview (gdg.yml)")
	preview := m.bs.renderPreview()
	rightPanel := lipgloss.NewStyle().
		Width(rightW).
		Height(bodyH).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		Foreground(lipgloss.Color("252")).
		Render(previewTitle + "\n\n" + preview)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// ── Footer ──
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(m.width).
		Padding(0, 2)
	footer := footerStyle.Render(m.help.View(tui.DefaultKeys))

	v.Content = lipgloss.JoinVertical(lipgloss.Left, header, stepInfo, body, footer)
	v.AltScreen = true
	return v
}

// ── Layout helpers ────────────────────────────────────────────────────────────

func (m configBuilderModel) leftWidth() int {
	w := m.width / 2
	if w < 30 {
		w = 30
	}
	return w
}

func (m configBuilderModel) rightWidth() int {
	return m.width - m.leftWidth()
}

func (m configBuilderModel) bodyHeight() int {
	h := m.height - tuiHeaderHeight - tuiFooterHeight
	if h < tuiMinBodyH {
		h = tuiMinBodyH
	}
	return h
}

// ── Screen construction per phase ─────────────────────────────────────────────

func (m *configBuilderModel) buildScreen() tui.Screen {
	w := m.leftWidth()

	switch m.phase {

	// ── Auth ──────────────────────────────────────────────────────────────

	case phaseAuthType:
		return tui.NewScreen(w,
			tui.NewSelectField(
				"Authentication Type",
				"How should GDG authenticate with Grafana?",
				[]tui.Option{
					{Label: "Basic (username + password)", Value: basicAuthForm.String()},
					{Label: "Token / Service account", Value: tokenAuthForm.String()},
					{Label: "Both (basic + token)", Value: bothAuthForm.String()},
				},
				&m.bs.authType,
			),
		)

	case phaseBasicCreds:
		return tui.NewScreen(w,
			tui.NewTextField("Grafana Username", "Username for basic authentication", &m.bs.userName),
			tui.NewTextField("Grafana Password", "", &m.bs.password).WithMask(),
		)

	case phaseTokenCreds:
		return tui.NewScreen(w,
			tui.NewTextField("Grafana Token", "API token or service account token", &m.bs.token).WithMask(),
		)

	// ── Server / User ─────────────────────────────────────────────────────

	case phaseServerSettings:
		return tui.NewScreen(w,
			tui.NewTextField("Grafana URL", "Include scheme (e.g. http://grafana.example.com)", &m.bs.url).
				WithValidate(validateGrafanaURL),
			tui.NewTextField("Output Path", "Local folder for storing backups", &m.bs.outputPath),
		)

	case phaseUserSettings:
		return tui.NewScreen(w,
			tui.NewConfirmField(
				"Generate random passwords for imported users?",
				"Yes: a random password is generated per user.\n"+
					"No: a deterministic password is derived from the username.",
				&m.bs.userRandomPassword,
			),
		)

	case phaseUserLengths:
		if m.bs.userMinLength == "" {
			m.bs.userMinLength = "8"
		}
		if m.bs.userMaxLength == "" {
			m.bs.userMaxLength = "20"
		}
		return tui.NewScreen(w,
			tui.NewTextField("Minimum Password Length", "Shortest password to generate (positive integer, default 8)", &m.bs.userMinLength).
				WithValidate(validatePositiveInt),
			tui.NewTextField("Maximum Password Length", "Longest password to generate (≥ min, default 20)", &m.bs.userMaxLength).
				WithValidate(validatePositiveInt),
		)

	// ── Folders ───────────────────────────────────────────────────────────

	case phaseFolderScope:
		return tui.NewScreen(w,
			tui.NewSelectField(
				"Folder Scope",
				"Monitor all Grafana folders or restrict to a specific allowlist?",
				[]tui.Option{
					{Label: "Monitor all folders (ignore filters)", Value: "all"},
					{Label: "Build a folder allowlist", Value: "allowlist"},
				},
				&m.bs.folderScope,
			),
		)

	case phaseFolderMenu:
		currentList := "none"
		if len(m.bs.folders) > 0 {
			currentList = strings.Join(m.bs.folders, ", ")
		}
		opts := []tui.Option{{Label: "Add a folder", Value: "add"}}
		if len(m.bs.folders) > 0 {
			opts = append(opts,
				tui.Option{Label: "Test a folder name against current list", Value: "test"},
				tui.Option{Label: "Done — use this list", Value: "done"},
			)
		}
		return tui.NewScreen(w,
			tui.NewSelectField(
				"Watched Folders",
				fmt.Sprintf("Current: [%s]", currentList),
				opts,
				&m.bs.folderAction,
			),
		)

	case phaseFolderAdd:
		m.bs.folderName = ""
		return tui.NewScreen(w,
			tui.NewTextField(
				"Folder Name or Regex",
				"Literal names are URL-encoded automatically.\nRegex patterns (containing *?[]|^$\\) are stored as-is.",
				&m.bs.folderName,
			),
		)

	case phaseFolderTest:
		currentList := strings.Join(m.bs.folders, ", ")
		m.bs.folderTestValue = ""
		return tui.NewScreen(w,
			tui.NewTextField(
				"Test Folder Name",
				fmt.Sprintf("Test against current list: [%s]", currentList),
				&m.bs.folderTestValue,
			),
		)

	case phaseFolderTestResult:
		m.bs.folderTestMore = true
		return tui.NewScreen(w,
			tui.NewNoteField("Test Results", m.bs.folderTestResult),
			tui.NewConfirmField("Test another folder name?", "", &m.bs.folderTestMore),
		)

	// ── Connections ───────────────────────────────────────────────────────

	case phaseConnectionToggle:
		return tui.NewScreen(w,
			tui.NewConfirmField(
				"Configure connection settings?",
				"Set up data source filters, default credentials, and credential rules.\n"+
					"Skip this to configure later by editing gdg.yml.",
				&m.bs.configureConnections,
			),
		)

	case phaseFilterToggle:
		return tui.NewScreen(w,
			tui.NewConfirmField(
				"Configure connection filters?",
				"Filters control which data sources are included or excluded.",
				&m.bs.addFilters,
			),
		)

	case phaseFilterInput:
		m.bs.filterField = ""
		m.bs.filterRegex = ""
		m.bs.filterInclusive = false
		m.bs.addMoreFilters = false
		desc := fmt.Sprintf("Filters so far: %s", summariseFilters(m.bs.filters))
		return tui.NewScreen(w,
			tui.NewTextField("Filter Field", desc+"\nJSON field to match (e.g. 'name', 'type', 'url')", &m.bs.filterField),
			tui.NewTextField("Filter Regex", "Regular expression to match against the field value", &m.bs.filterRegex),
			tui.NewConfirmField("Inclusive filter? (allowlist)", "Yes = keep only matches.  No = drop matches (denylist).", &m.bs.filterInclusive),
			tui.NewConfirmField("Add another filter?", "", &m.bs.addMoreFilters),
		)

	case phaseDefaultCreds:
		return tui.NewScreen(w,
			tui.NewTextField("Connection Default User", "Default user for data source connections", &m.bs.connectionUser),
			tui.NewTextField("Connection Default Password", "", &m.bs.connectionPassword).WithMask(),
		)

	case phaseCredRuleToggle:
		return tui.NewScreen(w,
			tui.NewConfirmField(
				"Configure credential rules?",
				"Map data sources to specific credential files.\n"+
					"A default catch-all rule (name=.*) is always appended.",
				&m.bs.addCredRules,
			),
		)

	case phaseCredRuleFile:
		m.bs.credSecureData = ""
		m.bs.credCurrentRules = nil
		return tui.NewScreen(w,
			tui.NewTextField(
				"Secure Data File",
				"Credentials filename (e.g. 'elastic.yaml').  'default.yaml' is reserved.",
				&m.bs.credSecureData,
			).WithValidate(validateSecureDataFile),
		)

	case phaseCredRuleMatcher:
		m.bs.credField = ""
		m.bs.credRegex = ""
		m.bs.credAddMoreMatcher = false
		matcherDesc := "none"
		if len(m.bs.credCurrentRules) > 0 {
			matcherDesc = summariseFilters(m.bs.credCurrentRules)
		}
		return tui.NewScreen(w,
			tui.NewTextField("Matching Field",
				fmt.Sprintf("Matchers so far: %s\nField to match (e.g. 'name', 'url', 'type')", matcherDesc),
				&m.bs.credField,
			),
			tui.NewTextField("Matching Regex", "Regular expression to match against the field value", &m.bs.credRegex),
			tui.NewConfirmField("Add another matcher for this credential file?", "", &m.bs.credAddMoreMatcher),
		)

	case phaseCredRuleCreds:
		m.bs.credUser = ""
		m.bs.credPassword = ""
		m.bs.addMoreCredRules = false
		return tui.NewScreen(w,
			tui.NewTextField(
				"User for this credential file",
				fmt.Sprintf("File: %s | Matchers: %s", m.bs.credSecureData, summariseFilters(m.bs.credCurrentRules)),
				&m.bs.credUser,
			),
			tui.NewTextField("Password for this credential file", "", &m.bs.credPassword).WithMask(),
			tui.NewConfirmField("Add another credential rule?", "", &m.bs.addMoreCredRules),
		)

	// ── Storage ───────────────────────────────────────────────────────────

	case phaseStorageToggle:
		return tui.NewScreen(w,
			tui.NewConfirmField(
				"Configure cloud storage?",
				"Set up a cloud storage engine for this context.\n"+
					"Skip and add one later via the config file.",
				&m.bs.configureStorage,
			),
		)

	case phaseStorageProvider:
		return tui.NewScreen(w,
			tui.NewSelectField(
				"Cloud Storage Provider",
				"For AWS S3, GCS, and Azure, GDG delegates auth to the provider SDK.\n"+
					"Only custom S3-compatible endpoints are configured here.",
				[]tui.Option{
					{Label: "Custom S3-compatible (Minio, Ceph, ...)", Value: string(providerCustom)},
					{Label: "AWS S3", Value: string(providerAWS)},
					{Label: "Google Cloud Storage (GCS)", Value: string(providerGCS)},
					{Label: "Azure Blob Storage", Value: string(providerAzure)},
				},
				&m.bs.storageProvider,
			),
		)

	case phaseStorageProviderInfo:
		cp := cloudProvider(m.bs.storageProvider)
		docURL := providerDocURLs[cp]
		info := fmt.Sprintf(
			"For %s, authentication is handled by the provider SDK — not configured here.\n\n"+
				"Documentation:\n  %s\n\n"+
				"Once credentials are in place, add a storage_engine entry to gdg.yml:\n"+
				"  cloud_type: %s\n  bucket_name: <your-bucket>",
			cp, docURL, cp,
		)
		return tui.NewScreen(w,
			tui.NewNoteField("Managed Provider Setup", info),
			tui.NewConfirmField("Understood — finish setup?", "", &m.bs.storageAssign),
		)

	case phaseStorageCustomConfig:
		return tui.NewScreen(w,
			tui.NewTextField("Storage Engine Label", "Unique key for this config in gdg.yml (e.g. my-minio)", &m.bs.storageLabel),
			tui.NewTextField("Endpoint URL", "Full URL of the S3-compatible endpoint (e.g. http://localhost:9000)", &m.bs.storageEndpoint),
			tui.NewTextField("Bucket Name", "", &m.bs.storageBucket),
			tui.NewTextField("Region", "AWS region or equivalent (default: us-east-1)", &m.bs.storageRegion),
		)

	case phaseStorageCustomCreds:
		return tui.NewScreen(w,
			tui.NewTextField("Access Key ID", "", &m.bs.storageAccessID),
			tui.NewTextField("Secret Access Key", "", &m.bs.storageSecretKey).WithMask(),
		)

	case phaseStorageCustomOptions:
		return tui.NewScreen(w,
			tui.NewTextField("Path Prefix (optional)", "Prefix applied to all object paths within the bucket", &m.bs.storagePrefix),
			tui.NewConfirmField("Auto-create bucket if it does not exist?", "", &m.bs.storageInitBucket),
			tui.NewConfirmField("Enable SSL?", "", &m.bs.storageSSL),
		)

	case phaseStorageAssign:
		ctx := m.bs.app.GetContext()
		return tui.NewScreen(w,
			tui.NewConfirmField(
				fmt.Sprintf("Assign this storage engine to context %q?", ctx),
				"",
				&m.bs.storageAssign,
			),
		)

	// ── Plugin ────────────────────────────────────────────────────────────

	case phasePluginToggle:
		return tui.NewScreen(w,
			tui.NewConfirmField(
				"Configure a cipher plugin?",
				"Encrypt credentials at rest using a WASM cipher plugin.\n"+
					"Skip this and configure it later in gdg.yml.",
				&m.bs.configurePlugin,
			),
		)

	case phasePluginSelect:
		plugins, err := m.bs.registryClient.CipherPlugins()
		if err != nil || len(plugins) == 0 {
			m.bs.pluginLoadErr = true
			m.bs.configurePlugin = false
			msg := "Could not load cipher plugins from the registry.\n" +
				"Plugin configuration will be skipped.\n" +
				"Check your network, or run 'gdg tools plugins rekey' later."
			return tui.NewScreen(w,
				tui.NewNoteField("Plugin Registry Unavailable", msg),
			)
		}
		m.bs.pluginLoadErr = false
		opts := make([]tui.Option, 0, len(plugins))
		for _, p := range plugins {
			label := p.Name
			if p.Description != "" {
				label = fmt.Sprintf("%s — %s", p.Name, p.Description)
			}
			opts = append(opts, tui.Option{Label: label, Value: p.Name})
		}
		return tui.NewScreen(w,
			tui.NewSelectField(
				"Cipher Plugin",
				"Select the plugin to use for encrypting credentials in this context.",
				opts,
				&m.bs.pluginName,
			),
		)

	case phasePluginVersion:
		entry, _ := m.bs.registryClient.Find(m.bs.pluginName)
		var opts []tui.Option
		if entry != nil {
			for _, v := range entry.Versions {
				label := v.Version
				if len(v.ConfigFields) > 0 {
					label += fmt.Sprintf(" (fields: %s)", strings.Join(v.ConfigFields, ", "))
				}
				opts = append(opts, tui.Option{Label: label, Value: v.Version})
			}
		}
		if len(opts) == 0 {
			opts = append(opts, tui.Option{Label: "latest", Value: "latest"})
		}
		return tui.NewScreen(w,
			tui.NewSelectField(
				"Plugin Version",
				fmt.Sprintf("Select a version of %q", m.bs.pluginName),
				opts,
				&m.bs.pluginVersion,
			),
		)

	case phasePluginConfig:
		m.bs.pluginConfigCurrentValue = ""
		if existing, ok := m.bs.pluginConfigValues[m.bs.pluginConfigFields[m.bs.pluginConfigIdx]]; ok {
			m.bs.pluginConfigCurrentValue = existing
		}
		field := m.bs.pluginConfigFields[m.bs.pluginConfigIdx]
		progress := fmt.Sprintf("Field %d of %d", m.bs.pluginConfigIdx+1, len(m.bs.pluginConfigFields))
		return tui.NewScreen(w,
			tui.NewTextField(
				fmt.Sprintf("Config field: %q  (%s)", field, progress),
				"Enter a raw value, env:VAR_NAME to read from an env variable,\n"+
					"or file:/path/to/file to read from a file.\n"+
					"Raw values are stored in gdg.yml — use env: or file: for secrets.",
				&m.bs.pluginConfigCurrentValue,
			),
		)

	default:
		return tui.NewScreen(w)
	}
}

// ── Apply results from completed screen ───────────────────────────────────────

func (m *configBuilderModel) applyPhase() {
	switch m.phase {
	case phaseAuthType:
		// authType already bound via pointer

	case phaseBasicCreds:
		m.bs.config.UserName = m.bs.userName
		m.bs.secure.Password = m.bs.password

	case phaseTokenCreds:
		m.bs.secure.Token = m.bs.token

	case phaseServerSettings:
		m.bs.config.URL = m.bs.url
		m.bs.config.OutputPath = m.bs.outputPath

	case phaseUserSettings:
		if !m.bs.userRandomPassword {
			m.bs.config.UserSettings = &config_domain.UserSettings{RandomPassword: false}
		}

	case phaseUserLengths:
		minL, _ := strconv.Atoi(strings.TrimSpace(m.bs.userMinLength))
		maxL, _ := strconv.Atoi(strings.TrimSpace(m.bs.userMaxLength))
		if minL <= 0 {
			minL = 8
		}
		if maxL <= 0 {
			maxL = 20
		}
		m.bs.config.UserSettings = &config_domain.UserSettings{
			RandomPassword: true,
			MinLength:      minL,
			MaxLength:      maxL,
		}

	case phaseFolderScope:
		if m.bs.folderScope == "all" {
			if m.bs.config.DashboardSettings == nil {
				m.bs.config.DashboardSettings = &config_domain.DashboardSettings{}
			}
			m.bs.config.DashboardSettings.IgnoreFilters = true
		}

	case phaseFolderMenu:
		if m.bs.folderAction == "done" {
			if len(m.bs.folders) == 0 {
				m.bs.folders = []string{"General"}
			}
			m.bs.config.MonitoredFolders = m.bs.folders
		}

	case phaseFolderAdd:
		name := strings.TrimSpace(m.bs.folderName)
		if name != "" {
			if looksLikeRegex(name) {
				m.bs.folders = append(m.bs.folders, name)
			} else {
				m.bs.folders = append(m.bs.folders, encodeFolderName(name))
			}
			m.bs.config.MonitoredFolders = m.bs.folders
		}

	case phaseFolderTest:
		testVal := strings.TrimSpace(m.bs.folderTestValue)
		if testVal != "" {
			matched, patterns := testFolderRegexMatch(m.bs.folders, testVal)
			if matched {
				m.bs.folderTestResult = fmt.Sprintf("MATCH: %q matched patterns: %s", testVal, strings.Join(patterns, ", "))
			} else {
				m.bs.folderTestResult = fmt.Sprintf("NO MATCH: %q did not match any of: [%s]", testVal, strings.Join(m.bs.folders, ", "))
			}
		} else {
			m.bs.folderTestResult = "No value entered."
		}

	case phaseFolderTestResult:
		// folderTestMore already bound

	case phaseConnectionToggle:
		// configureConnections already bound

	case phaseFilterToggle:
		// addFilters already bound

	case phaseFilterInput:
		rule, err := validateFilter(m.bs.filterField, m.bs.filterRegex, m.bs.filterInclusive)
		if err == nil {
			m.bs.filters = append(m.bs.filters, *rule)
		}
		m.bs.config.ConnectionSettings.FilterRules = m.bs.filters

	case phaseDefaultCreds:
		// connectionUser / connectionPassword stored for post-TUI file write

	case phaseCredRuleToggle:
		// addCredRules already bound

	case phaseCredRuleFile:
		// credSecureData already bound; credCurrentRules reset in buildScreen

	case phaseCredRuleMatcher:
		rule, err := validateCredentialRule(m.bs.credField, m.bs.credRegex)
		if err == nil {
			m.bs.credCurrentRules = append(m.bs.credCurrentRules, *rule)
		}

	case phaseCredRuleCreds:
		if len(m.bs.credCurrentRules) > 0 {
			m.bs.credRules = append(m.bs.credRules, &config_domain.RegexMatchesList{
				Rules:      m.bs.credCurrentRules,
				SecureData: strings.TrimSpace(m.bs.credSecureData),
			})
			m.bs.pendingCreds = append(m.bs.pendingCreds, pendingCredFile{
				secureData: strings.TrimSpace(m.bs.credSecureData),
				user:       m.bs.credUser,
				password:   m.bs.credPassword,
			})
		}
		m.bs.credCurrentRules = nil
		m.bs.config.ConnectionSettings.MatchingRules = appendDefaultCredentialRule(m.bs.credRules)

	case phaseStorageToggle:
		// configureStorage already bound

	case phaseStorageProvider:
		// storageProvider already bound

	case phaseStorageCustomConfig, phaseStorageCustomCreds, phaseStorageCustomOptions:
		// all values bound via pointers

	case phaseStorageAssign:
		if m.bs.storageAssign {
			m.bs.config.Storage = m.bs.storageLabel
		}

	case phasePluginToggle:
		// configurePlugin already bound

	case phasePluginSelect:
		// pluginName already bound (or pluginLoadErr set)

	case phasePluginVersion:
		if m.bs.registryClient != nil && m.bs.pluginName != "" && m.bs.pluginVersion != "" {
			_, versionEntry, resolvedURL, err := m.bs.registryClient.ResolvePlugin(m.bs.pluginName, m.bs.pluginVersion)
			if err == nil && versionEntry != nil {
				m.bs.pluginConfigFields = versionEntry.ConfigFields
				m.bs.pluginResolvedURL = resolvedURL
			}
		}
		m.bs.pluginConfigIdx = 0
		m.bs.pluginConfigValues = make(map[string]string)
		if len(m.bs.pluginConfigFields) == 0 {
			m.bs.pluginResult = &config_domain.PluginEntity{Url: m.bs.pluginResolvedURL}
		}

	case phasePluginConfig:
		if m.bs.pluginConfigIdx < len(m.bs.pluginConfigFields) {
			field := m.bs.pluginConfigFields[m.bs.pluginConfigIdx]
			m.bs.pluginConfigValues[field] = m.bs.pluginConfigCurrentValue
			m.bs.pluginConfigIdx++
		}
		if m.bs.pluginConfigIdx >= len(m.bs.pluginConfigFields) {
			m.bs.pluginResult = &config_domain.PluginEntity{
				Url:          m.bs.pluginResolvedURL,
				PluginConfig: m.bs.pluginConfigValues,
			}
		}
	}
}

// ── Forward phase transitions ─────────────────────────────────────────────────

func (m *configBuilderModel) nextPhase() builderPhase {
	switch m.phase {
	case phasePluginToggle:
		if m.bs.configurePlugin {
			return phasePluginSelect
		}
		return phaseAuthType
	case phasePluginSelect:
		if m.bs.pluginLoadErr {
			return phaseAuthType
		}
		return phasePluginVersion
	case phasePluginVersion:
		if len(m.bs.pluginConfigFields) == 0 {
			return phaseAuthType
		}
		return phasePluginConfig
	case phasePluginConfig:
		if m.bs.pluginConfigIdx < len(m.bs.pluginConfigFields) {
			return phasePluginConfig
		}
		return phaseAuthType

	case phaseAuthType:
		switch m.bs.authType {
		case basicAuthForm.String():
			return phaseBasicCreds
		case tokenAuthForm.String():
			return phaseTokenCreds
		case bothAuthForm.String():
			return phaseBasicCreds
		}
		return phaseServerSettings
	case phaseBasicCreds:
		if m.bs.authType == bothAuthForm.String() {
			return phaseTokenCreds
		}
		return phaseServerSettings
	case phaseTokenCreds:
		return phaseServerSettings
	case phaseServerSettings:
		return phaseUserSettings
	case phaseUserSettings:
		if m.bs.userRandomPassword {
			return phaseUserLengths
		}
		return phaseFolderScope
	case phaseUserLengths:
		return phaseFolderScope

	case phaseFolderScope:
		if m.bs.folderScope == "all" {
			return phaseConnectionToggle
		}
		return phaseFolderMenu
	case phaseFolderMenu:
		switch m.bs.folderAction {
		case "add":
			return phaseFolderAdd
		case "test":
			return phaseFolderTest
		default:
			return phaseConnectionToggle
		}
	case phaseFolderAdd:
		return phaseFolderMenu
	case phaseFolderTest:
		return phaseFolderTestResult
	case phaseFolderTestResult:
		if m.bs.folderTestMore {
			return phaseFolderTest
		}
		return phaseFolderMenu

	case phaseConnectionToggle:
		if m.bs.configureConnections {
			return phaseFilterToggle
		}
		return phaseStorageToggle
	case phaseFilterToggle:
		if m.bs.addFilters {
			return phaseFilterInput
		}
		return phaseDefaultCreds
	case phaseFilterInput:
		if m.bs.addMoreFilters {
			return phaseFilterInput
		}
		return phaseDefaultCreds
	case phaseDefaultCreds:
		return phaseCredRuleToggle
	case phaseCredRuleToggle:
		if m.bs.addCredRules {
			return phaseCredRuleFile
		}
		return phaseStorageToggle
	case phaseCredRuleFile:
		return phaseCredRuleMatcher
	case phaseCredRuleMatcher:
		if m.bs.credAddMoreMatcher {
			return phaseCredRuleMatcher
		}
		return phaseCredRuleCreds
	case phaseCredRuleCreds:
		if m.bs.addMoreCredRules {
			return phaseCredRuleFile
		}
		return phaseStorageToggle

	case phaseStorageToggle:
		if m.bs.configureStorage {
			return phaseStorageProvider
		}
		return phaseDone
	case phaseStorageProvider:
		if cloudProvider(m.bs.storageProvider) != providerCustom {
			return phaseStorageProviderInfo
		}
		return phaseStorageCustomConfig
	case phaseStorageProviderInfo:
		return phaseDone
	case phaseStorageCustomConfig:
		return phaseStorageCustomCreds
	case phaseStorageCustomCreds:
		return phaseStorageCustomOptions
	case phaseStorageCustomOptions:
		return phaseStorageAssign
	case phaseStorageAssign:
		return phaseDone

	default:
		return phaseDone
	}
}

// ── Reverse phase transitions (Esc / back navigation) ─────────────────────────
//
// prevPhase is a pure mapping — it never mutates state.  Loop-index mutations
// are handled in Update() before this is called.

func (m *configBuilderModel) prevPhase() builderPhase {
	switch m.phase {

	// ── Plugin flow ───────────────────────────────────────────────────────
	case phasePluginSelect:
		return phasePluginToggle
	case phasePluginVersion:
		return phasePluginSelect
	case phasePluginConfig:
		// idx == 0 case; idx > 0 handled in Update.
		return phasePluginVersion

	// ── Auth / server ─────────────────────────────────────────────────────
	case phaseAuthType:
		// startPhase == phaseAuthType → caller cancels before reaching here.
		// We only arrive here when startPhase == phasePluginToggle.
		if !m.bs.configurePlugin {
			return phasePluginToggle
		}
		if m.bs.pluginLoadErr {
			return phasePluginSelect
		}
		if len(m.bs.pluginConfigFields) == 0 {
			return phasePluginVersion
		}
		// Update() will restore pluginConfigIdx to len-1 when this is returned.
		return phasePluginConfig

	case phaseBasicCreds:
		return phaseAuthType
	case phaseTokenCreds:
		if m.bs.authType == bothAuthForm.String() {
			return phaseBasicCreds
		}
		return phaseAuthType
	case phaseServerSettings:
		switch m.bs.authType {
		case basicAuthForm.String():
			return phaseBasicCreds
		case tokenAuthForm.String():
			return phaseTokenCreds
		case bothAuthForm.String():
			return phaseTokenCreds
		}
		return phaseAuthType

	// ── User settings ─────────────────────────────────────────────────────
	case phaseUserSettings:
		return phaseServerSettings
	case phaseUserLengths:
		return phaseUserSettings

	// ── Folder scope ──────────────────────────────────────────────────────
	case phaseFolderScope:
		if m.bs.userRandomPassword {
			return phaseUserLengths
		}
		return phaseUserSettings
	case phaseFolderMenu:
		return phaseFolderScope
	case phaseFolderAdd:
		return phaseFolderMenu
	case phaseFolderTest:
		return phaseFolderMenu
	case phaseFolderTestResult:
		return phaseFolderTest

	// ── Connections ───────────────────────────────────────────────────────
	case phaseConnectionToggle:
		if m.bs.folderScope == "all" {
			return phaseFolderScope
		}
		return phaseFolderMenu
	case phaseFilterToggle:
		return phaseConnectionToggle
	case phaseFilterInput:
		// Loop iterations not individually indexed; return to toggle.
		return phaseFilterToggle
	case phaseDefaultCreds:
		if m.bs.addFilters {
			return phaseFilterInput
		}
		return phaseFilterToggle
	case phaseCredRuleToggle:
		return phaseDefaultCreds
	case phaseCredRuleFile:
		if len(m.bs.credRules) > 0 {
			return phaseCredRuleCreds
		}
		return phaseCredRuleToggle
	case phaseCredRuleMatcher:
		return phaseCredRuleFile
	case phaseCredRuleCreds:
		return phaseCredRuleMatcher

	// ── Storage ───────────────────────────────────────────────────────────
	case phaseStorageToggle:
		if !m.bs.configureConnections {
			return phaseConnectionToggle
		}
		if m.bs.addCredRules {
			return phaseCredRuleCreds
		}
		return phaseCredRuleToggle
	case phaseStorageProvider:
		return phaseStorageToggle
	case phaseStorageProviderInfo:
		return phaseStorageProvider
	case phaseStorageCustomConfig:
		return phaseStorageProvider
	case phaseStorageCustomCreds:
		return phaseStorageCustomConfig
	case phaseStorageCustomOptions:
		return phaseStorageCustomCreds
	case phaseStorageAssign:
		return phaseStorageCustomOptions

	default:
		return m.phase
	}
}

// ── Validation helpers ────────────────────────────────────────────────────────

func validatePositiveInt(s string) error {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}
