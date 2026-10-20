package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestPurchaseHandler_EbookNotFound_Returns404JSON(t *testing.T) {
	mockEbookSvc := new(mocks.MockEbookService)
	mockEbookSvc.On("FindByPublicID", "ebk_invalid").Return(nil, errors.New("not found"))

	h := NewPurchaseHandler(mockEbookSvc)

	req := httptest.NewRequest(http.MethodPost, "/purchase/create/ebk_invalid", strings.NewReader("clients[]=cli_1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "ebk_invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.PurchaseCreateHandler(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, false, body["success"])
}
