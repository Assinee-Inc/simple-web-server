package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/pkg/database"
)

// EmailVerificationMiddlewareTestSuite testa o comportamento do AuthMiddleware
// em relação à verificação de e-mail, garantindo que não ocorra loop de redirects.
type EmailVerificationMiddlewareTestSuite struct {
	suite.Suite
	userRepository authrepo.UserRepository
	mockSession    *mocks.MockSessionService
	router         *chi.Mux
}

func TestEmailVerificationMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(EmailVerificationMiddlewareTestSuite))
}

func (suite *EmailVerificationMiddlewareTestSuite) SetupSuite() {
	database.Connect()
	suite.userRepository = authrepo.NewGormUserRepository(database.DB)
}

func (suite *EmailVerificationMiddlewareTestSuite) SetupTest() {
	database.DB.Exec("DELETE FROM users WHERE email LIKE '%@mwtest.com'")

	suite.mockSession = new(mocks.MockSessionService)
	suite.router = suite.buildRouter()
}

// buildRouter monta um roteador com a mesma estrutura do main.go:
// /dashboard          → protegido por AuthMiddleware
// /email-not-verified → público, sem AuthGuard
func (suite *EmailVerificationMiddlewareTestSuite) buildRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(authmw.AuthMiddleware(suite.mockSession))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	// SEM AuthGuard — replica a correção feita em main.go
	r.Get("/email-not-verified", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}

func (suite *EmailVerificationMiddlewareTestSuite) setupSessionForUser(email string) {
	suite.mockSession.On("Get", mock.Anything, authsvc.UserEmailKey).Return(email)
	suite.mockSession.On("Get", mock.Anything, authsvc.CSRFTokenKey).Return("csrf-test-token")
}

// TestUnverifiedUser_DashboardRedirectsToEmailNotVerified garante que um usuário
// autenticado sem e-mail verificado é redirecionado para /email-not-verified ao
// tentar acessar o dashboard.
func (suite *EmailVerificationMiddlewareTestSuite) TestUnverifiedUser_DashboardRedirectsToEmailNotVerified() {
	user := suite.createUnverifiedUser()
	suite.setupSessionForUser(user.Email)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusSeeOther, w.Code)
	assert.Equal(suite.T(), "/email-not-verified", w.Header().Get("Location"))
}

// TestUnverifiedUser_EmailNotVerifiedPageIsAccessible garante que o usuário
// autenticado sem e-mail verificado consegue acessar /email-not-verified sem
// ser redirecionado de volta ao /dashboard — prevenindo o loop de redirects.
func (suite *EmailVerificationMiddlewareTestSuite) TestUnverifiedUser_EmailNotVerifiedPageIsAccessible() {
	req := httptest.NewRequest("GET", "/email-not-verified", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.NotEqual(suite.T(), "/dashboard", w.Header().Get("Location"),
		"REGRESSÃO: /email-not-verified não deve redirecionar para /dashboard — isso causaria loop infinito")
}

// TestRedirectLoopDoesNotOccur simula o cenário completo que causou o bug:
// browser em /dashboard → redirect /email-not-verified → 200 (sem continuar o loop).
func (suite *EmailVerificationMiddlewareTestSuite) TestRedirectLoopDoesNotOccur() {
	user := suite.createUnverifiedUser()
	suite.setupSessionForUser(user.Email)

	// Passo 1: usuário tenta acessar /dashboard
	req1 := httptest.NewRequest("GET", "/dashboard", nil)
	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, req1)

	assert.Equal(suite.T(), http.StatusSeeOther, w1.Code)
	nextURL := w1.Header().Get("Location")
	assert.Equal(suite.T(), "/email-not-verified", nextURL)

	// Passo 2: browser segue o redirect para /email-not-verified
	req2 := httptest.NewRequest("GET", nextURL, nil)
	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	// Deve parar aqui com 200 — não redirecionar para /dashboard novamente
	assert.Equal(suite.T(), http.StatusOK, w2.Code,
		"REGRESSÃO: /email-not-verified retornou redirect em vez de 200 — loop de redirects detectado")
}

// TestVerifiedUser_DashboardIsAccessible garante que um usuário com e-mail
// verificado acessa o dashboard normalmente, sem ser bloqueado pelo gate.
func (suite *EmailVerificationMiddlewareTestSuite) TestVerifiedUser_DashboardIsAccessible() {
	user := suite.createVerifiedUser()
	suite.setupSessionForUser(user.Email)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// --- helpers ---

func (suite *EmailVerificationMiddlewareTestSuite) createUnverifiedUser() *authmodel.User {
	user := authmodel.NewUser("Unverified User", "hashed", "unverified@mwtest.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)
	return user
}

func (suite *EmailVerificationMiddlewareTestSuite) createVerifiedUser() *authmodel.User {
	user := authmodel.NewUser("Verified User", "hashed", "verified@mwtest.com")
	err := suite.userRepository.Create(user)
	suite.Require().NoError(err)
	database.DB.Exec("UPDATE users SET email_verified_at = datetime('now') WHERE id = ?", user.ID)
	loaded := suite.userRepository.FindByUserEmail(user.Email)
	suite.Require().NotNil(loaded)
	return loaded
}
