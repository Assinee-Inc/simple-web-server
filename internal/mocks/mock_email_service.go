package mocks

import (
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	salesdto "github.com/Assinee-Inc/docffy/internal/sales/service/dto"
	"github.com/stretchr/testify/mock"
)

// MockAuthEmailService mocks authsvc.IEmailService
type MockAuthEmailService struct {
	mock.Mock
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
