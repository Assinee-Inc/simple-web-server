package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	"mime/multipart"
	"net/url"
	"strings"

	handler "github.com/anglesson/simple-web-server/internal/handler"
	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/handler/web"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock S3Storage for testing
type MockS3Storage struct {
	mock.Mock
}

func (m *MockS3Storage) UploadFile(file *multipart.FileHeader, key string) (string, error) {
	args := m.Called(file, key)
	return args.String(0), args.Error(1)
}

func (m *MockS3Storage) DeleteFile(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockS3Storage) GenerateDownloadLink(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockS3Storage) GenerateDownloadLinkWithExpiration(key string, expirationSeconds int) string {
	args := m.Called(key, expirationSeconds)
	return args.String(0)
}

// Mock FlashMessage for testing
type MockFlashMessage struct {
	mock.Mock
}

func (m *MockFlashMessage) Success(message string) {
	m.Called(message)
}

func (m *MockFlashMessage) Error(message string) {
	m.Called(message)
}

func (m *MockFlashMessage) Warning(message string) {
	m.Called(message)
}

func (m *MockFlashMessage) Info(message string) {
	m.Called(message)
}

type EbookHandlerTestSuite struct {
	suite.Suite
	sut                  *handler.EbookHandler
	mockEbookService     *mocks.MockEbookService
	mockCreatorService   *mocks.MockCreatorService
	mockFileService      *mocks.MockFileService
	mockS3Storage        *MockS3Storage
	mockFlashMessage     *MockFlashMessage
	mockTemplateRenderer *mocks.MockTemplateRenderer
	flashFactory         web.FlashMessageFactory
}

func (suite *EbookHandlerTestSuite) SetupTest() {
	suite.mockEbookService = new(mocks.MockEbookService)
	suite.mockCreatorService = new(mocks.MockCreatorService)
	suite.mockFileService = new(mocks.MockFileService)
	suite.mockS3Storage = new(MockS3Storage)
	suite.mockFlashMessage = new(MockFlashMessage)
	suite.mockTemplateRenderer = new(mocks.MockTemplateRenderer)

	suite.flashFactory = func(w http.ResponseWriter, r *http.Request) web.FlashMessagePort {
		return suite.mockFlashMessage
	}

	suite.sut = handler.NewEbookHandler(
		suite.mockEbookService,
		suite.mockCreatorService,
		suite.mockFileService,
		suite.mockS3Storage,
		suite.flashFactory,
		suite.mockTemplateRenderer,
	)
}

// Helper function to create a request with mocked user
func (suite *EbookHandlerTestSuite) createRequestWithUser(email string) *http.Request {
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(r.Context(), middleware.UserEmailKey, email)

	// Configuração padrão do mock para FindCreatorByUserID
	creator := &models.Creator{UserID: 1}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	return r.WithContext(ctx)
}

func (suite *EbookHandlerTestSuite) TestIndexView_UserNotFound() {
	// Arrange
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	// No user email in context

	// Act
	suite.sut.IndexView(w, r)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *EbookHandlerTestSuite) TestIndexView_ServiceError() {
	// Arrange
	w := httptest.NewRecorder()
	r := suite.createRequestWithUser("test@example.com")

	suite.mockEbookService.On("ListEbooksForUser", uint(1), mock.AnythingOfType("repository.EbookQuery")).
		Return(nil, errors.New("service error"))

	// Act
	suite.sut.IndexView(w, r)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateView_UserNotFound() {
	// Arrange
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	// No user email in context

	// Act
	suite.sut.CreateView(w, r)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *EbookHandlerTestSuite) TestCreateView_CreatorNotFound() {
	// Arrange
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Adicionar usuário ao contexto
	ctx := context.WithValue(r.Context(), middleware.UserEmailKey, "test@example.com")
	r = r.WithContext(ctx)

	// Configuração do mock para FindCreatorByUserID
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).
		Return(nil, errors.New("creator not found"))

	// Configuração para GetFilesByCreatorPaginated - mesmo que não seja chamado devido ao erro anterior
	pagination := models.NewPagination(0, 20)
	query := repository.FileQuery{
		Pagination: pagination,
	}
	suite.mockFileService.On("GetFilesByCreatorPaginated", uint(1), query).
		Return([]*models.File{}, int64(0), nil)

	// Act
	suite.sut.CreateView(w, r)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	suite.mockCreatorService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestShowView_UserNotFound() {
	// Arrange
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	// No user email in context

	// Act
	suite.sut.ShowView(w, r)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *EbookHandlerTestSuite) TestShowView_EbookNotFound() {
	// Arrange
	w := httptest.NewRecorder()
	r := suite.createRequestWithUser("test@example.com")

	// Mock chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	suite.mockEbookService.On("FindByID", uint(1)).
		Return(nil, errors.New("ebook not found"))

	// Act
	suite.sut.ShowView(w, r)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_Success() {
	// Arrange
	formData := url.Values{}
	formData.Set("title", "Test Ebook")
	formData.Set("description", "Test Description")
	formData.Set("sales_page", "Test Sales Page")
	formData.Set("value", "29,90") // Use comma for Brazilian format
	formData.Add("selected_files", "1")
	formData.Add("selected_files", "2")

	req := httptest.NewRequest("POST", "/ebook/create", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add user context directly
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Mock creator service
	creator := &models.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	// Mock file service for selected files
	file1 := &models.File{}
	file1.ID = 1
	file1.CreatorID = 1 // Set CreatorID to match the creator
	file2 := &models.File{}
	file2.ID = 2
	file2.CreatorID = 1 // Set CreatorID to match the creator
	suite.mockFileService.On("GetFileByID", uint(1)).Return(file1, nil)
	suite.mockFileService.On("GetFileByID", uint(2)).Return(file2, nil)

	// Mock ebook service
	suite.mockEbookService.On("Create", mock.AnythingOfType("*models.Ebook")).Return(nil)

	// Mock para ParseMultipartForm em processDirectUploads
	req.MultipartForm = &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}

	// Act
	suite.sut.CreateSubmit(w, req)

	// Assert
	resp := w.Result()

	assert.Equal(suite.T(), http.StatusSeeOther, resp.StatusCode)

	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockFileService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_ValidationErrors() {
	// Arrange
	formData := url.Values{}
	formData.Set("title", "")                                                                                                                              // Empty title
	formData.Set("description", "A very long description that exceeds the maximum allowed length of 120 characters and should trigger a validation error") // Too long
	formData.Set("sales_page", "")                                                                                                                         // Empty sales page
	formData.Set("value", "invalid")                                                                                                                       // Invalid value
	// No files selected

	req := httptest.NewRequest("POST", "/ebook/create", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = suite.createRequestWithUser("test@example.com")

	w := httptest.NewRecorder()

	// Act
	suite.sut.CreateSubmit(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(suite.T(), http.StatusSeeOther, resp.StatusCode)

	// Check if cookies were set with errors
	cookies := resp.Cookies()
	formCookie := findCookie(cookies, "form")
	errorsCookie := findCookie(cookies, "errors")

	assert.NotNil(suite.T(), formCookie, "Form cookie should be set")
	assert.NotNil(suite.T(), errorsCookie, "Errors cookie should be set")

	// Verify errors in cookie
	errorsValue, _ := url.QueryUnescape(errorsCookie.Value)
	var savedErrors map[string]string
	err := json.Unmarshal([]byte(errorsValue), &savedErrors)
	assert.NoError(suite.T(), err)

	// Should have errors for empty title, long description, empty sales page, invalid value, and no files
	assert.Contains(suite.T(), savedErrors, "title", "Should have title error")
	assert.Contains(suite.T(), savedErrors, "description", "Should have description error")
	assert.Contains(suite.T(), savedErrors, "sales_page", "Should have sales_page error")
	assert.Contains(suite.T(), savedErrors, "value", "Should have value error")
	assert.Contains(suite.T(), savedErrors, "files", "Should have files error")
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_DescriptionTooLong() {
	// Arrange
	formData := url.Values{}
	formData.Set("title", "Valid Title")
	formData.Set("description", "This is a very long description that definitely exceeds the maximum allowed length of 120 characters and should trigger a validation error because it's way too long for the database field and contains more than 120 characters which is the limit")
	formData.Set("sales_page", "Valid Sales Page")
	formData.Set("value", "29,90")
	formData.Set("selected_files", "1")

	req := httptest.NewRequest("POST", "/ebook/create", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add user context directly
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	// Setup MockCreatorService
	creator := &models.Creator{UserID: 1}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	// Mock para ParseMultipartForm
	req.MultipartForm = &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}

	w := httptest.NewRecorder()

	// Act
	suite.sut.CreateSubmit(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(suite.T(), http.StatusSeeOther, resp.StatusCode)

	// Check if cookies were set with errors
	cookies := resp.Cookies()
	errorsCookie := findCookie(cookies, "errors")
	formCookie := findCookie(cookies, "form")

	assert.NotNil(suite.T(), errorsCookie, "Errors cookie should be set")
	assert.NotNil(suite.T(), formCookie, "Form cookie should be set")

	// Verify description error in cookie
	errorsValue, _ := url.QueryUnescape(errorsCookie.Value)
	var savedErrors map[string]string
	err := json.Unmarshal([]byte(errorsValue), &savedErrors)
	assert.NoError(suite.T(), err)

	// Verify form data in cookie
	formValue, _ := url.QueryUnescape(formCookie.Value)
	var savedForm map[string]interface{}
	err = json.Unmarshal([]byte(formValue), &savedForm)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), savedErrors, "description", "Should have description error")
	assert.Contains(suite.T(), savedErrors["description"], "120", "Error should mention 120 character limit")

	// Verify form data is preserved
	assert.Equal(suite.T(), "Valid Title", savedForm["title"])
	assert.Equal(suite.T(), "Valid Sales Page", savedForm["sales_page"])
	assert.Equal(suite.T(), 29.9, savedForm["value"])
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_SimpleValidation() {
	// Arrange - simple test with just description too long
	formData := url.Values{}
	formData.Set("title", "Valid Title")
	formData.Set("description", "This is a very long description that definitely exceeds the maximum allowed length of 120 characters and should trigger a validation error because it's way too long for the database field and contains more than 120 characters which is the limit")
	formData.Set("sales_page", "Valid Sales Page")
	formData.Set("value", "29,90")
	formData.Set("selected_files", "1")

	req := httptest.NewRequest("POST", "/ebook/create", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add user context directly
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	// Setup MockCreatorService
	creator := &models.Creator{UserID: 1}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	// Mock para ParseMultipartForm
	req.MultipartForm = &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}

	// Debug: check if form data is being read correctly
	fmt.Printf("Form data encoded: %s\n", formData.Encode())

	w := httptest.NewRecorder()

	// Act
	suite.sut.CreateSubmit(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(suite.T(), http.StatusSeeOther, resp.StatusCode)

	// Check if cookies were set with errors
	cookies := resp.Cookies()
	errorsCookie := findCookie(cookies, "errors")

	assert.NotNil(suite.T(), errorsCookie, "Errors cookie should be set")

	// Verify description error in cookie
	errorsValue, _ := url.QueryUnescape(errorsCookie.Value)
	var savedErrors map[string]string
	err := json.Unmarshal([]byte(errorsValue), &savedErrors)
	assert.NoError(suite.T(), err)

	// Debug: print all errors
	fmt.Printf("All errors: %+v\n", savedErrors)

	// Should have description error for being too long
	assert.Contains(suite.T(), savedErrors, "description", "Should have description error")
}

func (suite *EbookHandlerTestSuite) TestUpdateSubmit_DeactivateEbook() {
	// Arrange - test deactivating an ebook (status = false)
	formData := url.Values{}
	formData.Set("title", "Valid Title")
	formData.Set("description", "Valid Description")
	formData.Set("sales_page", "Valid Sales Page")
	formData.Set("value", "29,90")
	// Note: status field is NOT set, which should result in status = false

	req := httptest.NewRequest("POST", "/ebook/update/1", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	// Setup route context for chi router
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Mock ebook service - return existing ebook
	existingEbook := &models.Ebook{
		Title:       "Old Title",
		Description: "Old Description",
		SalesPage:   "Old Sales Page",
		Value:       19.90,
		Status:      true, // Currently active
		CreatorID:   1,
	}
	suite.mockEbookService.On("FindByID", uint(1)).Return(existingEbook, nil)
	suite.mockEbookService.On("Update", mock.AnythingOfType("*models.Ebook")).Return(nil)

	// Mock creator service
	creator := &models.Creator{UserID: 1}
	creator.ID = 1 // Set ID through gorm.Model
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	// Mock flash message
	suite.mockFlashMessage.On("Success", "Dados do e-book foram atualizados!").Return()

	w := httptest.NewRecorder()

	// Mock para ParseMultipartForm
	req.MultipartForm = &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}

	// Act
	suite.sut.UpdateSubmit(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(suite.T(), http.StatusSeeOther, resp.StatusCode)

	// Verify that the ebook was updated with status = false
	suite.mockEbookService.AssertExpectations(suite.T())

	// Check that Update was called with the correct ebook data
	updateCalls := suite.mockEbookService.Calls
	assert.Greater(suite.T(), len(updateCalls), 0, "Update should have been called")

	// Find the Update call and verify the ebook status
	for _, call := range updateCalls {
		if call.Method == "Update" {
			ebook := call.Arguments.Get(0).(*models.Ebook)
			assert.Equal(suite.T(), "Valid Title", ebook.Title)
			assert.Equal(suite.T(), "Valid Description", ebook.Description)
			assert.Equal(suite.T(), "Valid Sales Page", ebook.SalesPage)
			assert.Equal(suite.T(), 29.90, ebook.Value)
			assert.False(suite.T(), ebook.Status, "Ebook status should be false (deactivated)")
			break
		}
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func TestEbookHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(EbookHandlerTestSuite))
}
