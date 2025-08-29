package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/stretchr/testify/mock"
)

type MockCreatorService struct {
	mock.Mock
}

func (m *MockCreatorService) CreateCreator(input service.InputCreateCreator) (*models.Creator, error) {
	args := m.Called(input)
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorService) FindCreatorByEmail(email string) (*models.Creator, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorService) FindCreatorByUserID(userID uint) (*models.Creator, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorService) FindByID(id uint) (*models.Creator, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorService) UpdateCreator(creator *models.Creator) error {
	// Mock implementation
	return nil
}
