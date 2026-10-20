package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authhandler "github.com/Assinee-Inc/docffy/internal/auth/handler"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGoogleLogin_NewUser_Returns201AndOnboardingRequired(t *testing.T) {
	mockUserService := new(mocks.MockUserService)
	mockCreatorService := new(mocks.MockCreatorService)
	tokenService := authsvc.NewTokenService("test-jwt-secret", 24*time.Hour)

	h := authhandler.NewAuthHandler(mockUserService, tokenService, mockCreatorService)

	now := time.Now()
	newUser := &authmodel.User{
		PublicID:        "usr_new_123",
		Username:        "Maria Silva",
		Email:           "maria@example.com",
		GoogleID:        "google-sub-123",
		EmailVerifiedAt: &now,
		TermsAcceptedAt: nil,
	}
	newUser.ID = 15

	mockUserService.On("AuthenticateWithGoogle", mock.Anything, "valid-google-id-token").Return(newUser, true, nil)
	mockCreatorService.On("FindCreatorByUserID", uint(15)).Return(nil, errors.New("criador não encontrado"))

	body, _ := json.Marshal(map[string]string{"id_token": "valid-google-id-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.GoogleLogin(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var resp authmodel.AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, int64(86400), resp.ExpiresIn)
	assert.True(t, resp.OnboardingRequired)
	assert.Equal(t, "usr_new_123", resp.User.PublicID)
	assert.Equal(t, "Maria Silva", resp.User.Username)
	assert.Equal(t, "maria@example.com", resp.User.Email)
	assert.Equal(t, "google-sub-123", resp.User.GoogleID)
	assert.True(t, resp.User.EmailVerified)
	assert.False(t, resp.User.TermsAccepted)
}

func TestGoogleLogin_ExistingCreator_Returns200AndOnboardingNotRequired(t *testing.T) {
	mockUserService := new(mocks.MockUserService)
	mockCreatorService := new(mocks.MockCreatorService)
	tokenService := authsvc.NewTokenService("test-jwt-secret", 24*time.Hour)

	h := authhandler.NewAuthHandler(mockUserService, tokenService, mockCreatorService)

	now := time.Now()
	existingUser := &authmodel.User{
		PublicID:        "usr_existing_77",
		Username:        "João Souza",
		Email:           "joao@example.com",
		GoogleID:        "google-sub-77",
		EmailVerifiedAt: &now,
		TermsAcceptedAt: &now,
	}
	existingUser.ID = 77

	existingCreator := &accountmodel.Creator{
		PublicID: "crt_77",
		UserID:   77,
		Name:     "João Souza",
		Email:    "joao@example.com",
	}

	mockUserService.On("AuthenticateWithGoogle", mock.Anything, "existing-google-token").Return(existingUser, false, nil)
	mockCreatorService.On("FindCreatorByUserID", uint(77)).Return(existingCreator, nil)

	body, _ := json.Marshal(map[string]string{"id_token": "existing-google-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.GoogleLogin(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp authmodel.AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.False(t, resp.OnboardingRequired)
	assert.Equal(t, "usr_existing_77", resp.User.PublicID)
	assert.True(t, resp.User.TermsAccepted)
}

func TestGoogleLogin_InvalidPayloadOrInvalidGoogleToken(t *testing.T) {
	mockUserService := new(mocks.MockUserService)
	mockCreatorService := new(mocks.MockCreatorService)
	tokenService := authsvc.NewTokenService("test-jwt-secret", 24*time.Hour)

	h := authhandler.NewAuthHandler(mockUserService, tokenService, mockCreatorService)

	t.Run("retorna 400 quando id_token está ausente", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/google", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.GoogleLogin(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, false, resp["success"])
	})

	t.Run("retorna 401 quando id_token Google é inválido ou não verificado", func(t *testing.T) {
		mockUserService.On("AuthenticateWithGoogle", mock.Anything, "invalid-token").Return(nil, false, authsvc.ErrInvalidGoogleToken)

		body, _ := json.Marshal(map[string]string{"id_token": "invalid-token"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.GoogleLogin(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Google")
	})
}

func TestMe_AuthenticatedUserAndCreator(t *testing.T) {
	mockUserService := new(mocks.MockUserService)
	mockCreatorService := new(mocks.MockCreatorService)
	tokenService := authsvc.NewTokenService("test-jwt-secret", 24*time.Hour)

	h := authhandler.NewAuthHandler(mockUserService, tokenService, mockCreatorService)

	now := time.Now()
	user := &authmodel.User{
		PublicID:        "usr_me_1",
		Username:        "Maria Silva",
		Email:           "maria@example.com",
		GoogleID:        "google-1",
		EmailVerifiedAt: &now,
		TermsAcceptedAt: &now,
	}
	user.ID = 21

	creator := &accountmodel.Creator{
		PublicID:               "crt_xyz789",
		UserID:                 21,
		Name:                   "Maria da Silva",
		SocialName:             "Maria Silva",
		Email:                  "maria@example.com",
		Phone:                  "11999999999",
		StripeConnectAccountID: "acct_123",
		OnboardingCompleted:    true,
		ChargesEnabled:         true,
		PayoutsEnabled:         true,
	}

	mockUserService.On("FindByID", uint(21)).Return(user)
	mockCreatorService.On("FindCreatorByUserID", uint(21)).Return(creator, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	ctx := authmw.WithAuthenticatedUser(req.Context(), user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, false, resp["onboarding_required"])

	userMap, ok := resp["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "usr_me_1", userMap["public_id"])
	assert.Equal(t, "maria@example.com", userMap["email"])

	creatorMap, ok := resp["creator"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "crt_xyz789", creatorMap["public_id"])
	assert.Equal(t, false, creatorMap["needs_stripe_onboarding"])
}

func TestLogout_Returns200JSON(t *testing.T) {
	mockUserService := new(mocks.MockUserService)
	mockCreatorService := new(mocks.MockCreatorService)
	tokenService := authsvc.NewTokenService("test-jwt-secret", 24*time.Hour)

	h := authhandler.NewAuthHandler(mockUserService, tokenService, mockCreatorService)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Sessão encerrada com sucesso", resp["message"])
}

func TestAuthHandler_LegacySSRViewsRemoved(t *testing.T) {
	h := &authhandler.AuthHandler{}
	handlerType := reflect.TypeOf(h)

	legacyMethods := []string{
		"LoginView",
		"ForgetPasswordView",
		"ForgetPasswordSubmit",
		"ResetPasswordView",
		"ResetPasswordSubmit",
		"CheckEmailView",
		"ConfirmAccountView",
		"EmailNotVerifiedView",
		"ResendConfirmationSubmit",
	}

	for _, method := range legacyMethods {
		_, found := handlerType.MethodByName(method)
		assert.Falsef(t, found, "Método SSR legado %q deve ter sido removido de AuthHandler", method)
	}
}
