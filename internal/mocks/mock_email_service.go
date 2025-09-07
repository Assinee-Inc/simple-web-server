package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service/dto"
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

func (m *MockEmailService) ResendDownloadLink(downloadDTO *dto.ResendDownloadLinkDTO) error {
	args := m.Called(downloadDTO)
	return args.Error(0)
}
