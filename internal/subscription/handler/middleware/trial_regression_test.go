package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	"github.com/anglesson/simple-web-server/internal/subscription/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/subscription/model"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TrialRegressionTestSuite struct {
	suite.Suite
	userRepository authrepo.UserRepository
}

func TestTrialRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(TrialRegressionTestSuite))
}

func (suite *TrialRegressionTestSuite) SetupSuite() {
	database.Connect()
	suite.userRepository = authrepo.NewGormUserRepository(database.DB)
}

func (suite *TrialRegressionTestSuite) SetupTest() {
	// Limpar dados de teste
	database.DB.Exec("DELETE FROM subscriptions")
	database.DB.Exec("DELETE FROM users")
}

// TestTrialMiddleware_Regression_UserWithActiveTrial_ShouldAllowAccess
func (suite *TrialRegressionTestSuite) TestTrialMiddleware_Regression_UserWithActiveTrial_ShouldAllowAccess() {
	user := authmodel.NewUser("testuser", "password123", "test@example.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	subscription := model.NewSubscription(user.ID, "default_plan")
	subscription.IsTrialActive = true
	subscription.TrialStartDate = time.Now()
	subscription.TrialEndDate = time.Now().AddDate(0, 0, 7)
	subscription.SubscriptionStatus = "inactive"
	subscription.Origin = "web"

	err = database.DB.Create(subscription).Error
	suite.Require().NoError(err)

	var dbSubscription model.Subscription
	err = database.DB.Where("user_id = ?", user.ID).First(&dbSubscription).Error
	suite.Require().NoError(err)

	assert.True(suite.T(), dbSubscription.IsTrialActive)
	assert.Equal(suite.T(), "inactive", dbSubscription.SubscriptionStatus)
	assert.True(suite.T(), time.Now().Before(dbSubscription.TrialEndDate))

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	assert.True(suite.T(), foundUser.IsInTrialPeriod(), "Usuário deve estar no período trial")
	assert.False(suite.T(), foundUser.IsSubscribed(), "Usuário não deve estar inscrito")
	assert.Greater(suite.T(), foundUser.DaysLeftInTrial(), 0, "Deve ter dias restantes no trial")

	req := httptest.NewRequest("GET", "/dashboard", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code, "Usuário com trial ativo deve ter acesso permitido")
}

// TestTrialMiddleware_Regression_UserWithoutSubscription_ShouldRedirectToSettings
func (suite *TrialRegressionTestSuite) TestTrialMiddleware_Regression_UserWithoutSubscription_ShouldRedirectToSettings() {
	user := authmodel.NewUser("testuser2", "password123", "test2@example.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	assert.Nil(suite.T(), foundUser.Subscription, "Usuário não deve ter subscription")

	assert.False(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Equal(suite.T(), 0, foundUser.DaysLeftInTrial())

	req := httptest.NewRequest("GET", "/dashboard", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusSeeOther, w.Code)
	assert.Equal(suite.T(), "/settings", w.Header().Get("Location"))
}

// TestTrialMiddleware_Regression_UserWithExpiredTrial_ShouldRedirectToSettings
func (suite *TrialRegressionTestSuite) TestTrialMiddleware_Regression_UserWithExpiredTrial_ShouldRedirectToSettings() {
	user := authmodel.NewUser("testuser3", "password123", "test3@example.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	subscription := model.NewSubscription(user.ID, "default_plan")
	subscription.IsTrialActive = false
	subscription.TrialEndDate = time.Now().AddDate(0, 0, -1)
	subscription.SubscriptionStatus = "inactive"

	err = database.DB.Create(subscription).Error
	suite.Require().NoError(err)

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	assert.False(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Equal(suite.T(), 0, foundUser.DaysLeftInTrial())

	req := httptest.NewRequest("GET", "/dashboard", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusSeeOther, w.Code)
	assert.Equal(suite.T(), "/settings", w.Header().Get("Location"))
}

// TestUserRepository_Regression_PreloadSubscription_WorksCorrectly
func (suite *TrialRegressionTestSuite) TestUserRepository_Regression_PreloadSubscription_WorksCorrectly() {
	user := authmodel.NewUser("testuser4", "password123", "test4@example.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	subscription := model.NewSubscription(user.ID, "test_plan")
	err = database.DB.Create(subscription).Error
	suite.Require().NoError(err)

	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)

	assert.NotNil(suite.T(), foundUser.Subscription, "Subscription deve ser carregada via Preload")
	assert.Equal(suite.T(), user.ID, foundUser.Subscription.UserID)
	assert.Equal(suite.T(), "test_plan", foundUser.Subscription.PlanID)
	assert.True(suite.T(), foundUser.Subscription.IsTrialActive)

	assert.True(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Greater(suite.T(), foundUser.DaysLeftInTrial(), 0)
}
