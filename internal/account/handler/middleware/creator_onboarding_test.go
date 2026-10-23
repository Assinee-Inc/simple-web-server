package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatorOnboardingMiddleware(t *testing.T) {
	t.Run("retorna 403 JSON CREATOR_ONBOARDING_REQUIRED quando usuário autenticado não possui Creator", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		mockCreatorSvc.On("FindCreatorByUserID", uint(10)).Return(nil, errors.New("criador não encontrado"))

		r := chi.NewRouter()
		r.Use(accountmw.CreatorOnboardingMiddleware(mockCreatorSvc))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		user := &authmodel.User{Email: "novo@example.com"}
		user.ID = 10
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, false, body["success"])
		assert.Equal(t, "CREATOR_ONBOARDING_REQUIRED", body["code"])
		assert.NotEmpty(t, body["error"])
	})

	t.Run("retorna 401 JSON quando não há usuário autenticado no contexto", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)

		r := chi.NewRouter()
		r.Use(accountmw.CreatorOnboardingMiddleware(mockCreatorSvc))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, false, body["success"])
	})

	t.Run("permite acesso e injeta Creator no contexto quando Creator existe", func(t *testing.T) {
		mockCreatorSvc := new(mocks.MockCreatorService)
		expectedCreator := &accountmodel.Creator{
			PublicID: "crt_abc",
			UserID:   25,
			Name:     "Maria da Silva",
			Email:    "maria@example.com",
		}
		expectedCreator.ID = 88

		mockCreatorSvc.On("FindCreatorByUserID", uint(25)).Return(expectedCreator, nil)

		var capturedCreator *accountmodel.Creator
		r := chi.NewRouter()
		r.Use(accountmw.CreatorOnboardingMiddleware(mockCreatorSvc))
		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			capturedCreator = accountmw.CreatorFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		user := &authmodel.User{Email: "maria@example.com"}
		user.ID = 25
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, capturedCreator)
		assert.Equal(t, uint(88), capturedCreator.ID)
		assert.Equal(t, "crt_abc", capturedCreator.PublicID)
	})
}
