package repository_test

import (
	"testing"
	"time"

	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	"github.com/anglesson/simple-web-server/internal/subscription/model"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	userRepository authrepo.UserRepository
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	database.Connect()
	suite.userRepository = authrepo.NewGormUserRepository(database.DB)
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Limpar dados de teste
	database.DB.Exec("DELETE FROM subscriptions")
	database.DB.Exec("DELETE FROM users")
}

func (suite *UserRepositoryTestSuite) createTestUserWithSubscription() *authmodel.User {
	// Criar usuário
	user := authmodel.NewUser("testuser", "password123", "test@example.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	// Criar subscription
	subscription := model.NewSubscription(user.ID, "test_plan")
	err = database.DB.Create(subscription).Error
	suite.Require().NoError(err)

	return user
}

func (suite *UserRepositoryTestSuite) TestFindByUserEmail_LoadsSubscription() {
	// Criar usuário com subscription
	user := suite.createTestUserWithSubscription()

	// Buscar usuário por email
	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)

	// Verificar que a subscription foi carregada
	assert.NotNil(suite.T(), foundUser.Subscription)
	assert.Equal(suite.T(), user.ID, foundUser.Subscription.UserID)
	assert.Equal(suite.T(), "test_plan", foundUser.Subscription.PlanID)
	assert.True(suite.T(), foundUser.Subscription.IsTrialActive)
	assert.Equal(suite.T(), "inactive", foundUser.Subscription.SubscriptionStatus)
}

func (suite *UserRepositoryTestSuite) TestFindByUserEmail_UserWithoutSubscription() {
	// Criar usuário sem subscription
	user := authmodel.NewUser("testuser2", "password123", "test2@example.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	// Buscar usuário por email
	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)

	// Verificar que a subscription é nil
	assert.Nil(suite.T(), foundUser.Subscription)
}

func (suite *UserRepositoryTestSuite) TestFindByUserEmail_UserNotFound() {
	// Buscar usuário que não existe
	foundUser := suite.userRepository.FindByUserEmail("nonexistent@example.com")
	assert.Nil(suite.T(), foundUser)
}

func (suite *UserRepositoryTestSuite) TestFindBySessionToken_LoadsSubscription() {
	// Criar usuário com subscription
	user := suite.createTestUserWithSubscription()

	// Definir session token
	user.SessionToken = "test-session-token"
	err := suite.userRepository.Save(user)
	suite.Require().NoError(err)

	// Buscar usuário por session token
	foundUser := suite.userRepository.FindBySessionToken("test-session-token")
	suite.Require().NotNil(foundUser)

	// Verificar que a subscription foi carregada
	assert.NotNil(suite.T(), foundUser.Subscription)
	assert.Equal(suite.T(), user.ID, foundUser.Subscription.UserID)
	assert.Equal(suite.T(), "test_plan", foundUser.Subscription.PlanID)
}

func (suite *UserRepositoryTestSuite) TestFindBySessionToken_UserNotFound() {
	// Buscar usuário com session token que não existe
	foundUser := suite.userRepository.FindBySessionToken("nonexistent-token")
	assert.Nil(suite.T(), foundUser)
}

func (suite *UserRepositoryTestSuite) TestUserTrialMethods_WithSubscription() {
	// Criar usuário com subscription ativa
	user := suite.createTestUserWithSubscription()

	// Buscar usuário para garantir que a subscription foi carregada
	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	// Testar métodos do trial
	assert.True(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Greater(suite.T(), foundUser.DaysLeftInTrial(), 0)
}

func (suite *UserRepositoryTestSuite) TestUserTrialMethods_WithoutSubscription() {
	// Criar usuário sem subscription
	user := authmodel.NewUser("testuser3", "password123", "test3@example.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	// Buscar usuário
	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	assert.Nil(suite.T(), foundUser.Subscription)

	// Testar métodos do trial (devem retornar false/0)
	assert.False(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Equal(suite.T(), 0, foundUser.DaysLeftInTrial())
}

func (suite *UserRepositoryTestSuite) TestUserTrialMethods_ExpiredTrial() {
	// Criar usuário com subscription
	user := suite.createTestUserWithSubscription()

	// Modificar a subscription para trial expirado
	subscription := &model.Subscription{}
	err := database.DB.Where("user_id = ?", user.ID).First(subscription).Error
	suite.Require().NoError(err)

	subscription.IsTrialActive = false
	subscription.TrialEndDate = time.Now().AddDate(0, 0, -1) // Trial expirado
	err = database.DB.Save(subscription).Error
	suite.Require().NoError(err)

	// Buscar usuário
	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	// Testar métodos do trial
	assert.False(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Equal(suite.T(), 0, foundUser.DaysLeftInTrial())
}

func (suite *UserRepositoryTestSuite) TestUserTrialMethods_ActiveSubscription() {
	// Criar usuário com subscription
	user := suite.createTestUserWithSubscription()

	// Modificar a subscription para ativa
	subscription := &model.Subscription{}
	err := database.DB.Where("user_id = ?", user.ID).First(subscription).Error
	suite.Require().NoError(err)

	subscription.SubscriptionStatus = "active"
	subscription.StripeCustomerID = "cus_test123"
	subscription.StripeSubscriptionID = "sub_test123"
	err = database.DB.Save(subscription).Error
	suite.Require().NoError(err)

	// Buscar usuário
	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	// Testar métodos do trial
	assert.True(suite.T(), foundUser.IsInTrialPeriod()) // Ainda no trial
	assert.True(suite.T(), foundUser.IsSubscribed())    // Mas também inscrito
	assert.Greater(suite.T(), foundUser.DaysLeftInTrial(), 0)
}
