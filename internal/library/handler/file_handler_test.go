package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	libraryhandler "github.com/Assinee-Inc/docffy/internal/library/handler"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	libraryrepo "github.com/Assinee-Inc/docffy/internal/library/repository"
	librarysvc "github.com/Assinee-Inc/docffy/internal/library/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFileService struct {
	mock.Mock
}

func (m *MockFileService) GenerateUploadURL(ctx context.Context, req librarysvc.PresignedUploadRequest, creatorID uint) (*librarysvc.PresignedUploadResponse, error) {
	args := m.Called(ctx, req, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarysvc.PresignedUploadResponse), args.Error(1)
}

func (m *MockFileService) CompleteUpload(ctx context.Context, req librarysvc.CompleteUploadRequest, creatorID uint) (*librarymodel.File, error) {
	args := m.Called(ctx, req, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
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

func TestFileHandler_FilePresignUpload_Unauthorized(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	reqBody := `{"file_name":"ebook.pdf","file_size":2048576,"content_type":"application/pdf"}`
	req := httptest.NewRequest("POST", "/file/upload-url", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	fileHandler.FilePresignUpload(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestFileHandler_FilePresignUpload_InvalidPayload(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	creator := &accountmodel.Creator{Name: "Lucas"}
	creator.ID = 42

	req := httptest.NewRequest("POST", "/file/upload-url", bytes.NewBufferString(`{invalid-json`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(accountmw.WithCreator(req.Context(), creator))

	rr := httptest.NewRecorder()
	fileHandler.FilePresignUpload(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "INVALID_PAYLOAD", body["code"])
}

func TestFileHandler_FilePresignUpload_ServiceError(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	creator := &accountmodel.Creator{Name: "Lucas"}
	creator.ID = 42

	reqDTO := librarysvc.PresignedUploadRequest{
		FileName:    "livro.exe",
		FileSize:    2048576,
		ContentType: "application/octet-stream",
	}

	mockFileService.On("GenerateUploadURL", mock.Anything, reqDTO, creator.ID).
		Return(nil, errors.New("Apenas arquivos PDF são permitidos ou tamanho excede 100MB"))

	reqJSON, _ := json.Marshal(reqDTO)
	req := httptest.NewRequest("POST", "/file/upload-url", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(accountmw.WithCreator(req.Context(), creator))

	rr := httptest.NewRecorder()
	fileHandler.FilePresignUpload(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "INVALID_FILE", body["code"])
}

func TestFileHandler_FilePresignUpload_Success(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	creator := &accountmodel.Creator{Name: "Lucas"}
	creator.ID = 42

	reqDTO := librarysvc.PresignedUploadRequest{
		FileName:    "apostila.pdf",
		FileSize:    2048576,
		ContentType: "application/pdf",
	}

	expectedResponse := &librarysvc.PresignedUploadResponse{
		Success:   true,
		UploadURL: "https://docffy-bucket.s3.us-east-1.amazonaws.com/files/42/apostila-123.pdf?X-Amz-...",
		Key:       "files/42/apostila-123.pdf",
		Method:    "PUT",
		Headers: map[string]string{
			"Content-Type": "application/pdf",
		},
		ExpiresIn: 900,
	}

	mockFileService.On("GenerateUploadURL", mock.Anything, reqDTO, creator.ID).
		Return(expectedResponse, nil)

	reqJSON, _ := json.Marshal(reqDTO)
	req := httptest.NewRequest("POST", "/file/upload-url", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(accountmw.WithCreator(req.Context(), creator))

	rr := httptest.NewRecorder()
	fileHandler.FilePresignUpload(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, expectedResponse.UploadURL, body["upload_url"])
	assert.Equal(t, expectedResponse.Key, body["key"])
	assert.Equal(t, "PUT", body["method"])
	assert.Equal(t, float64(900), body["expires_in"])
}

func TestFileHandler_FileCompleteUpload_Unauthorized(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	reqBody := `{"key":"files/42/doc.pdf","name":"Doc","original_name":"doc.pdf","size":1024}`
	req := httptest.NewRequest("POST", "/file/complete-upload", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	fileHandler.FileCompleteUpload(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestFileHandler_FileCompleteUpload_InvalidPayload(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	creator := &accountmodel.Creator{Name: "Lucas"}
	creator.ID = 42

	req := httptest.NewRequest("POST", "/file/complete-upload", bytes.NewBufferString(`{malformed-json`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(accountmw.WithCreator(req.Context(), creator))

	rr := httptest.NewRecorder()
	fileHandler.FileCompleteUpload(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "INVALID_PAYLOAD", body["code"])
}

func TestFileHandler_FileCompleteUpload_ForbiddenNamespace(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	creator := &accountmodel.Creator{Name: "Lucas"}
	creator.ID = 42

	reqDTO := librarysvc.CompleteUploadRequest{
		Key:          "files/99/outro-creator.pdf",
		Name:         "Guia Hack",
		OriginalName: "outro-creator.pdf",
		Size:         1024,
	}

	mockFileService.On("CompleteUpload", mock.Anything, reqDTO, creator.ID).
		Return(nil, librarysvc.ErrForbiddenNamespace)

	reqJSON, _ := json.Marshal(reqDTO)
	req := httptest.NewRequest("POST", "/file/complete-upload", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(accountmw.WithCreator(req.Context(), creator))

	rr := httptest.NewRecorder()
	fileHandler.FileCompleteUpload(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "FORBIDDEN", body["code"])
	assert.Equal(t, "A chave do arquivo não pertence ao criador autenticado", body["error"])
}

func TestFileHandler_FileCompleteUpload_Success(t *testing.T) {
	mockFileService := &MockFileService{}
	fileHandler := libraryhandler.NewFileHandler(mockFileService)

	creator := &accountmodel.Creator{Name: "Lucas"}
	creator.ID = 42

	reqDTO := librarysvc.CompleteUploadRequest{
		Key:          "files/42/guia-definitivo-123.pdf",
		Name:         "Guia Definitivo do Docffy",
		Description:  "Material complementar",
		OriginalName: "guia-definitivo.pdf",
		Size:         2048576,
		ContentType:  "application/pdf",
	}

	createdFile := librarymodel.NewFile(
		reqDTO.Name,
		reqDTO.OriginalName,
		reqDTO.Description,
		"pdf",
		reqDTO.Key,
		"https://bucket.s3.amazonaws.com/files/42/guia-definitivo-123.pdf",
		reqDTO.Size,
		creator.ID,
	)
	createdFile.PublicID = "fil_123456"

	mockFileService.On("CompleteUpload", mock.Anything, reqDTO, creator.ID).
		Return(createdFile, nil)

	reqJSON, _ := json.Marshal(reqDTO)
	req := httptest.NewRequest("POST", "/file/complete-upload", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(accountmw.WithCreator(req.Context(), creator))

	rr := httptest.NewRecorder()
	fileHandler.FileCompleteUpload(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var body map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Arquivo registrado com sucesso!", body["message"])

	fileData, ok := body["file"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, createdFile.Name, fileData["name"])
	assert.Equal(t, createdFile.PublicID, fileData["public_id"])
}
