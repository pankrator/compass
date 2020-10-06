// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// RevocationListRepository is an autogenerated mock type for the RevocationListRepository type
type RevocationListRepository struct {
	mock.Mock
}

// Contains provides a mock function with given fields: hash
func (_m *RevocationListRepository) Contains(hash string) bool {
	ret := _m.Called(hash)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(hash)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Insert provides a mock function with given fields: hash
func (_m *RevocationListRepository) Insert(hash string) error {
	ret := _m.Called(hash)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(hash)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
