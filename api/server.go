package api

import (
	log "github.com/sirupsen/logrus"
)

// GetServerInfo returns basic Grafana Server info
func (s *DashNGoImpl) GetServerInfo() map[string]interface{} {
	t, err := s.extended.Health()
	if err != nil {
		log.Panic("Unable to get server health info")
	}
	result := make(map[string]interface{})
	result["Database"] = t.Database
	result["Commit"] = t.Commit
	result["Version"] = t.Version

	return result

}
