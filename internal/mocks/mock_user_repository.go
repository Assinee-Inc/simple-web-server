package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUserEmail(emailUser string) *models.User {
	args := m.Called(emailUser)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.User)
}

func (m *MockUserRepository) Save(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(emailUser string) *models.User {
	args := m.Called(emailUser)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.User)
}

func (m *MockUserRepository) FindBySessionToken(token string) *models.User {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.User)
}

func (m *MockUserRepository) FindByStripeCustomerID(customerID string) *models.User {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.User)
}

func (m *MockUserRepository) FindByPasswordResetToken(token string) *models.User {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.User)
}

func (m *MockUserRepository) UpdatePasswordResetToken(user *models.User, token string) error {
	args := m.Called(user, token)
	return args.Error(0)
}
