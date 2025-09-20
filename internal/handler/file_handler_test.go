package handler_test

import (
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "github.com/anglesson/simple-web-server/internal/handler"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock FileService
type MockFileService struct {
	mock.Mock
}

func (m *MockFileService) UploadFile(file *multipart.FileHeader, description string, creatorID uint) (*models.File, error) {
	args := m.Called(file, description, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

func (m *MockFileService) GetFilesByCreator(creatorID uint) ([]*models.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

func (m *MockFileService) GetActiveByCreator(creatorID uint) ([]*models.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

func (m *MockFileService) GetFileByID(id uint) (*models.File, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

func (m *MockFileService) UpdateFile(id uint, name, description string) error {
	args := m.Called(id, name, description)
	return args.Error(0)
}

func (m *MockFileService) DeleteFile(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFileService) GetFilesByType(creatorID uint, fileType string) ([]*models.File, error) {
	args := m.Called(creatorID, fileType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

func (m *MockFileService) ValidateFile(file *multipart.FileHeader) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *MockFileService) GetFileType(ext string) string {
	args := m.Called(ext)
	return args.String(0)
}

func (m *MockFileService) GetFilesByCreatorPaginated(creatorID uint, query repository.FileQuery) ([]*models.File, int64, error) {
	args := m.Called(creatorID, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.File), args.Get(1).(int64), args.Error(2)
}

func TestNewFileHandler(t *testing.T) {
	// Arrange
	mockFileService := &MockFileService{}
	mockSessionManager := &mocks.MockSessionService{}
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}

	// Act
	fileHandler := handler.NewFileHandler(mockFileService, mockSessionManager, mockTemplateRenderer)

	// Assert
	assert.NotNil(t, fileHandler)
}

func TestFileHandler_FileIndexView(t *testing.T) {
	// Arrange
	mockFileService := &MockFileService{}
	mockSessionManager := &mocks.MockSessionService{}
	mockTemplateRenderer := &mocks.MockTemplateRenderer{}
	fileHandler := handler.NewFileHandler(mockFileService, mockSessionManager, mockTemplateRenderer)

	req, err := http.NewRequest("GET", "/file", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Act
	fileHandler.FileIndexView(rr, req)

	// Assert
	assert.Equal(t, http.StatusSeeOther, rr.Code) // Deve redirecionar para login
}
