package api

// Server info constants used by baseService.GetServerInfo and any callers.
const (
	SrvInfoDBKey               = "Database"
	SrvInfoCommitKey           = "Commit"
	SrvInfoVersionKey          = "Version"
	SrvInfoEnterpriseCommitKey = "EnterpriseCommit"
)

// GetServerInfo is promoted from baseService via DashNGoImpl embedding.
