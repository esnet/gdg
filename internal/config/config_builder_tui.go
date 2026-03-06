package config

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports"
	"gopkg.in/yaml.v3"
)

// ── Builder TUI phases ────────────────────────────────────────────────────────

type builderPhase int

const (
	phaseAuthType builderPhase = iota
	phaseBasicCreds
	phaseTokenCreds
	phaseServerSettings
	phaseFolderScope
	phaseFolderInput
	phaseConnectionToggle
	phaseFilterToggle
	phaseFilterInput
	phaseDefaultCreds
	phaseCredRuleToggle
	phaseCredRuleInput
	phaseStorageToggle
	phaseStorageProvider
	phaseStorageCustomConfig
	phaseStorageCustomCreds
	phaseStorageCustomOptions
	phaseStorageAssign
	phaseDone
)

func (p builderPhase) sectionName() string {
	switch p {
	case phaseAuthType, phaseBasicCreds, phaseTokenCreds:
		return "Authentication"
	case phaseServerSettings:
		return "Server Settings"
	case phaseFolderScope, phaseFolderInput:
		return "Watched Folders"
	case phaseConnectionToggle, phaseFilterToggle, phaseFilterInput,
		phaseDefaultCreds, phaseCredRuleToggle, phaseCredRuleInput:
		return "Connection Settings"
	case phaseStorageToggle, phaseStorageProvider, phaseStorageCustomConfig,
		phaseStorageCustomCreds, phaseStorageCustomOptions, phaseStorageAssign:
		return "Cloud Storage"
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
	encoder     ports.CipherEncoder

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

	// Folders
	folderScope    string // "all" or "allowlist"
	folderName     string
	addMoreFolders bool
	folders        []string

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
	addCredRules     bool
	credSecureData   string
	credField        string
	credRegex        string
	credUser         string
	credPassword     string
	addMoreCredRules bool
	credRules        []*config_domain.RegexMatchesList
	pendingCreds     []pendingCredFile

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
}

func (s *builderState) renderPreview() string {
	type previewConfig struct {
		ContextName string                                  `yaml:"context_name"`
		Contexts    map[string]*config_domain.GrafanaConfig `yaml:"contexts"`
	}
	p := previewConfig{
		ContextName: s.contextName,
		Contexts:    map[string]*config_domain.GrafanaConfig{s.contextName: s.config},
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
	tuiFooterHeight = 1
	tuiMinBodyH     = 10
)

// ── Builder model ─────────────────────────────────────────────────────────────

type configBuilderModel struct {
	phase     builderPhase
	bs        *builderState
	form      *huh.Form
	width     int
	height    int
	done      bool
	cancelled bool
}

func newConfigBuilderModel(
	app *config_domain.GDGAppConfiguration,
	name string,
	encoder ports.CipherEncoder,
) configBuilderModel {
	bs := &builderState{
		contextName: name,
		app:         app,
		encoder:     encoder,
		config:      config_domain.NewGrafanaConfig(config_domain.WithContextName(name)),
		secure:      &config_domain.SecureModel{},
	}
	bs.config.OrganizationName = "Main Org."
	bs.config.ConnectionSettings = &config_domain.ConnectionSettings{
		MatchingRules: make([]*config_domain.RegexMatchesList, 0),
	}

	m := configBuilderModel{
		phase:  phaseAuthType,
		bs:     bs,
		width:  100,
		height: 30,
	}
	m.form = m.buildForm()
	return m
}

// ── tea.Model interface ───────────────────────────────────────────────────────

func (m configBuilderModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m configBuilderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.form.WithWidth(m.leftWidth())
		m.form.WithHeight(m.bodyHeight())
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
	}

	// Forward to embedded form
	newModel, cmd := m.form.Update(msg)
	m.form = newModel.(*huh.Form)

	if m.form.State == huh.StateCompleted {
		m.applyPhase()
		m.phase = m.nextPhase()
		if m.phase == phaseDone {
			m.done = true
			return m, tea.Quit
		}
		m.form = m.buildForm()
		m.form.WithWidth(m.leftWidth())
		m.form.WithHeight(m.bodyHeight())
		return m, m.form.Init()
	}

	if m.form.State == huh.StateAborted {
		m.cancelled = true
		return m, tea.Quit
	}

	return m, cmd
}

func (m configBuilderModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

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

	formView := m.form.View()
	leftPanel := lipgloss.NewStyle().
		Width(leftW).
		Height(bodyH).
		Padding(1, 2).
		Render(formView)

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
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(m.width).
		Padding(0, 2).
		Render("ctrl+c: cancel  •  enter: submit  •  ↑/↓: navigate")

	return lipgloss.JoinVertical(lipgloss.Left, header, stepInfo, body, footer)
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

// ── Form creation per phase ───────────────────────────────────────────────────

func (m *configBuilderModel) buildForm() *huh.Form {
	switch m.phase {
	case phaseAuthType:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Authentication Type").
					Description("How should GDG authenticate with Grafana?").
					Options(
						huh.NewOption("Basic Authentication", basicAuthForm.String()),
						huh.NewOption("Token/Service Authentication", tokenAuthForm.String()),
						huh.NewOption("Both", bothAuthForm.String()),
					).
					Value(&m.bs.authType),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseBasicCreds:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Grafana Username").
					Description("Username for basic authentication").
					Value(&m.bs.userName),
				huh.NewInput().
					Title("Grafana Password").
					EchoMode(huh.EchoModePassword).
					Value(&m.bs.password),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseTokenCreds:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Grafana Token").
					Description("API token or service account token").
					EchoMode(huh.EchoModePassword).
					Value(&m.bs.token),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseServerSettings:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Grafana URL").
					Description("Include the scheme (e.g. http://grafana.example.com)").
					Validate(validateGrafanaURL).
					Value(&m.bs.url),
				huh.NewInput().
					Title("Output Path").
					Description("Local folder for storing backups").
					Value(&m.bs.outputPath),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseFolderScope:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Folder Scope").
					Description("Monitor all Grafana folders or restrict to a specific allowlist?").
					Options(
						huh.NewOption("Monitor all folders (ignore filters)", "all"),
						huh.NewOption("Build a folder allowlist", "allowlist"),
					).
					Value(&m.bs.folderScope),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseFolderInput:
		currentList := "none"
		if len(m.bs.folders) > 0 {
			currentList = strings.Join(m.bs.folders, ", ")
		}
		m.bs.folderName = ""
		m.bs.addMoreFolders = true
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Folder Name or Regex").
					Description(fmt.Sprintf("Current: [%s]\nEnter a folder name (auto-encoded) or regex pattern.", currentList)).
					Value(&m.bs.folderName),
				huh.NewConfirm().
					Title("Add another folder after this one?").
					Value(&m.bs.addMoreFolders),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseConnectionToggle:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Configure connection settings?").
					Description("Set up data source filters, default credentials, and credential rules.\nYou can skip this and configure later by editing gdg.yml.").
					Value(&m.bs.configureConnections),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseFilterToggle:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Configure connection filters?").
					Description("Filters control which data sources are included or excluded.").
					Value(&m.bs.addFilters),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseFilterInput:
		desc := fmt.Sprintf("Filters so far: %s", summariseFilters(m.bs.filters))
		m.bs.filterField = ""
		m.bs.filterRegex = ""
		m.bs.filterInclusive = false
		m.bs.addMoreFilters = false
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Filter Field").
					Description(desc+"\nJSON field to match (e.g. 'name', 'type', 'url')").
					Value(&m.bs.filterField),
				huh.NewInput().
					Title("Filter Regex").
					Description("Regular expression to match against the field value").
					Value(&m.bs.filterRegex),
				huh.NewConfirm().
					Title("Inclusive filter? (allowlist)").
					Description("Yes = keep only matches. No = drop matches (denylist).").
					Value(&m.bs.filterInclusive),
				huh.NewConfirm().
					Title("Add another filter?").
					Value(&m.bs.addMoreFilters),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseDefaultCreds:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Connection Default User").
					Description("Default user for data source connections").
					Value(&m.bs.connectionUser),
				huh.NewInput().
					Title("Connection Default Password").
					EchoMode(huh.EchoModePassword).
					Value(&m.bs.connectionPassword),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseCredRuleToggle:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Configure credential rules?").
					Description("Map data sources to specific credential files.\nA default catch-all rule (name=.*) is always appended.").
					Value(&m.bs.addCredRules),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseCredRuleInput:
		m.bs.credSecureData = ""
		m.bs.credField = ""
		m.bs.credRegex = ""
		m.bs.credUser = ""
		m.bs.credPassword = ""
		m.bs.addMoreCredRules = false
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Secure Data File").
					Description("Credentials filename (e.g. 'elastic.yaml'). 'default.yaml' is reserved.").
					Validate(validateSecureDataFile).
					Value(&m.bs.credSecureData),
				huh.NewInput().
					Title("Matching Field").
					Description("Field to match (e.g. 'name', 'url', 'type')").
					Value(&m.bs.credField),
				huh.NewInput().
					Title("Matching Regex").
					Description("Regular expression to match against the field value").
					Value(&m.bs.credRegex),
				huh.NewInput().
					Title("User for this credential file").
					Value(&m.bs.credUser),
				huh.NewInput().
					Title("Password for this credential file").
					EchoMode(huh.EchoModePassword).
					Value(&m.bs.credPassword),
				huh.NewConfirm().
					Title("Add another credential rule?").
					Value(&m.bs.addMoreCredRules),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseStorageToggle:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Configure cloud storage?").
					Description("Set up a cloud storage engine for this context.\nYou can skip and add one later via the config file.").
					Value(&m.bs.configureStorage),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseStorageProvider:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Cloud Storage Provider").
					Description("For AWS S3, GCS, and Azure, GDG delegates auth to the provider SDK.\nOnly custom S3-compatible endpoints are configured here.").
					Options(
						huh.NewOption("Custom S3-compatible (Minio, Ceph, ...)", string(providerCustom)),
						huh.NewOption("AWS S3", string(providerAWS)),
						huh.NewOption("Google Cloud Storage (GCS)", string(providerGCS)),
						huh.NewOption("Azure Blob Storage", string(providerAzure)),
					).
					Value(&m.bs.storageProvider),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseStorageCustomConfig:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Storage Engine Label").
					Description("Unique key for this config in gdg.yml (e.g. my-minio)").
					Value(&m.bs.storageLabel),
				huh.NewInput().
					Title("Endpoint URL").
					Description("Full URL of the S3-compatible endpoint (e.g. http://localhost:9000)").
					Value(&m.bs.storageEndpoint),
				huh.NewInput().
					Title("Bucket Name").
					Value(&m.bs.storageBucket),
				huh.NewInput().
					Title("Region").
					Description("AWS region or equivalent (default: us-east-1)").
					Value(&m.bs.storageRegion),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseStorageCustomCreds:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Access Key ID").
					Value(&m.bs.storageAccessID),
				huh.NewInput().
					Title("Secret Access Key").
					EchoMode(huh.EchoModePassword).
					Value(&m.bs.storageSecretKey),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseStorageCustomOptions:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Path Prefix (optional)").
					Description("Prefix applied to all object paths within the bucket").
					Value(&m.bs.storagePrefix),
				huh.NewConfirm().
					Title("Auto-create bucket if it does not exist?").
					Value(&m.bs.storageInitBucket),
				huh.NewConfirm().
					Title("Enable SSL?").
					Value(&m.bs.storageSSL),
			),
		).WithShowHelp(false).WithShowErrors(true)

	case phaseStorageAssign:
		ctx := m.bs.app.GetContext()
		return huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Assign this storage engine to the active context (%q)?", ctx)).
					Value(&m.bs.storageAssign),
			),
		).WithShowHelp(false).WithShowErrors(true)

	default:
		return huh.NewForm(huh.NewGroup())
	}
}

// ── Apply results from completed form ─────────────────────────────────────────

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

	case phaseFolderScope:
		if m.bs.folderScope == "all" {
			if m.bs.config.DashboardSettings == nil {
				m.bs.config.DashboardSettings = &config_domain.DashboardSettings{}
			}
			m.bs.config.DashboardSettings.IgnoreFilters = true
		}

	case phaseFolderInput:
		name := strings.TrimSpace(m.bs.folderName)
		if name != "" {
			if looksLikeRegex(name) {
				m.bs.folders = append(m.bs.folders, name)
			} else {
				m.bs.folders = append(m.bs.folders, encodeFolderName(name))
			}
		}
		// Always update config so preview reflects current state
		if len(m.bs.folders) > 0 {
			m.bs.config.MonitoredFolders = m.bs.folders
		}
		// If we're done (no more to add, or empty name), finalize
		if !m.bs.addMoreFolders || name == "" {
			if len(m.bs.folders) == 0 {
				m.bs.folders = []string{"General"}
			}
			m.bs.config.MonitoredFolders = m.bs.folders
		}

	case phaseConnectionToggle:
		// configureConnections already bound

	case phaseFilterToggle:
		// addFilters already bound

	case phaseFilterInput:
		rule, err := validateFilter(m.bs.filterField, m.bs.filterRegex, m.bs.filterInclusive)
		if err == nil {
			m.bs.filters = append(m.bs.filters, *rule)
		}
		// Update config for live preview
		m.bs.config.ConnectionSettings.FilterRules = m.bs.filters

	case phaseDefaultCreds:
		// connectionUser/connectionPassword stored for file write after TUI exits

	case phaseCredRuleToggle:
		// addCredRules already bound

	case phaseCredRuleInput:
		rule, err := validateCredentialRule(m.bs.credField, m.bs.credRegex)
		if err == nil {
			m.bs.credRules = append(m.bs.credRules, &config_domain.RegexMatchesList{
				Rules:      []config_domain.MatchingRule{*rule},
				SecureData: strings.TrimSpace(m.bs.credSecureData),
			})
			m.bs.pendingCreds = append(m.bs.pendingCreds, pendingCredFile{
				secureData: strings.TrimSpace(m.bs.credSecureData),
				user:       m.bs.credUser,
				password:   m.bs.credPassword,
			})
		}
		// Update config for live preview (always include the default catch-all)
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
	}
}

// ── Phase transitions ─────────────────────────────────────────────────────────

func (m *configBuilderModel) nextPhase() builderPhase {
	switch m.phase {
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
		return phaseFolderScope

	case phaseFolderScope:
		if m.bs.folderScope == "all" {
			return phaseConnectionToggle
		}
		return phaseFolderInput

	case phaseFolderInput:
		name := strings.TrimSpace(m.bs.folderName)
		if m.bs.addMoreFolders && name != "" {
			return phaseFolderInput // loop
		}
		return phaseConnectionToggle

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
			return phaseFilterInput // loop
		}
		return phaseDefaultCreds

	case phaseDefaultCreds:
		return phaseCredRuleToggle

	case phaseCredRuleToggle:
		if m.bs.addCredRules {
			return phaseCredRuleInput
		}
		return phaseStorageToggle

	case phaseCredRuleInput:
		if m.bs.addMoreCredRules {
			return phaseCredRuleInput // loop
		}
		return phaseStorageToggle

	case phaseStorageToggle:
		if m.bs.configureStorage {
			return phaseStorageProvider
		}
		return phaseDone

	case phaseStorageProvider:
		if cloudProvider(m.bs.storageProvider) != providerCustom {
			return phaseDone
		}
		return phaseStorageCustomConfig

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
