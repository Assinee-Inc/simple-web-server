package service_test

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/service"

	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/pkg/gov"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/suite"
)

var _ repository.ClientRepository = (*mocks.MockClientRepository)(nil)
var _ repository.CreatorRepository = (*mocks.MockCreatorRepository)(nil)
var _ gov.ReceitaFederalService = (*mocks.MockRFService)(nil)

type ClientServiceTestSuite struct {
	suite.Suite
	sut                   service.ClientService
	mockClientRepository  repository.ClientRepository
	mockCreatorRepository repository.CreatorRepository
	mockRFService         gov.ReceitaFederalService
}

func (suite *ClientServiceTestSuite) SetupTest() {
	suite.mockClientRepository = new(mocks.MockClientRepository)
	suite.mockCreatorRepository = new(mocks.MockCreatorRepository)
	suite.mockRFService = new(mocks.MockRFService)
	suite.sut = service.NewClientService(suite.mockClientRepository, suite.mockCreatorRepository, suite.mockRFService)
}

func (suite *ClientServiceTestSuite) TestCreateClient() {
	creator := &models.Creator{Email: "creator@mail.com"}

	input := models.CreateClientInput{
		Name:         "Name User",
		CPF:          "058.997.950-77",
		BirthDate:    "1990-12-12",
		Email:        "client@mail.com",
		Phone:        "12945678901",
		EmailCreator: creator.Email,
	}

	expectedName := "Name Receita Federal"
	expectedBirthDay := "12/12/1990"

	client := &models.Client{
		Validated: true,
		Name:      expectedName,
		CPF:       "05899795077",
		Birthdate: input.BirthDate,
		Email:     input.Email,
		Phone:     input.Phone,
		Creators:  []*models.Creator{creator},
	}

	suite.mockRFService.(*mocks.MockRFService).
		On("ConsultaCPF", "05899795077", "12/12/1990").
		Return(&gov.ReceitaFederalResponse{
			Status: true,
			Result: gov.ConsultaData{
				NomeDaPF:       expectedName,
				NumeroDeCPF:    input.CPF,
				DataNascimento: expectedBirthDay,
			},
		}, nil)

	suite.mockCreatorRepository.(*mocks.MockCreatorRepository).
		On("FindCreatorByUserEmail", creator.Email).
		Return(creator, nil)

	suite.mockClientRepository.(*mocks.MockClientRepository).
		On("FindByEmail", input.Email).
		Return(nil, nil)

	suite.mockClientRepository.(*mocks.MockClientRepository).
		On("Save", client).
		Return(nil)

	_, err := suite.sut.CreateClient(input)

	suite.NoError(err)
	suite.mockCreatorRepository.(*mocks.MockCreatorRepository).AssertCalled(suite.T(), "FindCreatorByUserEmail", creator.Email)
	suite.mockClientRepository.(*mocks.MockClientRepository).AssertCalled(suite.T(), "Save", client)
}

func (suite *ClientServiceTestSuite) TestShouldReturnErrorIfClientExists() {
	creator := &models.Creator{Email: "creator@mail.com"}

	input := models.CreateClientInput{
		Name:         "Name User",
		CPF:          "058.997.950-77",
		BirthDate:    "1990-12-12",
		Email:        "client@mail.com",
		Phone:        "12945678901",
		EmailCreator: creator.Email,
	}

	expectedName := "Name Receita Federal"
	expectedBirthDay := "12/12/1990"

	client := &models.Client{
		Validated: true,
		Name:      expectedName,
		CPF:       "05899795077",
		Birthdate: input.BirthDate,
		Email:     input.Email,
		Phone:     input.Phone,
		Creators:  []*models.Creator{creator},
	}

	suite.mockRFService.(*mocks.MockRFService).
		On("ConsultaCPF", "05899795077", "12/12/1990").
		Return(&gov.ReceitaFederalResponse{
			Status: true,
			Result: gov.ConsultaData{
				NomeDaPF:       expectedName,
				NumeroDeCPF:    input.CPF,
				DataNascimento: expectedBirthDay,
			},
		}, nil)

	suite.mockCreatorRepository.(*mocks.MockCreatorRepository).
		On("FindCreatorByUserEmail", creator.Email).
		Return(creator, nil)

	suite.mockClientRepository.(*mocks.MockClientRepository).
		On("FindByEmail", input.Email).
		Return(client, nil)

	_, err := suite.sut.CreateClient(input)

	suite.Error(err)
	suite.mockCreatorRepository.(*mocks.MockCreatorRepository).AssertCalled(suite.T(), "FindCreatorByUserEmail", creator.Email)
	suite.mockClientRepository.(*mocks.MockClientRepository).AssertCalled(suite.T(), "FindByEmail", client.Email)
}

func TestClientServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ClientServiceTestSuite))
}
