package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	handler "github.com/anglesson/simple-web-server/internal/sales/handler"
	"github.com/anglesson/simple-web-server/internal/mocks"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// MockTemplateRenderer implements template.TemplateRenderer for testing
type MockTemplateRenderer struct {
	mock.Mock
}

func (m *MockTemplateRenderer) View(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}, layout string) {
	m.Called(w, r, page, data, layout)
}

func (m *MockTemplateRenderer) ViewWithoutLayout(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}) {
	m.Called(w, r, page, data)
}

var _ salesvc.ClientService = (*mocks.MockClientService)(nil)
var _ template.TemplateRenderer = (*MockTemplateRenderer)(nil)

type ClientHandlerTestSuite struct {
	suite.Suite
	sut                  *handler.ClientHandler
	mockClientService    *mocks.MockClientService
	mockCreatorService   *mocks.MockCreatorService
	mockSessionManager   *mocks.MockSessionService
	mockTemplateRenderer *MockTemplateRenderer
}

func (suite *ClientHandlerTestSuite) SetupTest() {
	suite.mockClientService = mocks.NewMockClientService()
	suite.mockSessionManager = new(mocks.MockSessionService)
	suite.mockCreatorService = new(mocks.MockCreatorService)
	suite.mockTemplateRenderer = new(MockTemplateRenderer)

	suite.sut = handler.NewClientHandler(suite.mockClientService, suite.mockCreatorService, suite.mockSessionManager, suite.mockTemplateRenderer)
}

func (suite *ClientHandlerTestSuite) TestShouldUpdateClientSuccessfully() {
	creatorEmail := "creator@mail"
	clientID := uint(1)
	clientPublicID := "cli_abc123"

	existingClient := &salesmodel.Client{
		Model: gorm.Model{ID: clientID},
	}
	existingClient.PublicID = clientPublicID

	expectedInput := salesmodel.UpdateClientInput{
		ID:           clientID,
		Email:        "updated@mail.com",
		Phone:        "Updated Phone",
		EmailCreator: creatorEmail,
	}

	formData := strings.NewReader("cpf=Updated CPF&email=updated@mail.com&phone=Updated Phone")
	req := httptest.NewRequest(http.MethodPost, "/client/update/"+clientPublicID, formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set chi route context for URL param with PublicID
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", clientPublicID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, creatorEmail)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	suite.mockClientService.On("FindClientByPublicID", clientPublicID).Return(existingClient, nil).Once()
	suite.mockClientService.On("Update", expectedInput).Return(existingClient, nil).Once()
	suite.mockSessionManager.On("AddFlash", mock.Anything, mock.Anything, "Cliente foi atualizado!", "success").Return(nil).Once()

	suite.sut.ClientUpdateSubmit(rr, req)

	assert.Equal(suite.T(), http.StatusSeeOther, rr.Code)
	assert.Equal(suite.T(), "/client", rr.Header().Get("Location"))

	suite.mockClientService.AssertExpectations(suite.T())
	suite.mockSessionManager.AssertExpectations(suite.T())
}

func (suite *ClientHandlerTestSuite) TestClientExportCSV_UserNotFoundInContext() {
	req := httptest.NewRequest(http.MethodGet, "/client/export", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, nil)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	suite.sut.ClientExportCSV(rr, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, rr.Code)
	suite.mockClientService.AssertNotCalled(suite.T(), "ExportClients", mock.Anything)
}

func (suite *ClientHandlerTestSuite) TestClientExportCSV_ServiceError() {
	creatorEmail := "creator@mail"
	req := httptest.NewRequest(http.MethodGet, "/client/export", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, creatorEmail)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	serviceErr := errors.New("erro ao exportar clientes")
	suite.mockClientService.On("ExportClients", creatorEmail).Return((*[]salesmodel.Client)(nil), serviceErr).Once()
	suite.mockSessionManager.On("AddFlash", mock.Anything, mock.Anything, serviceErr.Error(), "error").Return(nil).Once()

	suite.sut.ClientExportCSV(rr, req)

	assert.Equal(suite.T(), http.StatusSeeOther, rr.Code)
	assert.Equal(suite.T(), "/client", rr.Header().Get("Location"))

	suite.mockClientService.AssertExpectations(suite.T())
	suite.mockSessionManager.AssertExpectations(suite.T())
}

func (suite *ClientHandlerTestSuite) TestClientExportCSV_EmptyClientList() {
	creatorEmail := "creator@mail"
	req := httptest.NewRequest(http.MethodGet, "/client/export", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, creatorEmail)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	emptyClients := &[]salesmodel.Client{}
	suite.mockClientService.On("ExportClients", creatorEmail).Return(emptyClients, nil).Once()

	suite.sut.ClientExportCSV(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Equal(suite.T(), "text/csv; charset=utf-8", rr.Header().Get("Content-Type"))
	assert.Contains(suite.T(), rr.Body.String(), "Nome,Email,Telefone,Data Nascimento")

	suite.mockClientService.AssertExpectations(suite.T())
}

func (suite *ClientHandlerTestSuite) TestClientExportCSV_Success() {
	creatorEmail := "creator@mail"
	req := httptest.NewRequest(http.MethodGet, "/client/export", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, creatorEmail)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	clients := &[]salesmodel.Client{
		{Name: "João Silva", Email: "joao@test.com", Phone: "11999999999", Birthdate: "1990-01-01"},
	}
	suite.mockClientService.On("ExportClients", creatorEmail).Return(clients, nil).Once()

	suite.sut.ClientExportCSV(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Equal(suite.T(), "text/csv; charset=utf-8", rr.Header().Get("Content-Type"))
	assert.Equal(suite.T(), "attachment; filename=clientes.csv", rr.Header().Get("Content-Disposition"))
	body := rr.Body.String()
	assert.Contains(suite.T(), body, "João Silva")
	assert.Contains(suite.T(), body, "joao@test.com")
	assert.Contains(suite.T(), body, "11999999999")

	suite.mockClientService.AssertExpectations(suite.T())
}

func TestClientHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ClientHandlerTestSuite))
}

func TestClientGetInitials(t *testing.T) {
	tests := []struct {
		name     string
		client   salesmodel.Client
		expected string
	}{
		{
			name:     "Two names",
			client:   salesmodel.Client{Name: "João Silva"},
			expected: "JS",
		},
		{
			name:     "Single name",
			client:   salesmodel.Client{Name: "João"},
			expected: "J",
		},
		{
			name:     "Three names",
			client:   salesmodel.Client{Name: "João Pedro Silva"},
			expected: "JS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.GetInitials()
			assert.Equal(t, tt.expected, result)
		})
	}
}
