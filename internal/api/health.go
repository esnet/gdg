package api

import (
	"context"
	"net/http"
)

// HealthResponse represents the response from the health endpoint, including commit hash, database status and version.
type HealthResponse struct {
	Commit   string `json:"commit,omitempty"`
	Database string `json:"database,omitempty"`
	Version  string `json:"version,omitempty"`
}

func (extended *ExtendedApi) Health() (*HealthResponse, error) {
	health := &HealthResponse{}
	err := extended.getRequestBuilder().
		Path("api/health").
		ToJSON(health).
		Method(http.MethodGet).Fetch(context.Background())
	return health, err
}
