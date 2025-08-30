package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Save(client *models.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *MockClientRepository) FindClientsByCreator(creator *models.Creator, query models.ClientFilter) (*[]models.Client, error) {
	args := m.Called(creator, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Client), args.Error(1)
}

func (m *MockClientRepository) FindByIDAndCreators(client *models.Client, clientID uint, creator string) error {
	args := m.Called(client, clientID, creator)
	return args.Error(0)
}

func (m *MockClientRepository) FindByClientsWhereEbookNotSend(creator *models.Creator, query models.ClientFilter) (*[]models.Client, error) {
	args := m.Called(creator, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Client), args.Error(1)
}

func (m *MockClientRepository) FindByClientsWhereEbookWasSend(creator *models.Creator, query models.ClientFilter) (*[]models.Client, error) {
	args := m.Called(creator, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]models.Client), args.Error(1)
}

func (m *MockClientRepository) InsertBatch(clients []*models.Client) error {
	args := m.Called(clients)
	return args.Error(0)
}

func (m *MockClientRepository) FindByEmail(email string) (*models.Client, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}
