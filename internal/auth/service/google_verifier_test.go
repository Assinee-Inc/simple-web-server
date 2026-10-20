package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleTokenVerifier_VerifyIDToken(t *testing.T) {
	expectedClientID := "my-google-client-id.apps.googleusercontent.com"
	futureExp := strconv.FormatInt(time.Now().Add(1*time.Hour).Unix(), 10)
	pastExp := strconv.FormatInt(time.Now().Add(-1*time.Hour).Unix(), 10)

	tests := []struct {
		name           string
		idToken        string
		clientID       string
		serverStatus   int
		serverResponse map[string]any
		expectErr      bool
		expectedSub    string
		expectedEmail  string
		expectedName   string
	}{
		{
			name:         "sucesso com iss https://accounts.google.com e email_verified true",
			idToken:      "valid-id-token-1",
			clientID:     expectedClientID,
			serverStatus: http.StatusOK,
			serverResponse: map[string]any{
				"iss":            "https://accounts.google.com",
				"aud":            expectedClientID,
				"sub":            "google-sub-12345",
				"email":          "maria@example.com",
				"email_verified": "true",
				"name":           "Maria Silva",
				"exp":            futureExp,
			},
			expectErr:     false,
			expectedSub:   "google-sub-12345",
			expectedEmail: "maria@example.com",
			expectedName:  "Maria Silva",
		},
		{
			name:         "sucesso com iss accounts.google.com e email_verified booleano true",
			idToken:      "valid-id-token-2",
			clientID:     expectedClientID,
			serverStatus: http.StatusOK,
			serverResponse: map[string]any{
				"iss":            "accounts.google.com",
				"aud":            expectedClientID,
				"sub":            "google-sub-999",
				"email":          "joao@example.com",
				"email_verified": true,
				"name":           "João Souza",
				"exp":            futureExp,
			},
			expectErr:     false,
			expectedSub:   "google-sub-999",
			expectedEmail: "joao@example.com",
			expectedName:  "João Souza",
		},
		{
			name:         "rejeita quando aud diverge do GoogleClientID configurado",
			idToken:      "wrong-aud-token",
			clientID:     expectedClientID,
			serverStatus: http.StatusOK,
			serverResponse: map[string]any{
				"iss":            "https://accounts.google.com",
				"aud":            "other-client-id.apps.googleusercontent.com",
				"sub":            "google-sub-12345",
				"email":          "maria@example.com",
				"email_verified": "true",
				"name":           "Maria Silva",
				"exp":            futureExp,
			},
			expectErr: true,
		},
		{
			name:         "rejeita quando iss é inválido",
			idToken:      "wrong-iss-token",
			clientID:     expectedClientID,
			serverStatus: http.StatusOK,
			serverResponse: map[string]any{
				"iss":            "https://evil.example.com",
				"aud":            expectedClientID,
				"sub":            "google-sub-12345",
				"email":          "maria@example.com",
				"email_verified": "true",
				"name":           "Maria Silva",
				"exp":            futureExp,
			},
			expectErr: true,
		},
		{
			name:         "rejeita quando exp está no passado",
			idToken:      "expired-token",
			clientID:     expectedClientID,
			serverStatus: http.StatusOK,
			serverResponse: map[string]any{
				"iss":            "https://accounts.google.com",
				"aud":            expectedClientID,
				"sub":            "google-sub-12345",
				"email":          "maria@example.com",
				"email_verified": "true",
				"name":           "Maria Silva",
				"exp":            pastExp,
			},
			expectErr: true,
		},
		{
			name:         "rejeita quando email_verified é false",
			idToken:      "unverified-email-token",
			clientID:     expectedClientID,
			serverStatus: http.StatusOK,
			serverResponse: map[string]any{
				"iss":            "https://accounts.google.com",
				"aud":            expectedClientID,
				"sub":            "google-sub-12345",
				"email":          "maria@example.com",
				"email_verified": "false",
				"name":           "Maria Silva",
				"exp":            futureExp,
			},
			expectErr: true,
		},
		{
			name:         "rejeita quando endpoint Google retorna status diferente de 200",
			idToken:      "invalid-token",
			clientID:     expectedClientID,
			serverStatus: http.StatusBadRequest,
			serverResponse: map[string]any{
				"error_description": "Invalid Value",
			},
			expectErr: true,
		},
		{
			name:      "rejeita quando idToken é vazio",
			idToken:   "",
			clientID:  expectedClientID,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.idToken, r.URL.Query().Get("id_token"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				_ = json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			verifier := authsvc.NewGoogleOIDCVerifierWithEndpoint(tt.clientID, server.URL, server.Client())
			identity, err := verifier.VerifyIDToken(context.Background(), tt.idToken)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, identity)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, identity)
			assert.Equal(t, tt.expectedSub, identity.Sub)
			assert.Equal(t, tt.expectedEmail, identity.Email)
			assert.Equal(t, tt.expectedName, identity.Name)
			assert.True(t, identity.EmailVerified)
		})
	}
}
