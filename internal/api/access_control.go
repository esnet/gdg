package api

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

const connectionPermissionsUri = "api/access-control/datasources/%s"
const connectRevokeUserPermissionUri = "api/access-control/datasources/%s/users/%d"

type ConnectionPermissions struct {
	ID               int      `json:"id"`
	RoleName         string   `json:"roleName"`
	IsManaged        bool     `json:"isManaged"`
	IsInherited      bool     `json:"isInherited"`
	IsServiceAccount bool     `json:"isServiceAccount"`
	BuiltInRole      string   `json:"builtInRole"`
	Actions          []string `json:"actions"`
	Permission       string   `json:"permission"`
}
type Permission string

const (
	NoPermission    Permission = ""
	QueryPermission Permission = "Query"
	EditPermission  Permission = "Edit"
	AdminPermission Permission = "Admin"
)

type PermissionOperation struct {
	Permission Permission `json:"permission"`
}

func (extended *ExtendedApi) GetConnectionPermission(connectionUid string) ([]*ConnectionPermissions, error) {
	var result []*ConnectionPermissions

	err := extended.getRequestBuilder().
		Path(fmt.Sprintf(connectionPermissionsUri, connectionUid)).
		ToJSON(&result).
		Method(http.MethodGet).
		Fetch(context.Background())

	return nil, err
}

// Revoke or Add Connection Access for User
func (extended *ExtendedApi) UpdateUserAccessPermission(connectionUid string, userId int64, permission Permission) error {
	permissionOp := PermissionOperation{Permission: NoPermission}
	if permission != "" {
		permissionOp = PermissionOperation{Permission: permission}
	}
	var buf = bytes.Buffer{}

	err := extended.getRequestBuilder().
		Path(fmt.Sprintf(connectRevokeUserPermissionUri, connectionUid, userId)).
		Header("Content-Type", "application/json").
		Header("Accepts", "application/json").
		BodyJSON(&permissionOp).
		Method(http.MethodPost).
		ToBytesBuffer(&buf).
		Fetch(context.Background())

	fmt.Println(buf.String())

	return err
}
