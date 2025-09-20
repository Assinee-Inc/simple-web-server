package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockSessionService is a mock type for the SessionService type
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) Set(r *http.Request, w http.ResponseWriter, key string, value interface{}) error {
	args := m.Called(r, w, key, value)
	return args.Error(0)
}

func (m *MockSessionService) Get(r *http.Request, key string) interface{} {
	args := m.Called(r, key)
	return args.Get(0)
}

func (m *MockSessionService) Pop(r *http.Request, w http.ResponseWriter, key string) interface{} {
	args := m.Called(r, w, key)
	return args.Get(0)
}

func (m *MockSessionService) Destroy(r *http.Request, w http.ResponseWriter) error {
	args := m.Called(r, w)
	return args.Error(0)
}

func (m *MockSessionService) AddFlash(w http.ResponseWriter, r *http.Request, message, key string) error {
	args := m.Called(w, r, message, key)
	return args.Error(0)
}

func (m *MockSessionService) GetFlashes(w http.ResponseWriter, r *http.Request, key string) []string {
	args := m.Called(w, r, key)
	var r0 []string
	if args.Get(0) != nil {
		r0 = args.Get(0).([]string)
	}
	return r0
}

func (m *MockSessionService) RegenerateCSRFToken(r *http.Request, w http.ResponseWriter) (string, error) {
	args := m.Called(r, w)
	return args.String(0), args.Error(1)
}

func (m *MockSessionService) GetUserEmailFromSession(r *http.Request) (string, error) {
	args := m.Called(r)
	return args.String(0), args.Error(1)
}

func (m *MockSessionService) InitSession(w http.ResponseWriter, r *http.Request, userID uint, userEmail string) error {
	args := m.Called(w, r, userID, userEmail)
	return args.Error(0)
}
