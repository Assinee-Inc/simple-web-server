package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockCreatorRepository struct {
	mock.Mock
}

func (m *MockCreatorRepository) Update(creator *models.Creator) error {
	m.Called(creator)
	return nil
}

func (m *MockCreatorRepository) FindCreatorByUserID(userID uint) (*models.Creator, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorRepository) FindCreatorByUserEmail(email string) (*models.Creator, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorRepository) FindByCPF(cpf string) (*models.Creator, error) {
	args := m.Called(cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorRepository) FindByID(id uint) (*models.Creator, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorRepository) Create(creator *models.Creator) error {
	args := m.Called(creator)
	return args.Error(0)
}
