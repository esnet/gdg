// Code generated by mockery v2.42.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// LicenseApi is an autogenerated mock type for the LicenseApi type
type LicenseApi struct {
	mock.Mock
}

type LicenseApi_Expecter struct {
	mock *mock.Mock
}

func (_m *LicenseApi) EXPECT() *LicenseApi_Expecter {
	return &LicenseApi_Expecter{mock: &_m.Mock}
}

// IsEnterprise provides a mock function with given fields:
func (_m *LicenseApi) IsEnterprise() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsEnterprise")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// LicenseApi_IsEnterprise_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEnterprise'
type LicenseApi_IsEnterprise_Call struct {
	*mock.Call
}

// IsEnterprise is a helper method to define mock.On call
func (_e *LicenseApi_Expecter) IsEnterprise() *LicenseApi_IsEnterprise_Call {
	return &LicenseApi_IsEnterprise_Call{Call: _e.mock.On("IsEnterprise")}
}

func (_c *LicenseApi_IsEnterprise_Call) Run(run func()) *LicenseApi_IsEnterprise_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LicenseApi_IsEnterprise_Call) Return(_a0 bool) *LicenseApi_IsEnterprise_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LicenseApi_IsEnterprise_Call) RunAndReturn(run func() bool) *LicenseApi_IsEnterprise_Call {
	_c.Call.Return(run)
	return _c
}

// NewLicenseApi creates a new instance of LicenseApi. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLicenseApi(t interface {
	mock.TestingT
	Cleanup(func())
}) *LicenseApi {
	mock := &LicenseApi{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
