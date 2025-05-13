package service

import (
	"log"
)

// GetServerInfo returns basic Grafana Server info
func (s *DashNGoImpl) GetServerInfo() map[string]any {
	t, err := s.extended.Health()
	if err != nil {
		log.Fatalf("Unable to get server health info, err: %v", err)
	}
	result := make(map[string]any)
	result["Database"] = t.Database
	result["Commit"] = t.Commit
	result["Version"] = t.Version

	return result
}
