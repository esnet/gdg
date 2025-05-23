// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	filters "github.com/esnet/gdg/internal/service/filters"
	mock "github.com/stretchr/testify/mock"

	models "github.com/grafana/grafana-openapi-client-go/models"

	types "github.com/esnet/gdg/internal/types"
)

// FoldersApi is an autogenerated mock type for the FoldersApi type
type FoldersApi struct {
	mock.Mock
}

type FoldersApi_Expecter struct {
	mock *mock.Mock
}

func (_m *FoldersApi) EXPECT() *FoldersApi_Expecter {
	return &FoldersApi_Expecter{mock: &_m.Mock}
}

// DeleteAllFolders provides a mock function with given fields: filter
func (_m *FoldersApi) DeleteAllFolders(filter filters.Filter) []string {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for DeleteAllFolders")
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

// FoldersApi_DeleteAllFolders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteAllFolders'
type FoldersApi_DeleteAllFolders_Call struct {
	*mock.Call
}

// DeleteAllFolders is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *FoldersApi_Expecter) DeleteAllFolders(filter interface{}) *FoldersApi_DeleteAllFolders_Call {
	return &FoldersApi_DeleteAllFolders_Call{Call: _e.mock.On("DeleteAllFolders", filter)}
}

func (_c *FoldersApi_DeleteAllFolders_Call) Run(run func(filter filters.Filter)) *FoldersApi_DeleteAllFolders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *FoldersApi_DeleteAllFolders_Call) Return(_a0 []string) *FoldersApi_DeleteAllFolders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FoldersApi_DeleteAllFolders_Call) RunAndReturn(run func(filters.Filter) []string) *FoldersApi_DeleteAllFolders_Call {
	_c.Call.Return(run)
	return _c
}

// DownloadFolderPermissions provides a mock function with given fields: filter
func (_m *FoldersApi) DownloadFolderPermissions(filter filters.Filter) []string {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for DownloadFolderPermissions")
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

// FoldersApi_DownloadFolderPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DownloadFolderPermissions'
type FoldersApi_DownloadFolderPermissions_Call struct {
	*mock.Call
}

// DownloadFolderPermissions is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *FoldersApi_Expecter) DownloadFolderPermissions(filter interface{}) *FoldersApi_DownloadFolderPermissions_Call {
	return &FoldersApi_DownloadFolderPermissions_Call{Call: _e.mock.On("DownloadFolderPermissions", filter)}
}

func (_c *FoldersApi_DownloadFolderPermissions_Call) Run(run func(filter filters.Filter)) *FoldersApi_DownloadFolderPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *FoldersApi_DownloadFolderPermissions_Call) Return(_a0 []string) *FoldersApi_DownloadFolderPermissions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FoldersApi_DownloadFolderPermissions_Call) RunAndReturn(run func(filters.Filter) []string) *FoldersApi_DownloadFolderPermissions_Call {
	_c.Call.Return(run)
	return _c
}

// DownloadFolders provides a mock function with given fields: filter
func (_m *FoldersApi) DownloadFolders(filter filters.Filter) []string {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for DownloadFolders")
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

// FoldersApi_DownloadFolders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DownloadFolders'
type FoldersApi_DownloadFolders_Call struct {
	*mock.Call
}

// DownloadFolders is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *FoldersApi_Expecter) DownloadFolders(filter interface{}) *FoldersApi_DownloadFolders_Call {
	return &FoldersApi_DownloadFolders_Call{Call: _e.mock.On("DownloadFolders", filter)}
}

func (_c *FoldersApi_DownloadFolders_Call) Run(run func(filter filters.Filter)) *FoldersApi_DownloadFolders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *FoldersApi_DownloadFolders_Call) Return(_a0 []string) *FoldersApi_DownloadFolders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FoldersApi_DownloadFolders_Call) RunAndReturn(run func(filters.Filter) []string) *FoldersApi_DownloadFolders_Call {
	_c.Call.Return(run)
	return _c
}

// ListFolderPermissions provides a mock function with given fields: filter
func (_m *FoldersApi) ListFolderPermissions(filter filters.Filter) map[*types.NestedHit][]*models.DashboardACLInfoDTO {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for ListFolderPermissions")
	}

	var r0 map[*types.NestedHit][]*models.DashboardACLInfoDTO
	if rf, ok := ret.Get(0).(func(filters.Filter) map[*types.NestedHit][]*models.DashboardACLInfoDTO); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[*types.NestedHit][]*models.DashboardACLInfoDTO)
		}
	}

	return r0
}

// FoldersApi_ListFolderPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListFolderPermissions'
type FoldersApi_ListFolderPermissions_Call struct {
	*mock.Call
}

// ListFolderPermissions is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *FoldersApi_Expecter) ListFolderPermissions(filter interface{}) *FoldersApi_ListFolderPermissions_Call {
	return &FoldersApi_ListFolderPermissions_Call{Call: _e.mock.On("ListFolderPermissions", filter)}
}

func (_c *FoldersApi_ListFolderPermissions_Call) Run(run func(filter filters.Filter)) *FoldersApi_ListFolderPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *FoldersApi_ListFolderPermissions_Call) Return(_a0 map[*types.NestedHit][]*models.DashboardACLInfoDTO) *FoldersApi_ListFolderPermissions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FoldersApi_ListFolderPermissions_Call) RunAndReturn(run func(filters.Filter) map[*types.NestedHit][]*models.DashboardACLInfoDTO) *FoldersApi_ListFolderPermissions_Call {
	_c.Call.Return(run)
	return _c
}

// ListFolders provides a mock function with given fields: filter
func (_m *FoldersApi) ListFolders(filter filters.Filter) []*types.NestedHit {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for ListFolders")
	}

	var r0 []*types.NestedHit
	if rf, ok := ret.Get(0).(func(filters.Filter) []*types.NestedHit); ok {
		r0 = rf(filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.NestedHit)
		}
	}

	return r0
}

// FoldersApi_ListFolders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListFolders'
type FoldersApi_ListFolders_Call struct {
	*mock.Call
}

// ListFolders is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *FoldersApi_Expecter) ListFolders(filter interface{}) *FoldersApi_ListFolders_Call {
	return &FoldersApi_ListFolders_Call{Call: _e.mock.On("ListFolders", filter)}
}

func (_c *FoldersApi_ListFolders_Call) Run(run func(filter filters.Filter)) *FoldersApi_ListFolders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *FoldersApi_ListFolders_Call) Return(_a0 []*types.NestedHit) *FoldersApi_ListFolders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FoldersApi_ListFolders_Call) RunAndReturn(run func(filters.Filter) []*types.NestedHit) *FoldersApi_ListFolders_Call {
	_c.Call.Return(run)
	return _c
}

// UploadFolderPermissions provides a mock function with given fields: filter
func (_m *FoldersApi) UploadFolderPermissions(filter filters.Filter) []string {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for UploadFolderPermissions")
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

// FoldersApi_UploadFolderPermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UploadFolderPermissions'
type FoldersApi_UploadFolderPermissions_Call struct {
	*mock.Call
}

// UploadFolderPermissions is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *FoldersApi_Expecter) UploadFolderPermissions(filter interface{}) *FoldersApi_UploadFolderPermissions_Call {
	return &FoldersApi_UploadFolderPermissions_Call{Call: _e.mock.On("UploadFolderPermissions", filter)}
}

func (_c *FoldersApi_UploadFolderPermissions_Call) Run(run func(filter filters.Filter)) *FoldersApi_UploadFolderPermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *FoldersApi_UploadFolderPermissions_Call) Return(_a0 []string) *FoldersApi_UploadFolderPermissions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FoldersApi_UploadFolderPermissions_Call) RunAndReturn(run func(filters.Filter) []string) *FoldersApi_UploadFolderPermissions_Call {
	_c.Call.Return(run)
	return _c
}

// UploadFolders provides a mock function with given fields: filter
func (_m *FoldersApi) UploadFolders(filter filters.Filter) []string {
	ret := _m.Called(filter)

	if len(ret) == 0 {
		panic("no return value specified for UploadFolders")
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

// FoldersApi_UploadFolders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UploadFolders'
type FoldersApi_UploadFolders_Call struct {
	*mock.Call
}

// UploadFolders is a helper method to define mock.On call
//   - filter filters.Filter
func (_e *FoldersApi_Expecter) UploadFolders(filter interface{}) *FoldersApi_UploadFolders_Call {
	return &FoldersApi_UploadFolders_Call{Call: _e.mock.On("UploadFolders", filter)}
}

func (_c *FoldersApi_UploadFolders_Call) Run(run func(filter filters.Filter)) *FoldersApi_UploadFolders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(filters.Filter))
	})
	return _c
}

func (_c *FoldersApi_UploadFolders_Call) Return(_a0 []string) *FoldersApi_UploadFolders_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FoldersApi_UploadFolders_Call) RunAndReturn(run func(filters.Filter) []string) *FoldersApi_UploadFolders_Call {
	_c.Call.Return(run)
	return _c
}

// NewFoldersApi creates a new instance of FoldersApi. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFoldersApi(t interface {
	mock.TestingT
	Cleanup(func())
}) *FoldersApi {
	mock := &FoldersApi{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
