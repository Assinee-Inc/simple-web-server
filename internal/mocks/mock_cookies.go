package mocks

import (
	"github.com/anglesson/simple-web-server/internal/handler/web"
	"github.com/stretchr/testify/mock"
)

type MockFlashMessage struct {
	mock.Mock
}

var _ web.FlashMessagePort = (*MockFlashMessage)(nil)

func (m *MockFlashMessage) Success(message string) {
	m.Called(message)
}

func (m *MockFlashMessage) Error(message string) {
	m.Called(message)
}
