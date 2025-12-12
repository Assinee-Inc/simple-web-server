package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anglesson/simple-web-server/internal/library/handlers"
	"github.com/anglesson/simple-web-server/internal/library/models"
	"github.com/stretchr/testify/mock"
)

type EbookServiceMock struct {
	mock.Mock
}

func (e *EbookServiceMock) CreateEbook(ebook models.Ebook) (*models.Ebook, error) {
	args := e.Called(ebook)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Ebook), nil
}

var sut *handlers.EbookHandler
var ebookSvc *EbookServiceMock

func init() {
	ebookSvc = new(EbookServiceMock)
	sut = handlers.NewEbookHandler(ebookSvc)
}

func TestCreateEbook(t *testing.T) {
	t.Run("Create ebook with success", func(t *testing.T) {
		// Define the expected ebook that the service mock will return
		ebookSvc.On("CreateEbook", mock.AnythingOfType("models.Ebook")).Return(&models.Ebook{
			Title:          "Any Title",
			ID:             "any-id",
			InfoProducerID: "any-producer-id",
		}, nil)

		//Prepare the request payload
		requestPayload := map[string]string{"title": "Any Title", "info_produtor_id": "any-producer-id"}
		body, err := json.Marshal(requestPayload)
		if err != nil {
			t.Fatalf("Failed to marshal request payload: %v", err)
		}

		// Create a new HTTP request
		req := httptest.NewRequest("POST", "/api/ebooks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder to capture the handler's response
		response := httptest.NewRecorder()

		// handler and call the CreateEbook method
		sut.CreateEbook(response, req)

		// Assert the HTTP status code
		if status := response.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}

		// Parse the response body
		var actualEbook models.Ebook
		bodyBytes := response.Body.Bytes()
		err = json.Unmarshal(bodyBytes, &actualEbook)
		if err != nil {
			t.Fatalf("Failed to parse server response '%s' into models.Ebook: %v", string(bodyBytes), err)
		}

		// Assert the returned ebook's properties
		if actualEbook.Title != "Any Title" || actualEbook.ID != "any-id" || actualEbook.InfoProducerID != "any-producer-id" {
			t.Errorf("Returned ebook properties mismatch: got %+v", actualEbook)
		}

		// Ensure all expectations on the mock were met
		ebookSvc.AssertExpectations(t)
	})

}
