package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	libraryhandler "github.com/Assinee-Inc/docffy/internal/library/handler"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

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
	sut                *libraryhandler.EbookHandler
	mockEbookService   *mocks.MockEbookService
	mockCreatorService *mocks.MockCreatorService
	mockFileService    *mocks.MockFileService
	mockS3Storage      *MockS3Storage
}

func (suite *EbookHandlerTestSuite) SetupTest() {
	suite.mockEbookService = new(mocks.MockEbookService)
	suite.mockCreatorService = new(mocks.MockCreatorService)
	suite.mockFileService = new(mocks.MockFileService)
	suite.mockS3Storage = new(MockS3Storage)

	suite.sut = libraryhandler.NewEbookHandler(
		suite.mockEbookService,
		suite.mockCreatorService,
		suite.mockFileService,
		suite.mockS3Storage,
	)
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_Success() {
	formData := url.Values{}
	formData.Set("title", "Test Ebook")
	formData.Set("description", "Test Description")
	formData.Set("sales_page", "Test Sales Page")
	formData.Set("value", "29,90")
	formData.Add("new_files", "fil_abc123")
	formData.Add("new_files", "fil_def456")

	req := httptest.NewRequest("POST", "/ebook/create", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	file1 := &librarymodel.File{}
	file1.ID = 1
	file1.PublicID = "fil_abc123"
	file1.CreatorID = 1
	file2 := &librarymodel.File{}
	file2.ID = 2
	file2.PublicID = "fil_def456"
	file2.CreatorID = 1
	suite.mockFileService.On("GetFileByPublicID", "fil_abc123").Return(file1, nil)
	suite.mockFileService.On("GetFileByPublicID", "fil_def456").Return(file2, nil)

	suite.mockEbookService.On("Create", mock.AnythingOfType("*model.Ebook")).Return(nil)

	req.MultipartForm = &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}

	suite.sut.CreateSubmit(w, req)

	resp := w.Result()
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	assert.Contains(suite.T(), resp.Header.Get("Content-Type"), "application/json")

	var body map[string]any
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), true, body["success"])

	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockFileService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) newRequestWithChiParams(method, path string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
}

func (suite *EbookHandlerTestSuite) TestRemoveFileFromEbook_Success() {
	file1 := &librarymodel.File{}
	file1.ID = 1
	file1.PublicID = "fil_file1pub"
	file1.CreatorID = 1
	file2 := &librarymodel.File{}
	file2.ID = 2
	file2.PublicID = "fil_file2pub"
	file2.CreatorID = 1

	ebook := &librarymodel.Ebook{}
	ebook.ID = 10
	ebook.PublicID = "ebk_ebook10pub"
	ebook.CreatorID = 1
	ebook.Files = []*librarymodel.File{file1, file2}

	creator := &accountmodel.Creator{}
	creator.ID = 1
	creator.UserID = 1

	suite.mockEbookService.On("FindByPublicID", "ebk_ebook10pub").Return(ebook, nil)
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)
	suite.mockFileService.On("GetFileByPublicID", "fil_file2pub").Return(file2, nil)
	suite.mockEbookService.On("RemoveFileAssociation", uint(10), uint(2)).Return(nil)

	req := suite.newRequestWithChiParams("POST", "/ebook/ebk_ebook10pub/remove-file/fil_file2pub", map[string]string{
		"id":     "ebk_ebook10pub",
		"fileId": "fil_file2pub",
	})
	w := httptest.NewRecorder()

	suite.sut.RemoveFileFromEbook(w, req)

	resp := w.Result()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(suite.T(), "Arquivo removido com sucesso", body["message"])

	suite.mockEbookService.AssertExpectations(suite.T())
	suite.mockCreatorService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestRemoveFileFromEbook_LastFile_ReturnsBadRequest() {
	file1 := &librarymodel.File{}
	file1.ID = 1
	file1.PublicID = "fil_file1pub"
	file1.CreatorID = 1

	ebook := &librarymodel.Ebook{}
	ebook.ID = 10
	ebook.PublicID = "ebk_ebook10pub"
	ebook.CreatorID = 1
	ebook.Files = []*librarymodel.File{file1}

	creator := &accountmodel.Creator{}
	creator.ID = 1
	creator.UserID = 1

	suite.mockEbookService.On("FindByPublicID", "ebk_ebook10pub").Return(ebook, nil)
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)
	suite.mockFileService.On("GetFileByPublicID", "fil_file1pub").Return(file1, nil)

	req := suite.newRequestWithChiParams("POST", "/ebook/ebk_ebook10pub/remove-file/fil_file1pub", map[string]string{
		"id":     "ebk_ebook10pub",
		"fileId": "fil_file1pub",
	})
	w := httptest.NewRecorder()

	suite.sut.RemoveFileFromEbook(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "pelo menos um arquivo")

	suite.mockEbookService.AssertNotCalled(suite.T(), "RemoveFileAssociation")
}

func (suite *EbookHandlerTestSuite) TestRemoveFileFromEbook_ServiceError_ReturnsInternalServerError() {
	file1 := &librarymodel.File{}
	file1.ID = 1
	file1.PublicID = "fil_file1pub"
	file2 := &librarymodel.File{}
	file2.ID = 2
	file2.PublicID = "fil_file2pub"

	ebook := &librarymodel.Ebook{}
	ebook.ID = 10
	ebook.PublicID = "ebk_ebook10pub"
	ebook.CreatorID = 1
	ebook.Files = []*librarymodel.File{file1, file2}

	creator := &accountmodel.Creator{}
	creator.ID = 1
	creator.UserID = 1

	suite.mockEbookService.On("FindByPublicID", "ebk_ebook10pub").Return(ebook, nil)
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)
	suite.mockFileService.On("GetFileByPublicID", "fil_file2pub").Return(file2, nil)
	suite.mockEbookService.On("RemoveFileAssociation", uint(10), uint(2)).Return(assert.AnError)

	req := suite.newRequestWithChiParams("POST", "/ebook/ebk_ebook10pub/remove-file/fil_file2pub", map[string]string{
		"id":     "ebk_ebook10pub",
		"fileId": "fil_file2pub",
	})
	w := httptest.NewRecorder()

	suite.sut.RemoveFileFromEbook(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) newCreateSubmitRequest(formData url.Values) *http.Request {
	req := httptest.NewRequest("POST", "/ebook/create", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)
	req.MultipartForm = &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}
	return req
}

func (suite *EbookHandlerTestSuite) baseFormData() url.Values {
	data := url.Values{}
	data.Set("title", "Título do Ebook Completo")
	data.Set("description", "Descrição curta do ebook")
	data.Set("sales_page", "Página de vendas do ebook")
	data.Set("value", "29,90")
	data.Add("new_files", "fil_abc123")
	return data
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_AuthorName_UsesFormValueWhenProvided() {
	formData := suite.baseFormData()
	formData.Set("author_name", "Escritor Fantástico")

	req := suite.newCreateSubmitRequest(formData)
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{SocialName: "Nome Social"}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	file := &librarymodel.File{}
	file.ID = 1
	file.PublicID = "fil_abc123"
	file.CreatorID = 1
	suite.mockFileService.On("GetFileByPublicID", "fil_abc123").Return(file, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.AuthorName == "Escritor Fantástico"
	})).Return(nil)

	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_AuthorName_UsesSocialNameWhenFormEmpty() {
	formData := suite.baseFormData()

	req := suite.newCreateSubmitRequest(formData)
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{Name: "João Silva Santos", SocialName: "João das Letras"}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	file := &librarymodel.File{}
	file.ID = 1
	file.PublicID = "fil_abc123"
	file.CreatorID = 1
	suite.mockFileService.On("GetFileByPublicID", "fil_abc123").Return(file, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.AuthorName == "João das Letras"
	})).Return(nil)

	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_AuthorName_UsesFirstLastNameWhenNoSocialName() {
	formData := suite.baseFormData()

	req := suite.newCreateSubmitRequest(formData)
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{Name: "João Silva Santos"}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	file := &librarymodel.File{}
	file.ID = 1
	file.PublicID = "fil_abc123"
	file.CreatorID = 1
	suite.mockFileService.On("GetFileByPublicID", "fil_abc123").Return(file, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.AuthorName == "João Santos"
	})).Return(nil)

	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) newUpdateSubmitRequest(publicID string, formData url.Values) *http.Request {
	req := httptest.NewRequest("POST", "/ebook/update/"+publicID, strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", publicID)
	req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
	req.MultipartForm = &multipart.Form{
		File: make(map[string][]*multipart.FileHeader),
	}
	return req
}

func (suite *EbookHandlerTestSuite) baseUpdateFormData() url.Values {
	data := url.Values{}
	data.Set("title", "Título do Ebook Completo")
	data.Set("description", "Descrição curta do ebook")
	data.Set("sales_page", "Página de vendas do ebook")
	data.Set("value", "29,90")
	data.Set("promotional_value", "0")
	return data
}

func (suite *EbookHandlerTestSuite) TestUpdateSubmit_AuthorName_UsesFormValueWhenProvided() {
	const publicID = "ebk_test001"

	formData := suite.baseUpdateFormData()
	formData.Set("author_name", "Escritor Fantástico")

	req := suite.newUpdateSubmitRequest(publicID, formData)
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{SocialName: "Nome Social"}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	ebook := &librarymodel.Ebook{}
	ebook.ID = 10
	ebook.PublicID = publicID
	ebook.CreatorID = 1
	suite.mockEbookService.On("FindByPublicID", publicID).Return(ebook, nil)

	suite.mockEbookService.On("Update", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.AuthorName == "Escritor Fantástico"
	})).Return(nil)

	suite.sut.UpdateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestUpdateSubmit_AuthorName_UsesSocialNameWhenFormEmpty() {
	const publicID = "ebk_test002"

	formData := suite.baseUpdateFormData()

	req := suite.newUpdateSubmitRequest(publicID, formData)
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{Name: "Maria Aparecida Ferreira", SocialName: "Mari F."}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	ebook := &librarymodel.Ebook{}
	ebook.ID = 11
	ebook.PublicID = publicID
	ebook.CreatorID = 1
	suite.mockEbookService.On("FindByPublicID", publicID).Return(ebook, nil)

	suite.mockEbookService.On("Update", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.AuthorName == "Mari F."
	})).Return(nil)

	suite.sut.UpdateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestUpdateSubmit_AuthorName_UsesFirstLastNameWhenNoSocialName() {
	const publicID = "ebk_test003"

	formData := suite.baseUpdateFormData()

	req := suite.newUpdateSubmitRequest(publicID, formData)
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{Name: "Carlos Eduardo Pereira"}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	ebook := &librarymodel.Ebook{}
	ebook.ID = 12
	ebook.PublicID = publicID
	ebook.CreatorID = 1
	suite.mockEbookService.On("FindByPublicID", publicID).Return(ebook, nil)

	suite.mockEbookService.On("Update", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.AuthorName == "Carlos Pereira"
	})).Return(nil)

	suite.sut.UpdateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_JSON_Success() {
	payload := map[string]any{
		"title":             "Ebook JSON Test",
		"description":       "Descrição via JSON",
		"sales_page":        "Texto de vendas",
		"author_name":       "Autor Teste",
		"value":             "49,90",
		"promotional_value": "29,90",
		"status":            true,
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.Title == "Ebook JSON Test" &&
			e.Description == "Descrição via JSON" &&
			e.SalesPage == "Texto de vendas" &&
			e.AuthorName == "Autor Teste" &&
			e.Value == 49.90 &&
			e.PromotionalValue == 29.90 &&
			e.Status == true
	})).Return(nil)

	w := httptest.NewRecorder()
	suite.sut.CreateSubmit(w, req)

	resp := w.Result()
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	assert.Contains(suite.T(), resp.Header.Get("Content-Type"), "application/json")

	var respBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), true, respBody["success"])

	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_JSON_NumericValues() {
	payload := map[string]any{
		"title":             "Ebook Numeric Values",
		"description":       "Descrição com valores numéricos",
		"sales_page":        "Texto",
		"author_name":       "Autor Numérico",
		"value":             49.9,
		"promotional_value": 29.9,
		"status":            true,
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.Title == "Ebook Numeric Values" &&
			e.Value == 49.9 &&
			e.PromotionalValue == 29.9 &&
			e.Status == true
	})).Return(nil)

	w := httptest.NewRecorder()
	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	var respBody map[string]any
	err = json.NewDecoder(w.Body).Decode(&respBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), true, respBody["success"])

	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_JSON_EmptyTitle_ReturnsValidationError() {
	payload := map[string]any{
		"title":       "",
		"description": "Sem título",
		"value":       "49,90",
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var respBody map[string]any
	err = json.NewDecoder(w.Body).Decode(&respBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Dados inválidos para criação do e-book", respBody["error"])
}

func (suite *EbookHandlerTestSuite) TestUpdateSubmit_JSON_Success() {
	const publicID = "ebk_update_json"
	payload := map[string]any{
		"title":             "Ebook Atualizado via JSON",
		"description":       "Nova descrição",
		"sales_page":        "Nova página",
		"author_name":       "Novo Autor",
		"value":             "59,90",
		"promotional_value": "39,90",
		"status":            true,
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/update/"+publicID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", publicID)
	req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	ebook := &librarymodel.Ebook{}
	ebook.ID = 20
	ebook.PublicID = publicID
	ebook.CreatorID = 1
	suite.mockEbookService.On("FindByPublicID", publicID).Return(ebook, nil)

	suite.mockEbookService.On("Update", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.Title == "Ebook Atualizado via JSON" &&
			e.Description == "Nova descrição" &&
			e.SalesPage == "Nova página" &&
			e.AuthorName == "Novo Autor" &&
			e.Value == 59.90 &&
			e.PromotionalValue == 39.90 &&
			e.Status == true
	})).Return(nil)

	w := httptest.NewRecorder()
	suite.sut.UpdateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var respBody map[string]any
	err = json.NewDecoder(w.Body).Decode(&respBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), true, respBody["success"])

	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_JSON_WithFilesAliases_NewFiles() {
	payload := map[string]any{
		"title":       "Ebook with NewFiles",
		"description": "Desc",
		"sales_page":  "Sales",
		"value":       "19.90",
		"new_files":   []string{"fil_new123"},
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	file := &librarymodel.File{}
	file.ID = 10
	file.PublicID = "fil_new123"
	file.CreatorID = 1
	suite.mockFileService.On("GetFileByPublicID", "fil_new123").Return(file, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return len(e.Files) == 1 && e.Files[0].PublicID == "fil_new123"
	})).Return(nil)

	w := httptest.NewRecorder()
	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockFileService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_JSON_WithFilesAlias_PublicID() {
	payload := map[string]any{
		"title":       "Ebook with Files Alias",
		"description": "Desc",
		"sales_page":  "Sales",
		"value":       "19.90",
		"files":       []string{"fil_alias456"},
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	file := &librarymodel.File{}
	file.ID = 20
	file.PublicID = "fil_alias456"
	file.CreatorID = 1
	suite.mockFileService.On("GetFileByPublicID", "fil_alias456").Return(file, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return len(e.Files) == 1 && e.Files[0].PublicID == "fil_alias456"
	})).Return(nil)

	w := httptest.NewRecorder()
	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockFileService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_JSON_WithFileIDsAlias_NumericID_FallbackToGetFileByID() {
	payload := map[string]any{
		"title":       "Ebook with Numeric File ID",
		"description": "Desc",
		"sales_page":  "Sales",
		"value":       "19.90",
		"file_ids":    []int{42},
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	// GetFileByPublicID("42") fails (returns nil, error)
	suite.mockFileService.On("GetFileByPublicID", "42").Return(nil, assert.AnError)

	// Fallback to GetFileByID(42) succeeds
	file := &librarymodel.File{}
	file.ID = 42
	file.PublicID = "fil_real42"
	file.CreatorID = 1
	suite.mockFileService.On("GetFileByID", uint(42)).Return(file, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return len(e.Files) == 1 && e.Files[0].ID == 42
	})).Return(nil)

	w := httptest.NewRecorder()
	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockFileService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestCreateSubmit_JSON_WithFiles_OwnershipMismatchIgnored() {
	payload := map[string]any{
		"title":       "Ebook Ownership Test",
		"description": "Desc",
		"sales_page":  "Sales",
		"value":       "19.90",
		"files":       []string{"fil_notmine"},
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	req = req.WithContext(ctx)

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	file := &librarymodel.File{}
	file.ID = 30
	file.PublicID = "fil_notmine"
	file.CreatorID = 999 // Different creator!
	suite.mockFileService.On("GetFileByPublicID", "fil_notmine").Return(file, nil)

	suite.mockEbookService.On("Create", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return len(e.Files) == 0
	})).Return(nil)

	w := httptest.NewRecorder()
	suite.sut.CreateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockFileService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func (suite *EbookHandlerTestSuite) TestUpdateSubmit_JSON_WithNewFilesAssociated() {
	const publicID = "ebk_update_files"
	payload := map[string]any{
		"title":       "Ebook Atualizado com Arquivos",
		"description": "Desc Atualizada",
		"sales_page":  "Sales Atualizada",
		"value":       "39.90",
		"files":       []any{"fil_pub99", 55},
	}
	bodyBytes, err := json.Marshal(payload)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", "/ebook/update/"+publicID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), authmw.UserIDContextKey, uint(1))
	ctx = context.WithValue(ctx, authmw.UserEmailKey, "test@example.com")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", publicID)
	req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

	creator := &accountmodel.Creator{}
	creator.ID = 1
	suite.mockCreatorService.On("FindCreatorByUserID", uint(1)).Return(creator, nil)

	ebook := &librarymodel.Ebook{}
	ebook.ID = 70
	ebook.PublicID = publicID
	ebook.CreatorID = 1
	suite.mockEbookService.On("FindByPublicID", publicID).Return(ebook, nil)

	file1 := &librarymodel.File{}
	file1.ID = 99
	file1.PublicID = "fil_pub99"
	file1.CreatorID = 1
	suite.mockFileService.On("GetFileByPublicID", "fil_pub99").Return(file1, nil)

	suite.mockFileService.On("GetFileByPublicID", "55").Return(nil, assert.AnError)
	file2 := &librarymodel.File{}
	file2.ID = 55
	file2.PublicID = "fil_real55"
	file2.CreatorID = 1
	suite.mockFileService.On("GetFileByID", uint(55)).Return(file2, nil)

	suite.mockEbookService.On("AppendFiles", uint(70), mock.MatchedBy(func(files []*librarymodel.File) bool {
		return len(files) == 2 && files[0].ID == 99 && files[1].ID == 55
	})).Return(nil)

	suite.mockEbookService.On("Update", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.Title == "Ebook Atualizado com Arquivos" && e.Value == 39.90
	})).Return(nil)

	w := httptest.NewRecorder()
	suite.sut.UpdateSubmit(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.mockCreatorService.AssertExpectations(suite.T())
	suite.mockFileService.AssertExpectations(suite.T())
	suite.mockEbookService.AssertExpectations(suite.T())
}

func TestEbookHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(EbookHandlerTestSuite))
}
