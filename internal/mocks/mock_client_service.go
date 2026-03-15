package mocks

import (
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/mock"
)

type MockClientService struct {
	mock.Mock
}

func NewMockClientService() *MockClientService {
	return &MockClientService{}
}

func (m *MockClientService) FindClientByPublicID(publicID string) (*salesmodel.Client, error) {
	args := m.Called(publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Client), args.Error(1)
}

func (m *MockClientService) FindCreatorsClientByID(clientID uint, creatorEmail string) (*salesmodel.Client, error) {
	args := m.Called(clientID, creatorEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Client), args.Error(1)
}

func (m *MockClientService) Update(input salesmodel.UpdateClientInput) (*salesmodel.Client, error) {
	args := m.Called(input)
	return args.Get(0).(*salesmodel.Client), args.Error(1)
}

func (m *MockClientService) ExportClients(creatorEmail string) (*[]salesmodel.Client, error) {
	args := m.Called(creatorEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]salesmodel.Client), args.Error(1)
}
