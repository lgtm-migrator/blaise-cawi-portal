// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	gin "github.com/gin-gonic/gin"

	mock "github.com/stretchr/testify/mock"
)

// LanguageManagerInterface is an autogenerated mock type for the LanguageManagerInterface type
type LanguageManagerInterface struct {
	mock.Mock
}

// IsWelsh provides a mock function with given fields: _a0
func (_m *LanguageManagerInterface) IsWelsh(_a0 *gin.Context) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*gin.Context) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SetWelsh provides a mock function with given fields: _a0, _a1
func (_m *LanguageManagerInterface) SetWelsh(_a0 *gin.Context, _a1 bool) {
	_m.Called(_a0, _a1)
}
