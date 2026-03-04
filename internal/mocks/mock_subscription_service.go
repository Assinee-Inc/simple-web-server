package mocks

import (
	"time"

	"github.com/anglesson/simple-web-server/internal/subscription/model"
	"github.com/stretchr/testify/mock"
)

type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) CreateSubscription(userID uint, planID string) (*model.Subscription, error) {
	args := m.Called(userID, planID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) FindByUserID(userID uint) (*model.Subscription, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) FindByStripeCustomerID(customerID string) (*model.Subscription, error) {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) FindByStripeSubscriptionID(subscriptionID string) (*model.Subscription, error) {
	args := m.Called(subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) ActivateSubscription(subscription *model.Subscription, stripeCustomerID, stripeSubscriptionID string) error {
	args := m.Called(subscription, stripeCustomerID, stripeSubscriptionID)
	return args.Error(0)
}

func (m *MockSubscriptionService) UpdateSubscriptionStatus(subscription *model.Subscription, status string, endDate *time.Time) error {
	args := m.Called(subscription, status, endDate)
	return args.Error(0)
}

func (m *MockSubscriptionService) CancelSubscription(subscription *model.Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockSubscriptionService) EndTrial(subscription *model.Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func (m *MockSubscriptionService) GetUserSubscriptionStatus(userID uint) (string, int, error) {
	args := m.Called(userID)
	return args.String(0), args.Int(1), args.Error(2)
}
