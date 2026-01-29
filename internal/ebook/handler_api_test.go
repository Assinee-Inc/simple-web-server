package ebook_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/anglesson/simple-web-server/internal/ebook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEbookHandler_CreateEbook_Success(t *testing.T) {
	// Arrange
	mockMgr := new(MockEbookManager)
	handler := ebook.NewHandler(mockMgr)

	reqDto := ebook.CreateEbookRequest{
		Title:            "Test Ebook",
		Price:            1999,
		PromotionalPrice: 999,
		CoverImage:       "https://any-url.com",
		InfoProducerID:   "producer-123",
		FileIDs:          []string{"a9a7275e-9339-4973-9d16-1b343b4b834b"},
	}

	expectedEbook := &ebook.Ebook{
		ID:               "ebook-456",
		Title:            "Test Ebook",
		Price:            1999,
		PromotionalPrice: 999,
		CoverImage:       "https://any-url.com",
		InfoProducerID:   "producer-123",
		FileIDs:          []string{"a9a7275e-9339-4973-9d16-1b343b4b834b"},
	}

	body, _ := json.Marshal(reqDto)
	req := httptest.NewRequest("POST", "/ebooks", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// Mocking the manager call
	mockMgr.On("CreateEbook", req.Context(), mock.MatchedBy(func(e *ebook.Ebook) bool {
		return (e.Title == reqDto.ToDomain().Title &&
			e.Price == reqDto.ToDomain().Price &&
			e.PromotionalPrice == reqDto.ToDomain().PromotionalPrice &&
			e.InfoProducerID == reqDto.ToDomain().InfoProducerID) &&
			(slices.Compare(e.FileIDs, reqDto.FileIDs) == 0)

	})).Return(expectedEbook, nil)

	// Act
	handler.CreateEbook(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)

	var returnedEbook ebook.Ebook
	err := json.Unmarshal(rr.Body.Bytes(), &returnedEbook)
	assert.NoError(t, err)
	assert.Equal(t, expectedEbook.ID, returnedEbook.ID)
	assert.Equal(t, expectedEbook.Title, returnedEbook.Title)

	mockMgr.AssertExpectations(t)
}

func TestHandler_Create_InvalidPayload(t *testing.T) {
	mockMgr := new(MockEbookManager)
	handler := ebook.NewHandler(mockMgr)

	invalidReq := map[string]any{"price": "vinte reais"}

	body, _ := json.Marshal(invalidReq)
	req := httptest.NewRequest("POST", "/ebooks", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler.CreateEbook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	mockMgr.AssertNotCalled(t, "CreateEbook", mock.Anything, mock.Anything)
}

func TestEbookHandler_CreateEbook_Validation(t *testing.T) {
	mockMgr := new(MockEbookManager)
	handler := ebook.NewHandler(nil)

	tests := []struct {
		name          string
		payload       ebook.CreateEbookRequest
		expectedField string
		expectedMsg   string
	}{
		{
			name: "título obrigatório",
			payload: ebook.CreateEbookRequest{
				Price: 1990,
				// Title is missing
			},
			expectedField: "title",
			expectedMsg:   "é obrigatório",
		},
		{
			name: "título muito longo",
			payload: ebook.CreateEbookRequest{
				Title: string(make([]byte, 121)), // Invalid title
				Price: 1990,
			},
			expectedField: "title",
			expectedMsg:   "máximo é 120 caracteres",
		},
		{
			name: "preço obrigatório",
			payload: ebook.CreateEbookRequest{
				Title: string(make([]byte, 120)),
				// Price is missing
			},
			expectedField: "price",
			expectedMsg:   "é obrigatório",
		},
		{
			name: "descrição muito longa",
			payload: ebook.CreateEbookRequest{
				Title:       "Valid Title",
				Description: string(make([]byte, 121)), // Invalid description
			},
			expectedField: "description",
			expectedMsg:   "máximo é 120 caracteres",
		},
		{
			name: "descrição de venda muito longa",
			payload: ebook.CreateEbookRequest{
				Title:            "Valid Title",
				SalesDescription: string(make([]byte, 256)), // Invalid sales description
			},
			expectedField: "sales_description",
			expectedMsg:   "máximo é 255 caracteres",
		},
		{
			name: "preço promocional maior que o preço",
			payload: ebook.CreateEbookRequest{
				Price:            1990,
				PromotionalPrice: 2000, // Invalid promotional price
			},
			expectedField: "promotional_price",
			expectedMsg:   "O campo 'promotional_price' deve ser menor que o campo 'price'",
		},
		{
			name: "url inválida",
			payload: ebook.CreateEbookRequest{
				CoverImage: "invalid-url",
			},
			expectedField: "cover_image",
			expectedMsg:   "Campo inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/ebooks", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler.CreateEbook(rr, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

			var errResp ebook.ErrorResponse
			err := json.Unmarshal(rr.Body.Bytes(), &errResp)

			assert.NoError(t, err)
			assert.Contains(t, errResp.Fields[tt.expectedField], tt.expectedMsg)

			mockMgr.AssertNotCalled(t, "CreateEbook", mock.Anything, mock.Anything)
		})
	}
}
