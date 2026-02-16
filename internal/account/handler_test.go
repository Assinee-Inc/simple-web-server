package account_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockManager struct {
	mock.Mock
}

func (m *MockManager) CreateAccount(account *account.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func TestPostAccount_Nil_Payload(t *testing.T) {
	req := httptest.NewRequest("POST", "/accounts", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(account.NewHandler(nil).PostAccount)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != 400 {
		t.Errorf("handler returned wrong status code: got %v want %v", status, 400)
	}

	if rr.Body.String() != "Invalid request payload\n" {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "Invalid request payload\n")
	}
}

func TestPostAccount_Missing_Required_Fields(t *testing.T) {
	payload := `{"name": "Test User", "cpf": "12345678901", "phone": "12345678901", "birth_date": 1234567890}`

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(account.NewHandler(nil).PostAccount)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, http.StatusBadRequest)
	assert.Equal(t, "Missing required fields\n", rr.Body.String(), "Expected error message for missing required fields")
}

func TestPostAccount_Manager_Error(t *testing.T) {
	payload := `{"name": "Test User", "email": "test@example.com", "cpf": "12345678901", "phone": "12345678901", "birth_date": 1234567890}`

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	mockManager := new(MockManager)
	mockManager.On("CreateAccount", mock.Anything).Return(account.ErrInvalidAccount)

	handler := http.HandlerFunc(account.NewHandler(mockManager).PostAccount)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != 400 {
		t.Errorf("handler returned wrong status code: got %v want %v", status, 400)
	}

	if rr.Body.String() != "Invalid account data\n" {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "Invalid account data\n")
	}
}

func TestPostAccount_Success(t *testing.T) {
	payload := `{"name": "Test User", "email": "test@example.com", "cpf": "12345678901", "phone": "12345678901", "birth_date": 1234567890}`
	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	mockManager := new(MockManager)
	mockManager.On("CreateAccount", mock.Anything).Return(nil)

	handler := http.HandlerFunc(account.NewHandler(mockManager).PostAccount)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response any

	bodyPayload := rr.Body.String()
	err := json.Unmarshal([]byte(bodyPayload), &response)

	assert.NoError(t, err)
	assert.Equal(t, "Test User", response.(map[string]any)["name"])
	assert.Equal(t, "test@example.com", response.(map[string]any)["email"])
	assert.Equal(t, "12345678901", response.(map[string]any)["cpf"])
	assert.Equal(t, "12345678901", response.(map[string]any)["phone"])
	assert.Equal(t, time.Unix(1234567890, 0).Format("2006-01-02T15:04:05-07:00"), response.(map[string]any)["birth_date"])
	mockManager.AssertExpectations(t)
}
