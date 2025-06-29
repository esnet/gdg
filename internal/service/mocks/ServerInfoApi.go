// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	mock "github.com/stretchr/testify/mock"
)

// NewServerInfoApi creates a new instance of ServerInfoApi. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewServerInfoApi(t interface {
	mock.TestingT
	Cleanup(func())
}) *ServerInfoApi {
	mock := &ServerInfoApi{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// ServerInfoApi is an autogenerated mock type for the ServerInfoApi type
type ServerInfoApi struct {
	mock.Mock
}

type ServerInfoApi_Expecter struct {
	mock *mock.Mock
}

func (_m *ServerInfoApi) EXPECT() *ServerInfoApi_Expecter {
	return &ServerInfoApi_Expecter{mock: &_m.Mock}
}

// GetServerInfo provides a mock function for the type ServerInfoApi
func (_mock *ServerInfoApi) GetServerInfo() map[string]any {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetServerInfo")
	}

	var r0 map[string]any
	if returnFunc, ok := ret.Get(0).(func() map[string]any); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]any)
		}
	}
	return r0
}

// ServerInfoApi_GetServerInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetServerInfo'
type ServerInfoApi_GetServerInfo_Call struct {
	*mock.Call
}

// GetServerInfo is a helper method to define mock.On call
func (_e *ServerInfoApi_Expecter) GetServerInfo() *ServerInfoApi_GetServerInfo_Call {
	return &ServerInfoApi_GetServerInfo_Call{Call: _e.mock.On("GetServerInfo")}
}

func (_c *ServerInfoApi_GetServerInfo_Call) Run(run func()) *ServerInfoApi_GetServerInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ServerInfoApi_GetServerInfo_Call) Return(stringToV map[string]any) *ServerInfoApi_GetServerInfo_Call {
	_c.Call.Return(stringToV)
	return _c
}

func (_c *ServerInfoApi_GetServerInfo_Call) RunAndReturn(run func() map[string]any) *ServerInfoApi_GetServerInfo_Call {
	_c.Call.Return(run)
	return _c
}
