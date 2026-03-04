package mocks

import (
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

var _ authsvc.UserService = (*MockUserService)(nil)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(input models.InputCreateUser) (*authmodel.User, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodel.User), args.Error(1)
}

func (m *MockUserService) AuthenticateUser(input models.InputLogin) (*authmodel.User, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodel.User), args.Error(1)
}

func (m *MockUserService) RequestPasswordReset(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(token, newPassword string) error {
	args := m.Called(token, newPassword)
	return args.Error(0)
}

func (m *MockUserService) FindByEmail(email string) *authmodel.User {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}
