package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHomeHandler_HomeView_ReturnsJSONStatus(t *testing.T) {
	h := NewHomeHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.HomeView(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var body map[string]any
	err := json.NewDecoder(w.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "ok", body["status"])
}

func TestErrorHandler_ErrorView_ReturnsJSONError(t *testing.T) {
	h := NewErrorHandler()

	tests := []struct {
		name         string
		code         int
		expectedCode string
	}{
		{"404 Not Found", http.StatusNotFound, "NOT_FOUND"},
		{"405 Method Not Allowed", http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED"},
		{"500 Internal Server Error", http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			h.ErrorView(w, req, tc.code)

			assert.Equal(t, tc.code, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

			var body map[string]any
			err := json.NewDecoder(w.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, false, body["success"])
			assert.Equal(t, tc.expectedCode, body["code"])
		})
	}
}
