package api

import (
	"context"
	"fmt"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/datasource_permissions"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"net/http"
)

// AddConnectionPermission adds permission to a given data source
// Notes:  The Swagger spec is incorrect.  It lists all parameters as query parameters while they should actually
// be passed in as part of the JSON body.
func (extended *ExtendedApi) AddConnectionPermission(p *datasource_permissions.AddPermissionParams) error {
	response := new(models.AddPermissionOKBody)

	url := fmt.Sprintf("/api/datasources/%s/permissions", p.DatasourceID)
	req := map[string]interface{}{
		"permission": int(*p.Permission),
		"userId":     int(*p.UserID),
		"teamId":     int(*p.TeamID),
	}
	if p.BuiltinRole != nil {
		req["builtinRole"] = *p.BuiltinRole
	}
	err := extended.getRequestBuilder().
		Path(url).
		BodyJSON(&req).
		Accept("application/json").
		//ToString(&ans).
		ToJSON(response).
		Method(http.MethodPost).Fetch(context.Background())

	return err

}
