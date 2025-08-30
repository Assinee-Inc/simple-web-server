package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockClientService struct {
	mock.Mock
}

func NewMockClientService() *MockClientService {
	return &MockClientService{}
}

func (m *MockClientService) CreateClient(input models.CreateClientInput) (*models.CreateClientOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CreateClientOutput), args.Error(1)
}

func (m *MockClientService) FindCreatorsClientByID(clientID uint, creatorEmail string) (*models.Client, error) {
	args := m.Called(clientID, creatorEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientService) Update(input models.UpdateClientInput) (*models.Client, error) {
	args := m.Called(input)
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientService) CreateBatchClient(clients []*models.Client) error {
	args := m.Called(clients)
	return args.Error(1)
}
