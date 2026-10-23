package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76"
)

type mockStripeConnectSvcForMW struct {
	mock.Mock
}

func (m *mockStripeConnectSvcForMW) CreateConnectAccount(c *accountmodel.Creator) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}
func (m *mockStripeConnectSvcForMW) CreateOnboardingLink(id, ref, ret string) (string, error) {
	args := m.Called(id, ref, ret)
	return args.String(0), args.Error(1)
}
func (m *mockStripeConnectSvcForMW) GetAccountDetails(id string) (*stripe.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Account), args.Error(1)
}
func (m *mockStripeConnectSvcForMW) UpdateCreatorFromAccount(c *accountmodel.Creator, acc *stripe.Account) error {
	return m.Called(c, acc).Error(0)
}
func (m *mockStripeConnectSvcForMW) GerarLinkRemediacao(c *accountmodel.Creator) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}
func (m *mockStripeConnectSvcForMW) AnalyzeRequirements(acc *stripe.Account) (*accountsvc.AccountRequirementsStatus, error) {
	args := m.Called(acc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountsvc.AccountRequirementsStatus), args.Error(1)
}
func (m *mockStripeConnectSvcForMW) GetPendingDocuments(acc *stripe.Account) []accountsvc.PendingRequirement {
	return nil
}
func (m *mockStripeConnectSvcForMW) HasCriticalPendencies(s *accountsvc.AccountRequirementsStatus) bool {
	return false
}
func (m *mockStripeConnectSvcForMW) ShouldGenerateRemediationLink(s *accountsvc.AccountRequirementsStatus) bool {
	return false
}

func TestStripeOnboardingMiddleware_JSONResponses(t *testing.T) {
	t.Run("retorna 403 JSON STRIPE_ONBOARDING_REQUIRED quando StripeConnectAccountID está vazio", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockStripeSvc := new(mockStripeConnectSvcForMW)

		creator := &accountmodel.Creator{
			PublicID:               "crt_1",
			UserID:                 10,
			Email:                  "maria@example.com",
			StripeConnectAccountID: "",
			OnboardingCompleted:    false,
		}
		mockCreatorSvc.On("FindCreatorByUserID", uint(10)).Return(creator, nil)

		r := chi.NewRouter()
		r.Use(accountmw.StripeOnboardingMiddleware(mockCreatorSvc, mockStripeSvc))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		user := &authmodel.User{Email: "maria@example.com"}
		user.ID = 10
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, false, body["success"])
		assert.Equal(t, "STRIPE_ONBOARDING_REQUIRED", body["code"])
		assert.NotEmpty(t, body["error"])
	})

	t.Run("retorna 403 JSON STRIPE_ONBOARDING_REQUIRED quando conta Stripe tem pendências após consulta", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockStripeSvc := new(mockStripeConnectSvcForMW)

		creator := &accountmodel.Creator{
			PublicID:               "crt_2",
			UserID:                 11,
			Email:                  "pendente@example.com",
			StripeConnectAccountID: "acct_pending",
			OnboardingCompleted:    false,
			ChargesEnabled:         false,
			PayoutsEnabled:         false,
		}
		mockCreatorSvc.On("FindCreatorByUserID", uint(11)).Return(creator, nil)

		stripeAcc := &stripe.Account{
			ID:               "acct_pending",
			DetailsSubmitted: false,
			ChargesEnabled:   false,
			PayoutsEnabled:   false,
		}
		mockStripeSvc.On("GetAccountDetails", "acct_pending").Return(stripeAcc, nil)
		mockStripeSvc.On("UpdateCreatorFromAccount", creator, stripeAcc).Return(nil)

		r := chi.NewRouter()
		r.Use(accountmw.StripeOnboardingMiddleware(mockCreatorSvc, mockStripeSvc))
		r.Get("/ebook", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/ebook", nil)
		user := &authmodel.User{Email: "pendente@example.com"}
		user.ID = 11
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, false, body["success"])
		assert.Equal(t, "STRIPE_ONBOARDING_REQUIRED", body["code"])
	})

	t.Run("permite acesso quando Creator tem Stripe Connect ativo e completo", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockStripeSvc := new(mockStripeConnectSvcForMW)

		creator := &accountmodel.Creator{
			PublicID:               "crt_active",
			UserID:                 12,
			Email:                  "ativo@example.com",
			StripeConnectAccountID: "acct_active",
			OnboardingCompleted:    true,
			ChargesEnabled:         true,
			PayoutsEnabled:         true,
		}
		mockCreatorSvc.On("FindCreatorByUserID", uint(12)).Return(creator, nil)

		r := chi.NewRouter()
		r.Use(accountmw.StripeOnboardingMiddleware(mockCreatorSvc, mockStripeSvc))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		user := &authmodel.User{Email: "ativo@example.com"}
		user.ID = 12
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("permite acesso livre às rotas excluídas (/stripe-connect/*)", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockStripeSvc := new(mockStripeConnectSvcForMW)

		r := chi.NewRouter()
		r.Use(accountmw.StripeOnboardingMiddleware(mockCreatorSvc, mockStripeSvc))
		r.Get("/stripe-connect/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/stripe-connect/status", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
