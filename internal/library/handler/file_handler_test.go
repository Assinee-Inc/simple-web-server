package handler_test

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	libraryhandler "github.com/Assinee-Inc/docffy/internal/library/handler"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	libraryrepo "github.com/Assinee-Inc/docffy/internal/library/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFileService struct {
	mock.Mock
}

func (m *MockFileService) UploadFile(file *multipart.FileHeader, name, description string, creatorID uint) (*librarymodel.File, error) {
	args := m.Called(file, description, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
}

func (m *MockFileService) GetFilesByCreator(creatorID uint) ([]*librarymodel.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func (m *MockFileService) GetActiveByCreator(creatorID uint) ([]*librarymodel.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func (m *MockFileService) GetFileByID(id uint) (*librarymodel.File, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
}

func (m *MockFileService) GetFileByPublicID(publicID string) (*librarymodel.File, error) {
	args := m.Called(publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
}

func (m *MockFileService) UpdateFile(id uint, name, description string) error {
	args := m.Called(id, name, description)
	return args.Error(0)
}

func (m *MockFileService) DeleteFile(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFileService) GetFilesByType(creatorID uint, fileType string) ([]*librarymodel.File, error) {
	args := m.Called(creatorID, fileType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func (m *MockFileService) ValidateFile(file *multipart.FileHeader) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *MockFileService) GetFileType(ext string) string {
	args := m.Called(ext)
	return args.String(0)
}

func (m *MockFileService) GetFilesByCreatorPaginated(creatorID uint, query libraryrepo.FileQuery) ([]*librarymodel.File, int64, error) {
	args := m.Called(creatorID, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*librarymodel.File), args.Get(1).(int64), args.Error(2)
}

func TestNewFileHandler(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)
	assert.NotNil(t, fileHandler)
}

func TestFileHandler_FileIndexView_Unauthorized(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	req, err := http.NewRequest("GET", "/file", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	fileHandler.FileIndexView(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
}

func TestFileHandler_FileIndexView_SuccessJSON(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	creator := &accountmodel.Creator{Name: "Maria"}
	creator.ID = 1

	files := []*librarymodel.File{{Name: "ebook.pdf"}}
	mockFileService.On("GetFilesByCreatorPaginated", uint(1), mock.Anything).Return(files, int64(1), nil)

	req := httptest.NewRequest("GET", "/file?page=1&page_size=20", nil)
	req = req.WithContext(accountmw.WithCreator(req.Context(), creator))

	rr := httptest.NewRecorder()
	fileHandler.FileIndexView(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, true, body["success"])
}
