package types

// LintRequest Dashboard Linting request
type LintRequest struct {
	StrictFlag    bool
	VerboseFlag   bool
	AutoFix       bool
	DashboardSlug string
	FolderName    string
}
