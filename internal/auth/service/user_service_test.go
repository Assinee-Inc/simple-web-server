package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var _ authrepo.UserRepository = (*mocks.MockUserRepository)(nil)

type mockGoogleVerifier struct {
	mock.Mock
}

func (m *mockGoogleVerifier) VerifyIDToken(ctx context.Context, idToken string) (*authmodel.GoogleIdentity, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authmodel.GoogleIdentity), args.Error(1)
}

type UserServiceTestSuite struct {
	suite.Suite
	sut                authsvc.UserService
	mockUserRepository *mocks.MockUserRepository
	mockVerifier       *mockGoogleVerifier
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.mockUserRepository = new(mocks.MockUserRepository)
	suite.mockVerifier = new(mockGoogleVerifier)
	suite.sut = authsvc.NewUserService(suite.mockUserRepository, suite.mockVerifier)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (suite *UserServiceTestSuite) TestAuthenticateWithGoogle_NewUserCreated() {
	ctx := context.Background()
	identity := &authmodel.GoogleIdentity{
		Sub:           "google-sub-new-123",
		Email:         "novo@example.com",
		EmailVerified: true,
		Name:          "Novo Criador",
	}

	suite.mockVerifier.On("VerifyIDToken", ctx, "valid-google-token").Return(identity, nil)
	suite.mockUserRepository.On("FindByGoogleID", "google-sub-new-123").Return(nil, nil)
	suite.mockUserRepository.On("FindByEmail", "novo@example.com").Return(nil)
	suite.mockUserRepository.On("Create", mock.MatchedBy(func(u *authmodel.User) bool {
		u.ID = 10
		return u.Email == "novo@example.com" &&
			u.GoogleID == "google-sub-new-123" &&
			u.Username == "Novo Criador" &&
			u.EmailVerifiedAt != nil &&
			u.TermsAcceptedAt == nil
	})).Return(nil)

	user, isNew, err := suite.sut.AuthenticateWithGoogle(ctx, "valid-google-token")

	suite.NoError(err)
	suite.True(isNew)
	suite.Require().NotNil(user)
	suite.Equal(uint(10), user.ID)
	suite.Equal("novo@example.com", user.Email)
	suite.Equal("google-sub-new-123", user.GoogleID)
	suite.NotNil(user.EmailVerifiedAt)
	suite.mockUserRepository.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestAuthenticateWithGoogle_ExistingUserByGoogleID() {
	ctx := context.Background()
	now := time.Now()
	existingUser := &authmodel.User{
		PublicID:        "usr_existing",
		Username:        "Maria",
		Email:           "maria@example.com",
		GoogleID:        "google-sub-existing",
		EmailVerifiedAt: &now,
	}
	existingUser.ID = 5

	identity := &authmodel.GoogleIdentity{
		Sub:           "google-sub-existing",
		Email:         "maria@example.com",
		EmailVerified: true,
		Name:          "Maria Silva",
	}

	suite.mockVerifier.On("VerifyIDToken", ctx, "valid-google-token").Return(identity, nil)
	suite.mockUserRepository.On("FindByGoogleID", "google-sub-existing").Return(existingUser, nil)

	user, isNew, err := suite.sut.AuthenticateWithGoogle(ctx, "valid-google-token")

	suite.NoError(err)
	suite.False(isNew)
	suite.Require().NotNil(user)
	suite.Equal(uint(5), user.ID)
	suite.Equal("google-sub-existing", user.GoogleID)
	suite.mockUserRepository.AssertNotCalled(suite.T(), "Create")
}

func (suite *UserServiceTestSuite) TestAuthenticateWithGoogle_LinksExistingUserByEmailWithoutPassword() {
	ctx := context.Background()
	legacyUser := &authmodel.User{
		PublicID:        "usr_legacy",
		Username:        "Criador Legado",
		Email:           "legado@example.com",
		GoogleID:        "",
		EmailVerifiedAt: nil,
	}
	legacyUser.ID = 99

	identity := &authmodel.GoogleIdentity{
		Sub:           "google-sub-linked-99",
		Email:         "legado@example.com",
		EmailVerified: true,
		Name:          "Criador Legado",
	}

	suite.mockVerifier.On("VerifyIDToken", ctx, "token-for-legacy").Return(identity, nil)
	suite.mockUserRepository.On("FindByGoogleID", "google-sub-linked-99").Return(nil, nil)
	suite.mockUserRepository.On("FindByEmail", "legado@example.com").Return(legacyUser)
	suite.mockUserRepository.On("Update", mock.MatchedBy(func(u *authmodel.User) bool {
		return u.ID == 99 && u.GoogleID == "google-sub-linked-99" && u.EmailVerifiedAt != nil
	})).Return(nil)

	user, isNew, err := suite.sut.AuthenticateWithGoogle(ctx, "token-for-legacy")

	suite.NoError(err)
	suite.False(isNew)
	suite.Require().NotNil(user)
	suite.Equal(uint(99), user.ID)
	suite.Equal("google-sub-linked-99", user.GoogleID)
	suite.NotNil(user.EmailVerifiedAt)
	suite.mockUserRepository.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestAuthenticateWithGoogle_RejectsUnverifiedEmailOrInvalidToken() {
	ctx := context.Background()

	t := suite.T()
	t.Run("erro do verificador Google", func(t *testing.T) {
		suite.SetupTest()
		suite.mockVerifier.On("VerifyIDToken", ctx, "bad-token").Return(nil, errors.New("invalid token"))

		user, isNew, err := suite.sut.AuthenticateWithGoogle(ctx, "bad-token")
		suite.Error(err)
		suite.False(isNew)
		suite.Nil(user)
	})

	t.Run("email_verified == false", func(t *testing.T) {
		suite.SetupTest()
		suite.mockVerifier.On("VerifyIDToken", ctx, "unverified-token").Return(&authmodel.GoogleIdentity{
			Sub:           "google-sub-unverified",
			Email:         "unverified@example.com",
			EmailVerified: false,
			Name:          "Unverified",
		}, nil)

		user, isNew, err := suite.sut.AuthenticateWithGoogle(ctx, "unverified-token")
		suite.ErrorIs(err, authsvc.ErrInvalidGoogleToken)
		suite.False(isNew)
		suite.Nil(user)
	})
}
