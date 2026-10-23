package mocks

import (
	"context"

	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/stretchr/testify/mock"
)

var _ authsvc.UserService = (*MockUserService)(nil)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) AuthenticateWithGoogle(ctx context.Context, idToken string) (*authmodel.User, bool, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*authmodel.User), args.Bool(1), args.Error(2)
}

func (m *MockUserService) FindByID(id uint) *authmodel.User {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserService) FindByEmail(email string) *authmodel.User {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authmodel.User)
}

func (m *MockUserService) Update(user *authmodel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) CreateUser(input authsvc.InputCreateUser) (uint, error) {
	args := m.Called(input)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockUserService) GenerateVerificationToken(email string) (string, error) {
	args := m.Called(email)
	return args.String(0), args.Error(1)
}
