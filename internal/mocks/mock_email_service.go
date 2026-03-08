package mocks

import (
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesdto "github.com/anglesson/simple-web-server/internal/sales/service/dto"
	"github.com/stretchr/testify/mock"
)

// MockAuthEmailService mocks authsvc.IEmailService
type MockAuthEmailService struct {
	mock.Mock
}

func (m *MockAuthEmailService) SendPasswordResetEmail(name, email string, resetLink string) {
	m.Called(name, email, resetLink)
}

func (m *MockAuthEmailService) SendAccountConfirmation(name, email, token string) {
	m.Called(name, email, token)
}

// MockSalesEmailService mocks salesvc.IEmailService
type MockSalesEmailService struct {
	mock.Mock
}

func (m *MockSalesEmailService) SendLinkToDownload(purchases []*salesmodel.Purchase) {
	m.Called(purchases)
}

func (m *MockSalesEmailService) ResendDownloadLink(downloadDTO *salesdto.ResendDownloadLinkDTO) error {
	args := m.Called(downloadDTO)
	return args.Error(0)
}
