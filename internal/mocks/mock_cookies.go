package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockFlashMessage struct {
	mock.Mock
}

func (m *MockFlashMessage) Success(message string) {
	m.Called(message)
}

func (m *MockFlashMessage) Error(message string) {
	m.Called(message)
}
