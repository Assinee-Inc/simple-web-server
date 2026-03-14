package service_test

import (
	"testing"
	"time"

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

func (suite *UserServiceTestSuite) TestGenerateVerificationToken_Success() {
	// Arrange
	user := authmodel.NewUser("Test User", "hashed", "valid@mail.com")
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByUserEmail", "valid@mail.com").Return(user)
	suite.mockUserRepository.(*mocks.MockUserRepository).On("Save", mock.AnythingOfType("*model.User")).Return(nil)

	// Act
	token, err := suite.sut.GenerateVerificationToken("valid@mail.com")

	// Assert
	suite.NoError(err)
	suite.NotEmpty(token)
	suite.mockUserRepository.(*mocks.MockUserRepository).AssertCalled(suite.T(), "Save", mock.AnythingOfType("*model.User"))
}

func (suite *UserServiceTestSuite) TestGenerateVerificationToken_UserNotFound() {
	// Arrange
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByUserEmail", "notfound@mail.com").Return(nil)

	// Act
	token, err := suite.sut.GenerateVerificationToken("notfound@mail.com")

	// Assert
	suite.Error(err)
	suite.Equal(authsvc.ErrUserNotFound, err)
	suite.Empty(token)
}

func (suite *UserServiceTestSuite) TestConfirmEmail_Success() {
	// Arrange
	user := authmodel.NewUser("Test User", "hashed", "valid@mail.com")
	user.EmailVerificationToken = "validtoken123"
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByEmailVerificationToken", "validtoken123").Return(user)
	suite.mockUserRepository.(*mocks.MockUserRepository).On("Save", mock.AnythingOfType("*model.User")).Return(nil)

	// Act
	confirmedUser, err := suite.sut.ConfirmEmail("validtoken123")

	// Assert
	suite.NoError(err)
	suite.NotNil(confirmedUser)
	suite.NotNil(confirmedUser.EmailVerifiedAt)
	suite.Empty(confirmedUser.EmailVerificationToken)
}

func (suite *UserServiceTestSuite) TestConfirmEmail_InvalidToken() {
	// Arrange
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByEmailVerificationToken", "badtoken").Return(nil)

	// Act
	confirmedUser, err := suite.sut.ConfirmEmail("badtoken")

	// Assert
	suite.Error(err)
	suite.Equal(authsvc.ErrInvalidVerificationToken, err)
	suite.Nil(confirmedUser)
}

func (suite *UserServiceTestSuite) TestResendConfirmation_Success() {
	// Arrange
	user := authmodel.NewUser("Test User", "hashed", "valid@mail.com")
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByUserEmail", "valid@mail.com").Return(user)
	suite.mockUserRepository.(*mocks.MockUserRepository).On("Save", mock.AnythingOfType("*model.User")).Return(nil)

	// Act
	updatedUser, err := suite.sut.ResendConfirmation("valid@mail.com")

	// Assert
	suite.NoError(err)
	suite.NotNil(updatedUser)
	suite.NotEmpty(updatedUser.EmailVerificationToken)
}

func (suite *UserServiceTestSuite) TestResendConfirmation_AlreadyVerified() {
	// Arrange
	verifiedUser := authmodel.NewUser("Verified", "hashed", "verified@mail.com")
	now := time.Now()
	verifiedUser.EmailVerifiedAt = &now
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByUserEmail", "verified@mail.com").Return(verifiedUser)

	// Act
	result, err := suite.sut.ResendConfirmation("verified@mail.com")

	// Assert
	suite.Error(err)
	suite.Equal(authsvc.ErrEmailAlreadyVerified, err)
	suite.Nil(result)
}

func (suite *UserServiceTestSuite) TestResendConfirmation_UserNotFound() {
	// Arrange
	suite.mockUserRepository.(*mocks.MockUserRepository).On("FindByUserEmail", "ghost@mail.com").Return(nil)

	// Act
	result, err := suite.sut.ResendConfirmation("ghost@mail.com")

	// Assert
	suite.NoError(err)
	suite.Nil(result)
}
