package handler_test

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	libraryhandler "github.com/anglesson/simple-web-server/internal/library/handler"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock S3Storage for testing
type MockS3Storage struct {
	mock.Mock
}

func (m *MockS3Storage) UploadFile(file *multipart.FileHeader, key, cacheControl string) (string, error) {
	args := m.Called(file, key, cacheControl)
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

func (m *MockS3Storage) GeneratePreviewLinkWithExpiration(key, contentType string, expirationSeconds int) string {
	args := m.Called(key, contentType, expirationSeconds)
	return args.String(0)
}

type EbookHandlerTestSuite struct {
	suite.Suite
	sut                  *libraryhandler.EbookHandler
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

	suite.sut = libraryhandler.NewEbookHandler(
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
	ctx := context.WithValue(r.Context(), authmw.UserEmailKey, email)

	// Configuração padrão do mock para FindCreatorByUserID
	creator := &accountmodel.Creator{UserID: 1}
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
	formData.Add("new_files", "1")
	formData.Add("new_files", "2")

	req := httptest.NewRequest("POST", "/ebook/create", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add user context directly
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Mock creator service
	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	// Mock file service for selected files
	file1 := &librarymodel.File{}
	file1.ID = 1
	file1.CreatorID = 1 // Set CreatorID to match the creator
	file2 := &librarymodel.File{}
	file2.ID = 2
	file2.CreatorID = 1 // Set CreatorID to match the creator
	suite.mockFileService.On("GetFileByID", uint(1)).Return(file1, nil)
	suite.mockFileService.On("GetFileByID", uint(2)).Return(file2, nil)

	// Mock ebook service
	suite.mockEbookService.On("Create", mock.AnythingOfType("*model.Ebook")).Return(nil)
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

func (suite *EbookHandlerTestSuite) newRequestWithChiParams(method, path string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, "test@example.com")
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
}

func (suite *EbookHandlerTestSuite) TestRemoveFileFromEbook_Success() {
	// Ebook com 2 arquivos — remoção deve funcionar
	file1 := &librarymodel.File{}
	file1.ID = 1
	file1.CreatorID = 1
	file2 := &librarymodel.File{}
	file2.ID = 2
	file2.CreatorID = 1

	ebook := &librarymodel.Ebook{}
	ebook.ID = 10
	ebook.CreatorID = 1
	ebook.Files = []*librarymodel.File{file1, file2}

	creator := &accountmodel.Creator{}
	creator.ID = 1
	creator.UserID = 1

	suite.mockEbookService.On("FindByID", uint(10)).Return(ebook, nil)
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)
	suite.mockEbookService.On("RemoveFileAssociation", uint(10), uint(2)).Return(nil)

	req := suite.newRequestWithChiParams("POST", "/ebook/10/remove-file/2", map[string]string{
		"id":     "10",
		"fileId": "2",
	})
	w := httptest.NewRecorder()

	suite.sut.RemoveFileFromEbook(w, req)

	resp := w.Result()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(suite.T(), "Arquivo removido com sucesso", body["message"])

	suite.mockEbookService.AssertExpectations(suite.T())
	suite.mockCreatorService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestRemoveFileFromEbook_LastFile_ReturnsBadRequest() {
	// Ebook com apenas 1 arquivo — remoção deve ser rejeitada
	file1 := &librarymodel.File{}
	file1.ID = 1
	file1.CreatorID = 1

	ebook := &librarymodel.Ebook{}
	ebook.ID = 10
	ebook.CreatorID = 1
	ebook.Files = []*librarymodel.File{file1}

	creator := &accountmodel.Creator{}
	creator.ID = 1
	creator.UserID = 1

	suite.mockEbookService.On("FindByID", uint(10)).Return(ebook, nil)
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	req := suite.newRequestWithChiParams("POST", "/ebook/10/remove-file/1", map[string]string{
		"id":     "10",
		"fileId": "1",
	})
	w := httptest.NewRecorder()

	suite.sut.RemoveFileFromEbook(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "pelo menos um arquivo")

	suite.mockEbookService.AssertNotCalled(suite.T(), "RemoveFileAssociation")
}

func (suite *EbookHandlerTestSuite) TestRemoveFileFromEbook_ServiceError_ReturnsInternalServerError() {
	// RemoveFileAssociation retorna erro → deve responder 500
	file1 := &librarymodel.File{}
	file1.ID = 1
	file2 := &librarymodel.File{}
	file2.ID = 2

	ebook := &librarymodel.Ebook{}
	ebook.ID = 10
	ebook.CreatorID = 1
	ebook.Files = []*librarymodel.File{file1, file2}

	creator := &accountmodel.Creator{}
	creator.ID = 1
	creator.UserID = 1

	suite.mockEbookService.On("FindByID", uint(10)).Return(ebook, nil)
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)
	suite.mockEbookService.On("RemoveFileAssociation", uint(10), uint(2)).Return(assert.AnError)

	req := suite.newRequestWithChiParams("POST", "/ebook/10/remove-file/2", map[string]string{
		"id":     "10",
		"fileId": "2",
	})
	w := httptest.NewRecorder()

	suite.sut.RemoveFileFromEbook(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	suite.mockEbookService.AssertExpectations(suite.T())
}

func TestEbookHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(EbookHandlerTestSuite))
}
