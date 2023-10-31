// Code generated by mockery v2.34.0. DO NOT EDIT.

package mocks

import (
	filters "github.com/esnet/gdg/internal/service/filters"
	mock "github.com/stretchr/testify/mock"

	models "github.com/esnet/grafana-swagger-api-golang/goclient/models"
)

// ConnectionPermissions is an autogenerated mock type for the ConnectionPermissions type
type ConnectionPermissions struct {
	mock.Mock
}

// DeleteAllConnectionPermissions provides a mock function with given fields: filter
func (_m *ConnectionPermissions) DeleteAllConnectionPermissions(filter filters.Filter) []string {
	ret := _m.Called(filter)

	var r0 []string
	if rf, ok := ret.Get(0).(func(filters.Filter) []string); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// DownloadConnectionPermissions provides a mock function with given fields: filter
func (_m *ConnectionPermissions) DownloadConnectionPermissions(filter filters.Filter) []string {
	ret := _m.Called(filter)

	var r0 []string
	if rf, ok := ret.Get(0).(func(filters.Filter) []string); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// ListConnectionPermissions provides a mock function with given fields: filter
func (_m *ConnectionPermissions) ListConnectionPermissions(filter filters.Filter) map[*models.DataSourceListItemDTO]*models.DataSourcePermissionsDTO {
	ret := _m.Called(filter)

	var r0 map[*models.DataSourceListItemDTO]*models.DataSourcePermissionsDTO
	if rf, ok := ret.Get(0).(func(filters.Filter) map[*models.DataSourceListItemDTO]*models.DataSourcePermissionsDTO); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[*models.DataSourceListItemDTO]*models.DataSourcePermissionsDTO)
		}
	}

	return r0
}

// UploadConnectionPermissions provides a mock function with given fields: filter
func (_m *ConnectionPermissions) UploadConnectionPermissions(filter filters.Filter) []string {
	ret := _m.Called(filter)

	var r0 []string
	if rf, ok := ret.Get(0).(func(filters.Filter) []string); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// NewConnectionPermissions creates a new instance of ConnectionPermissions. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConnectionPermissions(t interface {
	mock.TestingT
	Cleanup(func())
}) *ConnectionPermissions {
	mock := &ConnectionPermissions{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}