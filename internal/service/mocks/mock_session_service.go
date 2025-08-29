package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) GenerateSessionToken() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSessionService) GenerateCSRFToken() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSessionService) SetSessionToken(w http.ResponseWriter) {
	m.Called(w)
}

func (m *MockSessionService) SetCSRFToken(w http.ResponseWriter) {
	m.Called(w)
}

func (m *MockSessionService) ClearSessionToken(w http.ResponseWriter) {
	m.Called(w)
}

func (m *MockSessionService) ClearCSRFToken(w http.ResponseWriter) {
	m.Called(w)
}

func (m *MockSessionService) GetSessionToken(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

func (m *MockSessionService) GetCSRFToken(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

func (m *MockSessionService) ClearSession(w http.ResponseWriter) {
	m.Called(w)
}

func (m *MockSessionService) SetSession(w http.ResponseWriter) {
	m.Called(w)
}

func (m *MockSessionService) GetSession(w http.ResponseWriter, r *http.Request) (string, string) {
	args := m.Called(w, r)
	return args.String(0), args.String(1)
}

func (m *MockSessionService) InitSession(w http.ResponseWriter, email string) {
	m.Called(w, email)
}

func (m *MockSessionService) GetUserEmailFromSession(r *http.Request) (string, error) {
	args := m.Called(r)
	return args.String(0), args.Error(1)
}
