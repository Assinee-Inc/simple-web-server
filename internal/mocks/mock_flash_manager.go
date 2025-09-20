package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockFlashManager is a mock type for the FlashManager type
type MockFlashManager struct {
	mock.Mock
}

// AddFlash provides a mock function with given fields: w, r, message, key
func (_m *MockFlashManager) AddFlash(w http.ResponseWriter, r *http.Request, message string, key string) error {
	ret := _m.Called(w, r, message, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(http.ResponseWriter, *http.Request, string, string) error); ok {
		r0 = rf(w, r, message, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFlashes provides a mock function with given fields: w, r, key
func (_m *MockFlashManager) GetFlashes(w http.ResponseWriter, r *http.Request, key string) []string {
	ret := _m.Called(w, r, key)

	var r0 []string
	if rf, ok := ret.Get(0).(func(http.ResponseWriter, *http.Request, string) []string); ok {
		r0 = rf(w, r, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}
