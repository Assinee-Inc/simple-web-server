package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
)

func TestAuthMiddleware_BearerJWT(t *testing.T) {
	tokenService := authsvc.NewTokenService("test-jwt-secret-key", 24*time.Hour)

	validUser := &authmodel.User{
		PublicID: "usr_12345",
		Username: "Maria Silva",
		Email:    "maria@mwtest.com",
		GoogleID: "google-sub-123",
	}
	validUser.ID = 77

	validToken, _, err := tokenService.GenerateToken(validUser)
	require.NoError(t, err)

	expiredTokenSvc := authsvc.NewTokenService("test-jwt-secret-key", -1*time.Hour)
	expiredToken, _, err := expiredTokenSvc.GenerateToken(validUser)
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUserID   uint
		expectEmail    string
	}{
		{
			name:           "permite acesso com token Bearer válido e injeta UserID/Email/User no contexto",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectUserID:   77,
			expectEmail:    "maria@mwtest.com",
		},
		{
			name:           "retorna 401 JSON quando header Authorization está ausente",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "retorna 401 JSON quando esquema não é Bearer",
			authHeader:     "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "retorna 401 JSON quando token é inválido ou adulterado",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "retorna 401 JSON quando token está expirado",
			authHeader:     "Bearer " + expiredToken,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedUserID uint
			var capturedEmail string
			var capturedUser *authmodel.User

			r := chi.NewRouter()
			r.Use(authmw.AuthMiddleware(tokenService))
			r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
				capturedUserID, _ = authmw.GetAuthenticatedUserID(r.Context())
				capturedEmail = authmw.GetAuthenticatedUserEmail(r.Context())
				capturedUser = authmw.GetAuthenticatedUser(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusUnauthorized {
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
				var errBody map[string]any
				err := json.Unmarshal(rec.Body.Bytes(), &errBody)
				require.NoError(t, err)
				assert.Equal(t, false, errBody["success"])
				assert.NotEmpty(t, errBody["error"])
				return
			}

			assert.Equal(t, tt.expectUserID, capturedUserID)
			assert.Equal(t, tt.expectEmail, capturedEmail)
			require.NotNil(t, capturedUser)
			assert.Equal(t, tt.expectUserID, capturedUser.ID)
			assert.Equal(t, tt.expectEmail, capturedUser.Email)
		})
	}
}
