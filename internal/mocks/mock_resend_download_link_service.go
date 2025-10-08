package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockResendDownloadLinkService struct {
	mock.Mock
}

func (m *MockResendDownloadLinkService) ResendDownloadLinkByTransactionID(transactionID uint) error {
	args := m.Called(transactionID)
	return args.Error(0)
}

func (m *MockResendDownloadLinkService) ResendDownloadLinkByPurchaseID(purchaseID uint, overrideEmail string) error {
	args := m.Called(purchaseID, overrideEmail)
	return args.Error(0)
}
