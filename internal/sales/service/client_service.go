package service

import (
	"errors"

	accountrepo "github.com/anglesson/simple-web-server/internal/account/repository"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
)

type ClientService interface {
	FindCreatorsClientByID(clientID uint, creatorEmail string) (*salesmodel.Client, error)
	FindClientByPublicID(publicID string) (*salesmodel.Client, error)
	Update(input salesmodel.UpdateClientInput) (*salesmodel.Client, error)
	ExportClients(creatorEmail string) (*[]salesmodel.Client, error)
}

type clientServiceImpl struct {
	clientRepository  salesrepo.ClientRepository
	creatorRepository accountrepo.CreatorRepository
}

func NewClientService(
	clientRepository salesrepo.ClientRepository,
	creatorRepository accountrepo.CreatorRepository,
) ClientService {
	return &clientServiceImpl{
		clientRepository:  clientRepository,
		creatorRepository: creatorRepository,
	}
}

func (cs *clientServiceImpl) FindClientByPublicID(publicID string) (*salesmodel.Client, error) {
	if publicID == "" {
		return nil, errors.New("o id público do cliente deve ser informado")
	}
	return cs.clientRepository.FindByPublicID(publicID)
}

func (cs *clientServiceImpl) FindCreatorsClientByID(clientID uint, creatorEmail string) (*salesmodel.Client, error) {
	if clientID == 0 || creatorEmail == "" {
		return nil, errors.New("o id do cliente deve ser informado")
	}

	var client salesmodel.Client

	err := cs.clientRepository.FindByIDAndCreators(&client, clientID, creatorEmail)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (cs *clientServiceImpl) Update(input salesmodel.UpdateClientInput) (*salesmodel.Client, error) {
	if input.ID == 0 || input.EmailCreator == "" {
		return nil, errors.New("id do cliente e email do criador são obrigatórios")
	}

	client := &salesmodel.Client{}
	err := cs.clientRepository.FindByIDAndCreators(client, input.ID, input.EmailCreator)
	if err != nil {
		return nil, err
	}

	client.Email = input.Email
	client.Phone = input.Phone

	err = cs.clientRepository.Save(client)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (cs *clientServiceImpl) ExportClients(creatorEmail string) (*[]salesmodel.Client, error) {
	creator, err := cs.creatorRepository.FindCreatorByUserEmail(creatorEmail)
	if err != nil {
		return nil, err
	}
	return cs.clientRepository.FindClientsByPurchasesFromCreator(creator)
}

