package mocks

import (
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/mock"
)

type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Save(client *salesmodel.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *MockClientRepository) FindClientsByCreator(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error) {
	args := m.Called(creator, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]salesmodel.Client), args.Error(1)
}

func (m *MockClientRepository) FindByIDAndCreators(client *salesmodel.Client, clientID uint, creator string) error {
	args := m.Called(client, clientID, creator)
	return args.Error(0)
}

func (m *MockClientRepository) FindByPublicID(publicID string) (*salesmodel.Client, error) {
	args := m.Called(publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Client), args.Error(1)
}

func (m *MockClientRepository) FindByClientsWhereEbookNotSend(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error) {
	args := m.Called(creator, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]salesmodel.Client), args.Error(1)
}

func (m *MockClientRepository) FindByClientsWhereEbookWasSend(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error) {
	args := m.Called(creator, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]salesmodel.Client), args.Error(1)
}

func (m *MockClientRepository) FindClientsByPurchasesFromCreator(creator *accountmodel.Creator) (*[]salesmodel.Client, error) {
	args := m.Called(creator)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]salesmodel.Client), args.Error(1)
}

func (m *MockClientRepository) FindByEmail(email string) (*salesmodel.Client, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Client), args.Error(1)
}

func (m *MockClientRepository) FindByCPF(cpf string) (*salesmodel.Client, error) {
	args := m.Called(cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*salesmodel.Client), args.Error(1)
}
