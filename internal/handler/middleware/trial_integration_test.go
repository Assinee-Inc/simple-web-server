package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TrialMiddlewareIntegrationTestSuite struct {
	suite.Suite
	userRepository repository.UserRepository
}

func TestTrialMiddlewareIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TrialMiddlewareIntegrationTestSuite))
}

func (suite *TrialMiddlewareIntegrationTestSuite) SetupSuite() {
	// Conectar ao banco de teste
	database.Connect()
	suite.userRepository = repository.NewGormUserRepository(database.DB)
}

func (suite *TrialMiddlewareIntegrationTestSuite) SetupTest() {
	// Limpar dados de teste
	database.DB.Exec("DELETE FROM subscriptions")
	database.DB.Exec("DELETE FROM users")
}

func (suite *TrialMiddlewareIntegrationTestSuite) createTestUserWithSubscription(inTrial bool, subscribed bool) *models.User {
	// Criar usuário com email único
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	user := models.NewUser("testuser", "password123", email)
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)

	// Criar subscription
	subscription := models.NewSubscription(user.ID, "test_plan")

	if !inTrial {
		subscription.IsTrialActive = false
		subscription.TrialEndDate = time.Now().AddDate(0, 0, -1) // Trial expirado
	}

	if subscribed {
		subscription.SubscriptionStatus = "active"
		subscription.StripeCustomerID = "cus_test123"
		subscription.StripeSubscriptionID = "sub_test123"
	}

	err = database.DB.Create(subscription).Error
	suite.Require().NoError(err)

	return user
}

func (suite *TrialMiddlewareIntegrationTestSuite) TestTrialMiddleware_WithRealUser_InTrialPeriod() {
	// Criar usuário com trial ativo
	user := suite.createTestUserWithSubscription(true, false)

	// Criar request com contexto do usuário
	req := httptest.NewRequest("GET", "/dashboard", nil)
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Handler que será chamado se o middleware passar
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Aplicar middleware
	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	// Verificar que o acesso foi permitido
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *TrialMiddlewareIntegrationTestSuite) TestTrialMiddleware_WithRealUser_Subscribed() {
	// Criar usuário inscrito
	user := suite.createTestUserWithSubscription(false, true)

	// Criar request com contexto do usuário
	req := httptest.NewRequest("GET", "/dashboard", nil)
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Handler que será chamado se o middleware passar
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Aplicar middleware
	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	// Verificar que o acesso foi permitido
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *TrialMiddlewareIntegrationTestSuite) TestTrialMiddleware_WithRealUser_NoTrialNoSubscription() {
	// Criar usuário sem trial e sem subscription
	user := suite.createTestUserWithSubscription(false, false)

	// Criar request com contexto do usuário
	req := httptest.NewRequest("GET", "/dashboard", nil)
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Handler que será chamado se o middleware passar
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Aplicar middleware
	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	// Verificar que foi redirecionado para settings
	assert.Equal(suite.T(), http.StatusSeeOther, w.Code)
	assert.Equal(suite.T(), "/settings", w.Header().Get("Location"))
}

func (suite *TrialMiddlewareIntegrationTestSuite) TestTrialMiddleware_WithRealUser_NoUser() {
	// Criar request sem contexto de usuário
	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()

	// Handler que será chamado se o middleware passar
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Aplicar middleware
	handler := middleware.TrialMiddleware(nextHandler)
	handler.ServeHTTP(w, req)

	// Verificar que foi redirecionado para login quando não há usuário no contexto
	assert.Equal(suite.T(), http.StatusSeeOther, w.Code)
	assert.Equal(suite.T(), "/login", w.Header().Get("Location"))
}

func (suite *TrialMiddlewareIntegrationTestSuite) TestTrialMiddleware_ExcludedPaths() {
	excludedPaths := []string{"/settings", "/logout"}

	for _, path := range excludedPaths {
		suite.Run("Path_"+path, func() {
			// Criar usuário sem trial e sem subscription
			user := suite.createTestUserWithSubscription(false, false)

			// Criar request com contexto do usuário
			req := httptest.NewRequest("GET", path, nil)
			ctx := context.WithValue(req.Context(), middleware.UserEmailKey, user.Email)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Handler que será chamado se o middleware passar
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Aplicar middleware
			handler := middleware.TrialMiddleware(nextHandler)
			handler.ServeHTTP(w, req)

			// Verificar que o acesso foi permitido mesmo sem trial/subscription
			assert.Equal(suite.T(), http.StatusOK, w.Code, "Path %s should be allowed", path)
		})
	}
}

func (suite *TrialMiddlewareIntegrationTestSuite) TestUserRepository_LoadsSubscriptionWithUser() {
	// Criar usuário com subscription
	user := suite.createTestUserWithSubscription(true, false)

	// Buscar usuário por email
	foundUser := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(foundUser)
	suite.Require().NotNil(foundUser.Subscription)

	// Verificar que a subscription foi carregada corretamente
	assert.Equal(suite.T(), user.ID, foundUser.Subscription.UserID)
	assert.Equal(suite.T(), "test_plan", foundUser.Subscription.PlanID)
	assert.True(suite.T(), foundUser.Subscription.IsTrialActive)
	assert.Equal(suite.T(), "inactive", foundUser.Subscription.SubscriptionStatus)

	// Verificar que os métodos do trial funcionam
	assert.True(suite.T(), foundUser.IsInTrialPeriod())
	assert.False(suite.T(), foundUser.IsSubscribed())
	assert.Greater(suite.T(), foundUser.DaysLeftInTrial(), 0)
}
