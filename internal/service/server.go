package service

import (
	"log"
)

const (
	SrvInfoDBKey               = "Database"
	SrvInfoCommitKey           = "Commit"
	SrvInfoVersionKey          = "Version"
	SrvInfoEnterpriseCommitKey = "EnterpriseCommit"
)

// GetServerInfo returns basic Grafana Server info
func (s *DashNGoImpl) GetServerInfo() map[string]any {
	response, err := s.GetClient().Health.GetHealth()
	if err != nil {
		log.Fatalf("Unable to get server health info, err: %v", err)
	}
	t := response.GetPayload()
	result := make(map[string]any)
	result[SrvInfoDBKey] = t.Database
	result[SrvInfoCommitKey] = t.Commit
	result[SrvInfoVersionKey] = t.Version
	result[SrvInfoEnterpriseCommitKey] = t.EnterpriseCommit

	return result
}
