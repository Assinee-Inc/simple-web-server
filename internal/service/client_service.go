package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/pkg/gov"

	"github.com/anglesson/simple-web-server/internal/models"
)

type ClientService interface {
	CreateClient(input models.CreateClientInput) (*models.CreateClientOutput, error)
	FindCreatorsClientByID(clientID uint, creatorEmail string) (*models.Client, error)
	Update(input models.UpdateClientInput) (*models.Client, error)
	CreateBatchClient(clients []*models.Client) error
}

type clientServiceImpl struct {
	clientRepository      repository.ClientRepository
	creatorRepository     repository.CreatorRepository
	receitaFederalService gov.ReceitaFederalService
}

func NewClientService(
	clientRepository repository.ClientRepository,
	creatorRepository repository.CreatorRepository,
	receitaFederalService gov.ReceitaFederalService,
) ClientService {
	return &clientServiceImpl{
		clientRepository:      clientRepository,
		creatorRepository:     creatorRepository,
		receitaFederalService: receitaFederalService,
	}
}

func (cs *clientServiceImpl) CreateClient(input models.CreateClientInput) (*models.CreateClientOutput, error) {
	// Validate input
	if err := validateClientInput(input); err != nil {
		return nil, err
	}

	creator, err := cs.creatorRepository.FindCreatorByUserEmail(input.EmailCreator)
	if err != nil {
		return nil, err
	}

	clientExists, err := cs.clientRepository.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if clientExists != nil {
		return nil, errors.New("cliente já existe")
	}

	// Clean CPF and phone
	cleanCPF := cleanCPF(input.CPF)
	cleanPhone := cleanPhone(input.Phone)

	// Parse birth date - try DD/MM/YYYY format first (from jmask), then YYYY-MM-DD
	var birthDate time.Time
	birthDate, err = time.Parse("02/01/2006", input.BirthDate)
	if err != nil {
		// If that fails, try YYYY-MM-DD format (from HTML date input)
		birthDate, err = time.Parse("2006-01-02", input.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("formato de data de nascimento inválido: %w", err)
		}
	}

	client := models.NewClient(input.Name, cleanCPF, input.BirthDate, input.Email, cleanPhone, creator)

	if err := cs.validateReceita(client, birthDate); err != nil {
		return nil, err
	}

	err = cs.clientRepository.Save(client)
	if err != nil {
		return nil, err
	}
	return &models.CreateClientOutput{
		ID:        int(client.ID),
		Name:      client.Name,
		CPF:       client.CPF,
		Phone:     client.Phone,
		Email:     client.Email,
		BirthDate: client.Birthdate,
		UpdatedAt: client.UpdatedAt.String(),
		CreatedAt: client.CreatedAt.String(),
	}, nil
}

func (cs *clientServiceImpl) FindCreatorsClientByID(clientID uint, creatorEmail string) (*models.Client, error) {
	if clientID == 0 || creatorEmail == "" {
		return nil, errors.New("o id do cliente deve ser informado")
	}

	var client models.Client

	err := cs.clientRepository.FindByIDAndCreators(&client, clientID, creatorEmail)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (cs *clientServiceImpl) Update(input models.UpdateClientInput) (*models.Client, error) {
	if input.ID == 0 || input.EmailCreator == "" {
		return nil, errors.New("id do cliente e email do criador são obrigatórios")
	}

	// Find the existing client
	client := &models.Client{}
	err := cs.clientRepository.FindByIDAndCreators(client, input.ID, input.EmailCreator)
	if err != nil {
		return nil, err
	}

	// Update only email and phone
	client.Email = input.Email
	client.Phone = input.Phone

	err = cs.clientRepository.Save(client)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (cs *clientServiceImpl) CreateBatchClient(clients []*models.Client) error {
	err := cs.clientRepository.InsertBatch(clients)
	if err != nil {
		return err
	}
	return nil
}

func (cs *clientServiceImpl) validateReceita(client *models.Client, birthDate time.Time) error {
	if cs.receitaFederalService == nil {
		return errors.New("serviço da receita federal não está disponível")
	}

	response, err := cs.receitaFederalService.ConsultaCPF(client.Name, client.CPF, birthDate.Format("02/01/2006"))
	if err != nil {
		return err
	}

	if !response.Status {
		return errors.New("CPF inválido ou não encontrado na receita federal")
	}

	if response.Result.NomeDaPF == "" {
		return errors.New("nome não encontrado na receita federal")
	}

	client.Name = response.Result.NomeDaPF
	client.Validated = true
	return nil
}
