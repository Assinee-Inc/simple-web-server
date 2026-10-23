package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	accounthandler "github.com/Assinee-Inc/docffy/internal/account/handler"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76"
)

type mockStripeConnectService struct {
	mock.Mock
}

func (m *mockStripeConnectService) CreateConnectAccount(creator *accountmodel.Creator) (string, error) {
	args := m.Called(creator)
	return args.String(0), args.Error(1)
}
func (m *mockStripeConnectService) CreateOnboardingLink(accountID, refreshURL, returnURL string) (string, error) {
	args := m.Called(accountID, refreshURL, returnURL)
	return args.String(0), args.Error(1)
}
func (m *mockStripeConnectService) GetAccountDetails(accountID string) (*stripe.Account, error) {
	args := m.Called(accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.Account), args.Error(1)
}
func (m *mockStripeConnectService) UpdateCreatorFromAccount(creator *accountmodel.Creator, account *stripe.Account) error {
	args := m.Called(creator, account)
	return args.Error(0)
}
func (m *mockStripeConnectService) GerarLinkRemediacao(creator *accountmodel.Creator) (string, error) {
	args := m.Called(creator)
	return args.String(0), args.Error(1)
}
func (m *mockStripeConnectService) AnalyzeRequirements(account *stripe.Account) (*accountsvc.AccountRequirementsStatus, error) {
	args := m.Called(account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*accountsvc.AccountRequirementsStatus), args.Error(1)
}
func (m *mockStripeConnectService) GetPendingDocuments(account *stripe.Account) []accountsvc.PendingRequirement {
	args := m.Called(account)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]accountsvc.PendingRequirement)
}
func (m *mockStripeConnectService) HasCriticalPendencies(status *accountsvc.AccountRequirementsStatus) bool {
	return m.Called(status).Bool(0)
}
func (m *mockStripeConnectService) ShouldGenerateRemediationLink(status *accountsvc.AccountRequirementsStatus) bool {
	return m.Called(status).Bool(0)
}

func TestCreatorHandler_OnboardCreator(t *testing.T) {
	t.Run("retorna 201 Created com dados do criador e provisiona conta Stripe Connect", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockStripeSvc := new(mockStripeConnectService)
		h := accounthandler.NewCreatorHandler(mockCreatorSvc, mockStripeSvc)

		user := &authmodel.User{
			PublicID: "usr_123",
			Email:    "maria@example.com",
		}
		user.ID = 10

		createdCreator := &accountmodel.Creator{
			PublicID:            "crt_xyz789",
			UserID:              10,
			Name:                "Maria da Silva",
			SocialName:          "Maria Silva",
			Email:               "maria@example.com",
			CPF:                 "52998224725",
			Phone:               "11999999999",
			OnboardingCompleted: false,
		}
		createdCreator.ID = 50

		mockCreatorSvc.On("OnboardCreator", uint(10), accountsvc.InputOnboardingCreator{
			Name:          "Maria da Silva",
			SocialName:    "Maria Silva",
			CPF:           "52998224725",
			BirthDate:     "15/05/1990",
			PhoneNumber:   "11999999999",
			TermsAccepted: true,
		}).Return(createdCreator, nil)

		mockStripeSvc.On("CreateConnectAccount", createdCreator).Return("acct_123", nil)
		mockCreatorSvc.On("UpdateCreator", mock.MatchedBy(func(c *accountmodel.Creator) bool {
			return c.StripeConnectAccountID == "acct_123"
		})).Return(nil)

		payload, _ := json.Marshal(map[string]any{
			"name":           "Maria da Silva",
			"social_name":    "Maria Silva",
			"cpf":            "52998224725",
			"birthdate":      "15/05/1990",
			"phone":          "11999999999",
			"terms_accepted": true,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/account/onboarding", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		h.OnboardCreator(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, true, resp["success"])
		creatorMap, ok := resp["creator"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "crt_xyz789", creatorMap["public_id"])
		assert.Equal(t, "Maria da Silva", creatorMap["name"])
		assert.Equal(t, "acct_123", creatorMap["stripe_connect_account_id"])
		assert.Equal(t, false, creatorMap["onboarding_completed"])
	})

	t.Run("retorna 400 Bad Request quando validação falha (ex: menor de 18 anos ou CPF inválido)", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockStripeSvc := new(mockStripeConnectService)
		h := accounthandler.NewCreatorHandler(mockCreatorSvc, mockStripeSvc)

		user := &authmodel.User{Email: "maria@example.com"}
		user.ID = 10

		mockCreatorSvc.On("OnboardCreator", uint(10), mock.Anything).Return(nil, errors.New("o criador deve ter 18 anos ou mais"))

		payload, _ := json.Marshal(map[string]any{
			"name":           "Menor",
			"cpf":            "52998224725",
			"birthdate":      "15/05/2015",
			"phone":          "11999999999",
			"terms_accepted": true,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/account/onboarding", bytes.NewReader(payload))
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		h.OnboardCreator(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "18 anos")
	})

	t.Run("retorna 409 Conflict quando CPF já existe", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockStripeSvc := new(mockStripeConnectService)
		h := accounthandler.NewCreatorHandler(mockCreatorSvc, mockStripeSvc)

		user := &authmodel.User{Email: "maria@example.com"}
		user.ID = 10

		mockCreatorSvc.On("OnboardCreator", uint(10), mock.Anything).Return(nil, accountsvc.ErrCreatorAlreadyExists)

		payload, _ := json.Marshal(map[string]any{
			"name":           "Maria da Silva",
			"cpf":            "52998224725",
			"birthdate":      "15/05/1990",
			"phone":          "11999999999",
			"terms_accepted": true,
		})

		req := httptest.NewRequest(http.MethodPost, "/api/account/onboarding", bytes.NewReader(payload))
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		h.OnboardCreator(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, false, resp["success"])
	})

	t.Run("retorna 401 Unauthorized quando não há usuário autenticado", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockStripeSvc := new(mockStripeConnectService)
		h := accounthandler.NewCreatorHandler(mockCreatorSvc, mockStripeSvc)

		req := httptest.NewRequest(http.MethodPost, "/api/account/onboarding", bytes.NewReader([]byte(`{}`)))
		rec := httptest.NewRecorder()

		h.OnboardCreator(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("remove métodos SSR legados RegisterView e RegisterCreatorSSR", func(t *testing.T) {
		h := &accounthandler.CreatorHandler{}
		hType := reflect.TypeOf(h)
		_, hasRegView := hType.MethodByName("RegisterView")
		_, hasRegSSR := hType.MethodByName("RegisterCreatorSSR")
		assert.False(t, hasRegView)
		assert.False(t, hasRegSSR)
	})
}
