package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockPaymentGateway struct {
	mock.Mock
}

func (m *MockPaymentGateway) CreateCustomer(email, name string) (string, error) {
	args := m.Called(email, name)
	return args.String(0), args.Error(1)
}

func (m *MockPaymentGateway) CreateSubscription(customerID, priceID string) (string, error) {
	args := m.Called(customerID, priceID)
	return args.String(0), args.Error(1)
}

func (m *MockPaymentGateway) CancelSubscription(subscriptionID string) error {
	args := m.Called(subscriptionID)
	return args.Error(0)
}
