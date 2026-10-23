package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	"github.com/Assinee-Inc/docffy/internal/subscription/handler/middleware"
	"github.com/Assinee-Inc/docffy/internal/subscription/model"
	"github.com/Assinee-Inc/docffy/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TrialRegressionTestSuite struct {
	suite.Suite
	db             *gorm.DB
	userRepository authrepo.UserRepository
}

func TestTrialRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(TrialRegressionTestSuite))
}

func (suite *TrialRegressionTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(suite.T(), err)
	err = db.AutoMigrate(&authmodel.User{}, &model.Subscription{})
	require.NoError(suite.T(), err)

	suite.db = db
	database.DB = db
	suite.userRepository = authrepo.NewGormUserRepository(db)
}

func (suite *TrialRegressionTestSuite) SetupTest() {
	suite.db.Exec("DELETE FROM subscriptions")
	suite.db.Exec("DELETE FROM users")
}

func (suite *TrialRegressionTestSuite) TestTrialMiddleware_Regression_UserWithActiveTrial_ShouldAllowAccess() {
	user := authmodel.NewGoogleUser("testuser", "test@example.com", "google-1")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	subscription := model.NewSubscription(user.ID, "default_plan")
	subscription.IsTrialActive = true
	subscription.TrialStartDate = time.Now()
	subscription.TrialEndDate = time.Now().AddDate(0, 0, 7)
	subscription.SubscriptionStatus = "inactive"
	subscription.Origin = "web"

	err = suite.db.Create(subscription).Error
	suite.Require().NoError(err)

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

func (suite *TrialRegressionTestSuite) TestTrialMiddleware_Regression_UserWithoutSubscription_ShouldReturn403JSON() {
	user := authmodel.NewGoogleUser("testuser2", "test2@example.com", "google-2")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	var body map[string]any
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(suite.T(), false, body["success"])
	assert.Equal(suite.T(), "SUBSCRIPTION_REQUIRED", body["code"])
}

func (suite *TrialRegressionTestSuite) TestTrialMiddleware_Regression_UserWithExpiredTrial_ShouldReturn403JSON() {
	user := authmodel.NewGoogleUser("testuser3", "test3@example.com", "google-3")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	subscription := model.NewSubscription(user.ID, "default_plan")
	subscription.IsTrialActive = false
	subscription.TrialEndDate = time.Now().AddDate(0, 0, -1)
	subscription.SubscriptionStatus = "inactive"

	err = suite.db.Create(subscription).Error
	suite.Require().NoError(err)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	var body map[string]any
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(suite.T(), false, body["success"])
	assert.Equal(suite.T(), "SUBSCRIPTION_REQUIRED", body["code"])
}

func (suite *TrialRegressionTestSuite) TestUserRepository_Regression_PreloadSubscription_WorksCorrectly() {
	user := authmodel.NewGoogleUser("testuser4", "test4@example.com", "google-4")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	subscription := model.NewSubscription(user.ID, "test_plan")
	err = suite.db.Create(subscription).Error
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
