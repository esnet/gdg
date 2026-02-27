package domain

// AlertRuleFilterParams defines the filtering criteria used to narrow down alert rules during import or export operations.
// It allows filtering by folder name, a set of labels, and provides an option to ignore folder-based rules.
type AlertRuleFilterParams struct {
	Folder               string
	Label                []string
	IgnoreWatchedFolders bool
}
