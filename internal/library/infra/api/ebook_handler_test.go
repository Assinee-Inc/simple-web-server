package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anglesson/simple-web-server/internal/library/domain"
)

type EbookServiceMock struct {
	CreateCalled   bool
	CreateEbookArg *domain.Ebook
}

func (e *EbookServiceMock) CreateEbook(ebook domain.Ebook) (*domain.Ebook, error) {
	e.CreateCalled = true
	e.CreateEbookArg = &ebook
	return &ebook, nil
}

func TestCreateEbook(t *testing.T) {
	t.Run("Create ebook with success", func(t *testing.T) {
		ebookSvc := &EbookServiceMock{}

		payload := map[string]string{"title": "Any Title"}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/ebooks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()

		ebookHandler := NewEbookHandler(ebookSvc)
		ebookHandler.CreateEbook(response, req)

		var ebook domain.Ebook
		bodyBytes := response.Body.Bytes()

		err := json.Unmarshal(bodyBytes, &ebook)
		if err != nil {
			t.Errorf("Failed to parse server response '%s' into map: %v", string(bodyBytes), err)
		}

		if ebook.Title != payload["title"] {
			t.Errorf("Error to compare values want '%s' got '%s'", payload["title"], ebook.Title)
		}

		if !ebookSvc.CreateCalled {
			t.Errorf("Store ebook method was not called")
		}

		if ebookSvc.CreateEbookArg.Title != payload["title"] {
			t.Errorf("Store ebook method was called with wrong ebook title")
		}

		if status := response.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}
	})

}
