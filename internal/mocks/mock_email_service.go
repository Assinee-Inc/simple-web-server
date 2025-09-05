package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendPasswordResetEmail(name, email string, resetLink string) {
	m.Called(name, email, resetLink)
}

func (m *MockEmailService) SendAccountConfirmation(name, email, token string) {
	m.Called(name, email, token)
}

func (m *MockEmailService) SendLinkToDownload(purchases []*models.Purchase) {
	m.Called(purchases)
}
