package handler_test

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	handler "github.com/anglesson/simple-web-server/internal/handler"
	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/models"
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

type EbookHandlerTestSuite struct {
	suite.Suite
	sut                  *handler.EbookHandler
	mockEbookService     *mocks.MockEbookService
	mockCreatorService   *mocks.MockCreatorService
	mockFileService      *mocks.MockFileService
	mockS3Storage        *MockS3Storage
	mockSessionManager   *mocks.MockSessionService
	mockTemplateRenderer *mocks.MockTemplateRenderer
}

func (suite *EbookHandlerTestSuite) SetupTest() {
	suite.mockEbookService = new(mocks.MockEbookService)
	suite.mockCreatorService = new(mocks.MockCreatorService)
	suite.mockFileService = new(mocks.MockFileService)
	suite.mockS3Storage = new(MockS3Storage)
	suite.mockSessionManager = new(mocks.MockSessionService)
	suite.mockTemplateRenderer = new(mocks.MockTemplateRenderer)

	suite.sut = handler.NewEbookHandler(
		suite.mockEbookService,
		suite.mockCreatorService,
		suite.mockFileService,
		suite.mockS3Storage,
		suite.mockSessionManager,
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
	suite.mockSessionManager.On("AddFlash", mock.Anything, mock.Anything, "E-book criado com sucesso!", "success").Return(nil)

	// Mock para ParseMultipartForm em processDirectUploads
	req.MultipartForm = &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}

	// Act
	suite.sut.CreateSubmit(w, req)

	// Assert
	resp := w.Result()

	assert.Equal(suite.T(), http.StatusSeeOther, resp.StatusCode)
	assert.Equal(suite.T(), "/ebook", w.Header().Get("Location"))

	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockFileService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
	suite.mockSessionManager.AssertExpectations(suite.T())
}

func TestEbookHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(EbookHandlerTestSuite))
}
