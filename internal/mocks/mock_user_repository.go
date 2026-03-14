package mocks

import (
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	"github.com/stretchr/testify/mock"
)

var _ authrepo.UserRepository = (*MockUserRepository)(nil)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *authmodel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUserEmail(emailUser string) *authmodel.User {
	args := m.Called(emailUser)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserRepository) Save(user *authmodel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(emailUser string) *authmodel.User {
	args := m.Called(emailUser)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserRepository) FindBySessionToken(token string) *authmodel.User {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserRepository) FindByStripeCustomerID(customerID string) *authmodel.User {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserRepository) FindByPublicID(publicID string) *authmodel.User {
	args := m.Called(publicID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserRepository) FindByPasswordResetToken(token string) *authmodel.User {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserRepository) FindByEmailVerificationToken(token string) *authmodel.User {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserRepository) UpdatePasswordResetToken(user *authmodel.User, token string) error {
	args := m.Called(user, token)
	return args.Error(0)
}
