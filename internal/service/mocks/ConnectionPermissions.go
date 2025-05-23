// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	filters "github.com/esnet/gdg/internal/service/filters"
	mock "github.com/stretchr/testify/mock"

	types "github.com/esnet/gdg/internal/types"
)

// ConnectionPermissions is an autogenerated mock type for the ConnectionPermissions type
type ConnectionPermissions struct {
	mock.Mock
}

type ConnectionPermissions_Expecter struct {
	mock *mock.Mock
}

func (_m *ConnectionPermissions) EXPECT() *ConnectionPermissions_Expecter {
	return &ConnectionPermissions_Expecter{mock: &_m.Mock}
}

// DeleteAllConnectionPermissions provides a mock function with given fields: filter
func (_m *ConnectionPermissions) DeleteAllConnectionPermissions(filter filters.Filter) []string {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for DeleteAllConnectionPermissions")
	}

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

// ConnectionPermissions_DeleteAllConnectionPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteAllConnectionPermissions'
type ConnectionPermissions_DeleteAllConnectionPermissions_Call struct {
	*mock.Call
}

// DeleteAllConnectionPermissions is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *ConnectionPermissions_Expecter) DeleteAllConnectionPermissions(filter interface{}) *ConnectionPermissions_DeleteAllConnectionPermissions_Call {
	return &ConnectionPermissions_DeleteAllConnectionPermissions_Call{Call: _e.mock.On("DeleteAllConnectionPermissions", filter)}
}

func (_c *ConnectionPermissions_DeleteAllConnectionPermissions_Call) Run(run func(filter filters.Filter)) *ConnectionPermissions_DeleteAllConnectionPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *ConnectionPermissions_DeleteAllConnectionPermissions_Call) Return(_a0 []string) *ConnectionPermissions_DeleteAllConnectionPermissions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ConnectionPermissions_DeleteAllConnectionPermissions_Call) RunAndReturn(run func(filters.Filter) []string) *ConnectionPermissions_DeleteAllConnectionPermissions_Call {
	_c.Call.Return(run)
	return _c
}

// DownloadConnectionPermissions provides a mock function with given fields: filter
func (_m *ConnectionPermissions) DownloadConnectionPermissions(filter filters.Filter) []string {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for DownloadConnectionPermissions")
	}

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

// ConnectionPermissions_DownloadConnectionPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DownloadConnectionPermissions'
type ConnectionPermissions_DownloadConnectionPermissions_Call struct {
	*mock.Call
}

// DownloadConnectionPermissions is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *ConnectionPermissions_Expecter) DownloadConnectionPermissions(filter interface{}) *ConnectionPermissions_DownloadConnectionPermissions_Call {
	return &ConnectionPermissions_DownloadConnectionPermissions_Call{Call: _e.mock.On("DownloadConnectionPermissions", filter)}
}

func (_c *ConnectionPermissions_DownloadConnectionPermissions_Call) Run(run func(filter filters.Filter)) *ConnectionPermissions_DownloadConnectionPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *ConnectionPermissions_DownloadConnectionPermissions_Call) Return(_a0 []string) *ConnectionPermissions_DownloadConnectionPermissions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ConnectionPermissions_DownloadConnectionPermissions_Call) RunAndReturn(run func(filters.Filter) []string) *ConnectionPermissions_DownloadConnectionPermissions_Call {
	_c.Call.Return(run)
	return _c
}

// ListConnectionPermissions provides a mock function with given fields: filter
func (_m *ConnectionPermissions) ListConnectionPermissions(filter filters.Filter) []types.ConnectionPermissionItem {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for ListConnectionPermissions")
	}

	var r0 []types.ConnectionPermissionItem
	if rf, ok := ret.Get(0).(func(filters.Filter) []types.ConnectionPermissionItem); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.ConnectionPermissionItem)
		}
	}

	return r0
}

// ConnectionPermissions_ListConnectionPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListConnectionPermissions'
type ConnectionPermissions_ListConnectionPermissions_Call struct {
	*mock.Call
}

// ListConnectionPermissions is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *ConnectionPermissions_Expecter) ListConnectionPermissions(filter interface{}) *ConnectionPermissions_ListConnectionPermissions_Call {
	return &ConnectionPermissions_ListConnectionPermissions_Call{Call: _e.mock.On("ListConnectionPermissions", filter)}
}

func (_c *ConnectionPermissions_ListConnectionPermissions_Call) Run(run func(filter filters.Filter)) *ConnectionPermissions_ListConnectionPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *ConnectionPermissions_ListConnectionPermissions_Call) Return(_a0 []types.ConnectionPermissionItem) *ConnectionPermissions_ListConnectionPermissions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ConnectionPermissions_ListConnectionPermissions_Call) RunAndReturn(run func(filters.Filter) []types.ConnectionPermissionItem) *ConnectionPermissions_ListConnectionPermissions_Call {
	_c.Call.Return(run)
	return _c
}

// UploadConnectionPermissions provides a mock function with given fields: filter
func (_m *ConnectionPermissions) UploadConnectionPermissions(filter filters.Filter) []string {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for UploadConnectionPermissions")
	}

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

// ConnectionPermissions_UploadConnectionPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UploadConnectionPermissions'
type ConnectionPermissions_UploadConnectionPermissions_Call struct {
	*mock.Call
}

// UploadConnectionPermissions is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *ConnectionPermissions_Expecter) UploadConnectionPermissions(filter interface{}) *ConnectionPermissions_UploadConnectionPermissions_Call {
	return &ConnectionPermissions_UploadConnectionPermissions_Call{Call: _e.mock.On("UploadConnectionPermissions", filter)}
}

func (_c *ConnectionPermissions_UploadConnectionPermissions_Call) Run(run func(filter filters.Filter)) *ConnectionPermissions_UploadConnectionPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *ConnectionPermissions_UploadConnectionPermissions_Call) Return(_a0 []string) *ConnectionPermissions_UploadConnectionPermissions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ConnectionPermissions_UploadConnectionPermissions_Call) RunAndReturn(run func(filters.Filter) []string) *ConnectionPermissions_UploadConnectionPermissions_Call {
	_c.Call.Return(run)
	return _c
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
