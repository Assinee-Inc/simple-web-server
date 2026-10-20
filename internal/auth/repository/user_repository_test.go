package repository_test

import (
	"testing"
	"time"

	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	"github.com/Assinee-Inc/docffy/internal/subscription/model"
	"github.com/Assinee-Inc/docffy/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db             *gorm.DB
	userRepository authrepo.UserRepository
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(suite.T(), err)
	err = db.AutoMigrate(&authmodel.User{}, &model.Subscription{})
	require.NoError(suite.T(), err)

	suite.db = db
	database.DB = db
	suite.userRepository = authrepo.NewGormUserRepository(db)
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.db.Exec("DELETE FROM subscriptions")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserRepositoryTestSuite) createTestUserWithSubscription() *authmodel.User {
	user := authmodel.NewGoogleUser("testuser", "test@example.com", "google-sub-1")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	subscription := model.NewSubscription(user.ID, "test_plan")
	err = suite.db.Create(subscription).Error
	suite.Require().NoError(err)

	return user
}

func (suite *UserRepositoryTestSuite) TestFindByGoogleID_SuccessAndNotFound() {
	user := authmodel.NewGoogleUser("Google User", "googleuser@example.com", "google-id-abc-123")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	found, err := suite.userRepository.FindByGoogleID("google-id-abc-123")
	suite.Require().NoError(err)
	suite.Require().NotNil(found)
	assert.Equal(suite.T(), user.ID, found.ID)
	assert.Equal(suite.T(), "googleuser@example.com", found.Email)
	assert.Equal(suite.T(), "google-id-abc-123", found.GoogleID)

	notFound, err := suite.userRepository.FindByGoogleID("non-existent-google-id")
	suite.Require().NoError(err)
	assert.Nil(suite.T(), notFound)
}

func (suite *UserRepositoryTestSuite) TestFindByID_And_Update_And_Delete() {
	user := authmodel.NewGoogleUser("Initial Name", "initial@example.com", "")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	found := suite.userRepository.FindByID(user.ID)
	suite.Require().NotNil(found)
	assert.Equal(suite.T(), "initial@example.com", found.Email)

	found.GoogleID = "linked-google-sub-999"
	err = suite.userRepository.Update(found)
	suite.Require().NoError(err)

	byGoogle, err := suite.userRepository.FindByGoogleID("linked-google-sub-999")
	suite.Require().NoError(err)
	suite.Require().NotNil(byGoogle)
	assert.Equal(suite.T(), user.ID, byGoogle.ID)

	err = suite.userRepository.Delete(user.ID)
	suite.Require().NoError(err)
	assert.Nil(suite.T(), suite.userRepository.FindByID(user.ID))
}

func (suite *UserRepositoryTestSuite) TestFindByUserEmail_LoadsSubscription() {
	user := suite.createTestUserWithSubscription()

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)

	assert.NotNil(suite.T(), foundUser.Subscription)
	assert.Equal(suite.T(), user.ID, foundUser.Subscription.UserID)
	assert.Equal(suite.T(), "test_plan", foundUser.Subscription.PlanID)
	assert.True(suite.T(), foundUser.Subscription.IsTrialActive)
	assert.Equal(suite.T(), "inactive", foundUser.Subscription.SubscriptionStatus)
}

func (suite *UserRepositoryTestSuite) TestFindByUserEmail_UserWithoutSubscription() {
	user := authmodel.NewGoogleUser("testuser2", "test2@example.com", "google-sub-2")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	assert.Nil(suite.T(), foundUser.Subscription)
}

func (suite *UserRepositoryTestSuite) TestFindByUserEmail_UserNotFound() {
	foundUser := suite.userRepository.FindByUserEmail("nonexistent@example.com")
	assert.Nil(suite.T(), foundUser)
}

func (suite *UserRepositoryTestSuite) TestUserTrialMethods_WithSubscription() {
	user := suite.createTestUserWithSubscription()

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	assert.True(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Greater(suite.T(), foundUser.DaysLeftInTrial(), 0)
}

func (suite *UserRepositoryTestSuite) TestUserTrialMethods_WithoutSubscription() {
	user := authmodel.NewGoogleUser("testuser3", "test3@example.com", "google-sub-3")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	assert.Nil(suite.T(), foundUser.Subscription)

	assert.False(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Equal(suite.T(), 0, foundUser.DaysLeftInTrial())
}

func (suite *UserRepositoryTestSuite) TestUserTrialMethods_ExpiredTrial() {
	user := suite.createTestUserWithSubscription()

	subscription := &model.Subscription{}
	err := suite.db.Where("user_id = ?", user.ID).First(subscription).Error
	suite.Require().NoError(err)

	subscription.IsTrialActive = false
	subscription.TrialEndDate = time.Now().AddDate(0, 0, -1)
	err = suite.db.Save(subscription).Error
	suite.Require().NoError(err)

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	assert.False(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Equal(suite.T(), 0, foundUser.DaysLeftInTrial())
}

func (suite *UserRepositoryTestSuite) TestUserTrialMethods_ActiveSubscription() {
	user := suite.createTestUserWithSubscription()

	subscription := &model.Subscription{}
	err := suite.db.Where("user_id = ?", user.ID).First(subscription).Error
	suite.Require().NoError(err)

	subscription.SubscriptionStatus = "active"
	subscription.StripeCustomerID = "cus_test123"
	subscription.StripeSubscriptionID = "sub_test123"
	err = suite.db.Save(subscription).Error
	suite.Require().NoError(err)

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	assert.True(suite.T(), foundUser.IsInTrialPeriod())
	assert.True(suite.T(), foundUser.IsSubscribed())
	assert.Greater(suite.T(), foundUser.DaysLeftInTrial(), 0)
}
