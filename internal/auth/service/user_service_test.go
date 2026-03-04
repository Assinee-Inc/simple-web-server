package service_test

import (
	"testing"

	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var _ authrepo.UserRepository = (*mocks.MockUserRepository)(nil)

type UserServiceTestSuite struct {
	suite.Suite
	sut                authsvc.UserService
	mockUserRepository authrepo.UserRepository
	mockEncrypter      utils.Encrypter
	testInput          authsvc.InputCreateUser
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.setUpInput()
	suite.setupMocks()
}

func (suite *UserServiceTestSuite) setUpInput() {
	suite.testInput = authsvc.InputCreateUser{
		Username:             "Valid UserName",
		Email:                "valid@mail.com",
		Password:             "Password123!",
		PasswordConfirmation: "Password123!",
	}
}

func (suite *UserServiceTestSuite) setupMocks() {
	suite.mockUserRepository = new(mocks.MockUserRepository)
	suite.mockEncrypter = new(mocks.MockEncrypter)
	suite.sut = authsvc.NewUserService(suite.mockUserRepository, suite.mockEncrypter)
}

func (suite *UserServiceTestSuite) setupSuccessfulMockExpectations() {
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByUserEmail", suite.testInput.Email).Return(nil)
	suite.mockUserRepository.(*mocks.MockUserRepository).On("Create", mock.AnythingOfType("*model.User")).Return(nil)
	suite.mockEncrypter.(*mocks.MockEncrypter).On("HashPassword", suite.testInput.Password).Return("HashedPassword123!")
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (suite *UserServiceTestSuite) TestCreateUser_Success() {
	// Arrange
	suite.setupSuccessfulMockExpectations()

	// Act
	userID, err := suite.sut.CreateUser(suite.testInput)

	// Assert
	suite.NoError(err)
	suite.Equal(uint(0), userID) // ID is 0 because mock Create doesn't set it
	suite.mockUserRepository.(*mocks.MockUserRepository).AssertCalled(suite.T(), "Create", mock.AnythingOfType("*model.User"))
}

func (suite *UserServiceTestSuite) TestCreateUser_ShouldCallHashPassword() {
	// Arrange
	suite.setupSuccessfulMockExpectations()

	// Act
	userID, err := suite.sut.CreateUser(suite.testInput)

	// Assert
	suite.NoError(err)
	suite.Equal(uint(0), userID) // ID is 0 because mock Create doesn't set it
	suite.mockEncrypter.(*mocks.MockEncrypter).AssertCalled(suite.T(), "HashPassword", suite.testInput.Password)
	suite.mockUserRepository.(*mocks.MockUserRepository).AssertCalled(suite.T(), "Create", mock.AnythingOfType("*model.User"))
}

func (suite *UserServiceTestSuite) TestShouldReturnErrorIfPasswordAndConfirmationAreDifferent() {
	// Arrange
	suite.testInput.PasswordConfirmation = "DifferentPassword"

	// Act
	userID, err := suite.sut.CreateUser(suite.testInput)

	// Assert
	suite.Error(err)
	suite.Equal(uint(0), userID)
}

func (suite *UserServiceTestSuite) TestShouldReturnErrorIfUserAlreadyExists() {
	// Arrange
	existingUser := authmodel.NewUser("Existing User", "HashedPassword123!", "valid@mail.com")
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByUserEmail", suite.testInput.Email).Return(existingUser)

	// Act
	userID, err := suite.sut.CreateUser(suite.testInput)

	// Assert
	suite.Error(err)
	suite.Equal(uint(0), userID)
	suite.mockUserRepository.(*mocks.MockUserRepository).AssertCalled(suite.T(), "FindByUserEmail", suite.testInput.Email)
	suite.mockUserRepository.(*mocks.MockUserRepository).AssertNotCalled(suite.T(), "Create")
}
