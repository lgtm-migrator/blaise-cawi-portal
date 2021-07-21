// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	authenticate "github.com/ONSdigital/blaise-cawi-portal/authenticate"
	busapi "github.com/ONSdigital/blaise-cawi-portal/busapi"

	mock "github.com/stretchr/testify/mock"
)

// JWTCryptoInterface is an autogenerated mock type for the JWTCryptoInterface type
type JWTCryptoInterface struct {
	mock.Mock
}

// DecryptJWT provides a mock function with given fields: _a0
func (_m *JWTCryptoInterface) DecryptJWT(_a0 interface{}) (*authenticate.UACClaims, error) {
	ret := _m.Called(_a0)

	var r0 *authenticate.UACClaims
	if rf, ok := ret.Get(0).(func(interface{}) *authenticate.UACClaims); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*authenticate.UACClaims)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EncryptJWT provides a mock function with given fields: _a0, _a1
func (_m *JWTCryptoInterface) EncryptJWT(_a0 string, _a1 *busapi.UacInfo) (string, error) {
	ret := _m.Called(_a0, _a1)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, *busapi.UacInfo) string); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *busapi.UacInfo) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}