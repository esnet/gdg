package apiExtend

import (
	"context"
	"fmt"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/users"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"net/http"
)

func (extended *ExtendedApi) SearchUsers(query *users.SearchUsersParams) ([]*models.UserSearchHitDTO, error) {
	usersList := make([]*models.UserSearchHitDTO, 0)
	err := extended.req.
		Path("/api/users").
		Param("page", fmt.Sprintf("%d", *query.Page)).
		Param("perpage", fmt.Sprintf("%d", *query.Perpage)).
		ToJSON(&usersList).
		Method(http.MethodGet).Fetch(context.Background())
	return usersList, err

}
