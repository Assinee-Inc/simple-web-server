package mocks

import (
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	"github.com/stretchr/testify/mock"
)

type MockCreatorService struct {
	mock.Mock
}

func (m *MockCreatorService) CreateCreator(input accountmodel.InputCreateCreator) (*accountmodel.Creator, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountmodel.Creator), args.Error(1)
}

func (m *MockCreatorService) FindCreatorByEmail(email string) (*accountmodel.Creator, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountmodel.Creator), args.Error(1)
}

func (m *MockCreatorService) FindCreatorByUserID(userID uint) (*accountmodel.Creator, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountmodel.Creator), args.Error(1)
}

func (m *MockCreatorService) FindByID(id uint) (*accountmodel.Creator, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountmodel.Creator), args.Error(1)
}

func (m *MockCreatorService) UpdateCreator(creator *accountmodel.Creator) error {
	args := m.Called(creator)
	return args.Error(0)
}
