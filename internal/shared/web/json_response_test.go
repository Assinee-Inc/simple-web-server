package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		payload        any
		expectedStatus int
		expectedBody   map[string]any
	}{
		{
			name:   "retorna 200 OK com payload estruturado",
			status: http.StatusOK,
			payload: map[string]any{
				"success": true,
				"message": "Operação concluída",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"success": true,
				"message": "Operação concluída",
			},
		},
		{
			name:   "retorna 201 Created com struct serializada",
			status: http.StatusCreated,
			payload: struct {
				Success bool   `json:"success"`
				ID      string `json:"id"`
			}{
				Success: true,
				ID:      "usr_123",
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]any{
				"success": true,
				"id":      "usr_123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			RespondJSON(rec, tt.status, tt.payload)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var got map[string]any
			err := json.Unmarshal(rec.Body.Bytes(), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, got)
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		code           string
		message        string
		details        any
		expectedStatus int
		expectedBody   map[string]any
	}{
		{
			name:           "erro simples sem code e sem details",
			status:         http.StatusUnauthorized,
			code:           "",
			message:        "Não autorizado",
			details:        nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]any{
				"success": false,
				"error":   "Não autorizado",
			},
		},
		{
			name:           "erro com code semântico e details",
			status:         http.StatusForbidden,
			code:           "CREATOR_ONBOARDING_REQUIRED",
			message:        "Conclua o cadastro de criador para acessar este recurso",
			details:        map[string]any{"step": "onboarding"},
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]any{
				"success": false,
				"code":    "CREATOR_ONBOARDING_REQUIRED",
				"error":   "Conclua o cadastro de criador para acessar este recurso",
				"details": map[string]any{"step": "onboarding"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			RespondError(rec, tt.status, tt.code, tt.message, tt.details)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

			var got map[string]any
			err := json.Unmarshal(rec.Body.Bytes(), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, got)
		})
	}
}

func TestRespondValidationError(t *testing.T) {
	rec := httptest.NewRecorder()
	details := map[string]string{
		"cpf":   "CPF inválido",
		"email": "E-mail obrigatório",
	}

	RespondValidationError(rec, details)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	var got map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &got)
	require.NoError(t, err)
	assert.Equal(t, false, got["success"])
	assert.Equal(t, "VALIDATION_ERROR", got["code"])
	assert.NotEmpty(t, got["error"])
	assert.Equal(t, map[string]any{
		"cpf":   "CPF inválido",
		"email": "E-mail obrigatório",
	}, got["details"])
}
