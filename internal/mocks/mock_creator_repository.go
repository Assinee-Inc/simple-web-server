package mocks

import (
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	"github.com/stretchr/testify/mock"
)

type MockCreatorRepository struct {
	mock.Mock
}

func (m *MockCreatorRepository) Update(creator *accountmodel.Creator) error {
	m.Called(creator)
	return nil
}

func (m *MockCreatorRepository) FindCreatorByUserID(userID uint) (*accountmodel.Creator, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountmodel.Creator), args.Error(1)
}

func (m *MockCreatorRepository) FindCreatorByUserEmail(email string) (*accountmodel.Creator, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountmodel.Creator), args.Error(1)
}

func (m *MockCreatorRepository) FindByCPF(cpf string) (*accountmodel.Creator, error) {
	args := m.Called(cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountmodel.Creator), args.Error(1)
}

func (m *MockCreatorRepository) FindByID(id uint) (*accountmodel.Creator, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountmodel.Creator), args.Error(1)
}

func (m *MockCreatorRepository) Create(creator *accountmodel.Creator) error {
	args := m.Called(creator)
	return args.Error(0)
}
