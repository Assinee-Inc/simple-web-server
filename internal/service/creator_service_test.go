package service_test

import (
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/gov"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

const (
	validName        = "Valid Name"
	validEmail       = "valid@mail.com"
	validCPF         = "058.997.950-77"
	validBirthDate   = "1990-12-12"
	validPhoneNumber = "(12) 94567-8901"
	validPassword    = "ValidPassword123!"
)

type CreatorServiceTestSuite struct {
	suite.Suite
	sut                     service.CreatorService
	mockCreatorRepo         repository.CreatorRepository
	mockRFService           gov.ReceitaFederalService
	mockUserService         service.UserService
	mockSubscriptionService service.SubscriptionService
	mockPaymentGateway      service.PaymentGateway
	testInput               models.InputCreateCreator
	testInputUser           models.InputCreateUser
}

func TestCreatorServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CreatorServiceTestSuite))
}

func (suite *CreatorServiceTestSuite) SetupTest() {
	suite.setupTestInput()
	suite.setupMocks()
}

func (suite *CreatorServiceTestSuite) setupTestInput() {
	suite.testInput = models.InputCreateCreator{
		Name:                 validName,
		BirthDate:            validBirthDate,
		PhoneNumber:          validPhoneNumber,
		Email:                validEmail,
		CPF:                  validCPF,
		Password:             validPassword,
		PasswordConfirmation: validPassword,
		TermsAccepted:        "on",
	}

	suite.testInputUser = models.InputCreateUser{
		Username:             validName,
		Email:                validEmail,
		Password:             validPassword,
		PasswordConfirmation: validPassword,
	}
}

func (suite *CreatorServiceTestSuite) setupMocks() {
	suite.mockCreatorRepo = new(mocks.MockCreatorRepository)
	suite.mockRFService = new(mocks.MockRFService)
	suite.mockUserService = new(mocks.MockUserService)
	suite.mockSubscriptionService = new(mocks.MockSubscriptionService)
	suite.mockPaymentGateway = new(mocks.MockPaymentGateway)
	suite.sut = service.NewCreatorService(suite.mockCreatorRepo, suite.mockRFService, suite.mockUserService, suite.mockSubscriptionService, suite.mockPaymentGateway)
}

func (suite *CreatorServiceTestSuite) setupSuccessfulMockExpectations(validatedName string) {
	cleanCPF := "05899795077"   // CPF without formatting
	cleanPhone := "12945678901" // Phone without formatting
	birthDate, _ := time.Parse("2006-01-02", validBirthDate)

	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).
		On("FindByCPF", cleanCPF).
		Return(nil, nil)

	suite.mockRFService.(*mocks.MockRFService).
		On("ConsultaCPF", cleanCPF, birthDate.Format("02/01/2006")).
		Return(&gov.ReceitaFederalResponse{
			Status: true,
			Result: gov.ConsultaData{
				NomeDaPF:       validatedName,
				NumeroDeCPF:    validCPF,
				DataNascimento: "12/12/1990",
			},
		}, nil)

	expectedUser := &models.User{Model: gorm.Model{ID: 1}}
	matcher := mock.MatchedBy(func(input models.InputCreateUser) bool {
		return input.Email == validEmail && input.Password == validPassword && input.PasswordConfirmation == validPassword
	})
	suite.mockUserService.(*mocks.MockUserService).
		On("CreateUser", matcher).
		Return(expectedUser, nil)

	expectedCreator := &models.Creator{
		Name:      validatedName,
		Email:     validEmail,
		Phone:     cleanPhone,
		CPF:       cleanCPF,
		BirthDate: birthDate,
		UserID:    expectedUser.ID,
	}
	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).On("Create", expectedCreator).Return(nil)

	// Mock PaymentGateway expectations
	suite.mockPaymentGateway.(*mocks.MockPaymentGateway).
		On("CreateCustomer", validEmail, validatedName).
		Return("cus_123", nil)

	// Mock SubscriptionService expectations
	suite.mockSubscriptionService.(*mocks.MockSubscriptionService).
		On("CreateSubscription", expectedUser.ID, "default_plan").
		Return(&models.Subscription{UserID: expectedUser.ID}, nil)

	suite.mockSubscriptionService.(*mocks.MockSubscriptionService).
		On("ActivateSubscription", mock.AnythingOfType("*models.Subscription"), "cus_123", "").
		Return(nil)
}

func (suite *CreatorServiceTestSuite) TestCreateCreator_Success() {
	suite.setupSuccessfulMockExpectations(validName)

	creator, err := suite.sut.CreateCreator(suite.testInput)

	suite.NoError(err)
	suite.NotNil(creator)
	suite.mockUserService.(*mocks.MockUserService).AssertExpectations(suite.T())
	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).AssertExpectations(suite.T())
	suite.mockRFService.(*mocks.MockRFService).AssertExpectations(suite.T())
}

func (suite *CreatorServiceTestSuite) TestCreateCreator_ShouldCallUserService() {
	// Arrange
	suite.setupSuccessfulMockExpectations(validName)

	// Act
	_, err := suite.sut.CreateCreator(suite.testInput)

	// Assert
	suite.NoError(err)
	suite.mockUserService.(*mocks.MockUserService).
		AssertCalled(suite.T(), "CreateUser", mock.MatchedBy(func(input models.InputCreateUser) bool {
			return input.Email == validEmail && input.Password == validPassword && input.PasswordConfirmation == validPassword
		}))
}

func (suite *CreatorServiceTestSuite) TestCreateCreator_ShouldUpdateCreatorWithReceitaFederalData() {
	config.AppConfig.AppMode = "PRODUCTION"
	// Arrange
	validatedName := "Name Receita Federal"
	suite.setupSuccessfulMockExpectations(validatedName)

	// Act
	_, err := suite.sut.CreateCreator(suite.testInput)

	// Assert
	suite.NoError(err)
	suite.mockRFService.(*mocks.MockRFService).AssertCalled(
		suite.T(),
		"ConsultaCPF",
		"05899795077",
		"12/12/1990",
	)
}

func (suite *CreatorServiceTestSuite) TestCreateCreator_ShouldThrowErrorIfCreatorHasARegister() {
	// Arrange
	cleanCPF := "05899795077"
	existingCreator := &models.Creator{CPF: cleanCPF}

	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).
		On("FindByCPF", cleanCPF).
		Return(existingCreator, nil)

	// Act
	creator, err := suite.sut.CreateCreator(suite.testInput)

	// Assert
	suite.Error(err)
	suite.Assert().Nil(creator)
	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).AssertCalled(suite.T(), "FindByCPF", cleanCPF)
	suite.mockRFService.(*mocks.MockRFService).AssertNotCalled(suite.T(), "ConsultaCPF")
	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).AssertNotCalled(suite.T(), "Create")
}

func (suite *CreatorServiceTestSuite) TestCreateCreator_FailsIfDataNotExistsInReceitaFederal() {
	config.AppConfig.AppMode = "PRODUCTION"
	// Arrange
	cleanCPF := "05899795077"
	birthDate, _ := time.Parse("2006-01-02", validBirthDate)

	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).
		On("FindByCPF", cleanCPF).
		Return(nil, nil)

	suite.mockRFService.(*mocks.MockRFService).On("ConsultaCPF",
		cleanCPF,
		birthDate.Format("02/01/2006")).
		Return(&gov.ReceitaFederalResponse{
			Status: false,
			Result: gov.ConsultaData{},
		}, nil)

	// Act
	creator, err := suite.sut.CreateCreator(suite.testInput)

	// Assert
	suite.Error(err)
	suite.Assert().Nil(creator)
	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).AssertNotCalled(suite.T(), "Create")
	suite.mockRFService.(*mocks.MockRFService).AssertCalled(
		suite.T(),
		"ConsultaCPF",
		cleanCPF,
		birthDate.Format("02/01/2006"),
	)
}

func (suite *CreatorServiceTestSuite) TestShouldThrowErrorIfAnyDataIsInvalid() {
	// Arrange
	suite.testInput.Email = "invalid_mail" // invalid mail

	// Act
	creator, err := suite.sut.CreateCreator(suite.testInput)

	// Assert
	suite.Error(err)
	suite.Assert().Nil(creator)
	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).AssertNotCalled(suite.T(), "FindByCPF")
	suite.mockCreatorRepo.(*mocks.MockCreatorRepository).AssertNotCalled(suite.T(), "Create")
	suite.mockRFService.(*mocks.MockRFService).AssertNotCalled(suite.T(), "ConsultaCPF")
}
