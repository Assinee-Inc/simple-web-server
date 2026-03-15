package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Teste básico da criação do handler
func TestSalesPageHandler_Creation(t *testing.T) {
	// Setup
	mockEbookService := new(mocks.MockEbookService)
	mockCreatorService := new(mocks.MockCreatorService)
	mockTemplateRenderer := new(mocks.MockTemplateRenderer)

	// Act
	handler := NewSalesPageHandler(mockEbookService, mockCreatorService, mockTemplateRenderer)

	// Assert
	assert.NotNil(t, handler, "Handler deve ser criado")
}

// ─── SalesPageView ──────────────────────────────────────────────────────────

type SalesPageViewSuite struct {
	suite.Suite
	handler              *SalesPageHandler
	mockEbookService     *mocks.MockEbookService
	mockCreatorService   *mocks.MockCreatorService
	mockTemplateRenderer *mocks.MockTemplateRenderer
}

func (s *SalesPageViewSuite) SetupTest() {
	s.mockEbookService = new(mocks.MockEbookService)
	s.mockCreatorService = new(mocks.MockCreatorService)
	s.mockTemplateRenderer = new(mocks.MockTemplateRenderer)
	s.handler = NewSalesPageHandler(s.mockEbookService, s.mockCreatorService, s.mockTemplateRenderer)
}

func (s *SalesPageViewSuite) reqWithID(id string) *http.Request {
	req := httptest.NewRequest("GET", "/p/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func (s *SalesPageViewSuite) TestMissingID_Returns404() {
	req := httptest.NewRequest("GET", "/p/", nil)
	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, req)
	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

func (s *SalesPageViewSuite) TestEbookNotFound_Returns404() {
	s.mockEbookService.On("FindByPublicID", "ebk_abc").Return(nil, errors.New("not found"))
	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_abc"))
	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

func (s *SalesPageViewSuite) TestEbookInactive_Returns404() {
	ebook := &librarymodel.Ebook{Status: false}
	s.mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_abc"))
	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

func (s *SalesPageViewSuite) TestCreatorNotFound_Returns500() {
	ebook := &librarymodel.Ebook{Status: true}
	ebook.ID = 1
	s.mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	s.mockCreatorService.On("FindByID", uint(0)).Return(nil, errors.New("not found"))
	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_abc"))
	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}

func (s *SalesPageViewSuite) TestAuthorNameEmpty_UsesCreatorDisplayName() {
	creator := &accountmodel.Creator{Name: "João Silva Santos", SocialName: "João das Letras"}
	creator.ID = 7
	ebook := &librarymodel.Ebook{Status: true, AuthorName: "", CreatorID: 7}
	ebook.ID = 1

	s.mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	s.mockCreatorService.On("FindByID", uint(7)).Return(creator, nil)
	s.mockEbookService.On("Update", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.AuthorName == "João das Letras"
	})).Return(nil)
	s.mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "purchase/sales-page", mock.Anything, "guest")

	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_abc"))

	assert.Equal(s.T(), http.StatusOK, w.Code)
	s.mockEbookService.AssertExpectations(s.T())
}

func (s *SalesPageViewSuite) TestAuthorNameSet_KeepsExistingValue() {
	creator := &accountmodel.Creator{Name: "João Silva Santos", SocialName: "João das Letras"}
	creator.ID = 7
	ebook := &librarymodel.Ebook{Status: true, AuthorName: "Escritor Fantástico", CreatorID: 7}
	ebook.ID = 1

	s.mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	s.mockCreatorService.On("FindByID", uint(7)).Return(creator, nil)
	s.mockEbookService.On("Update", mock.MatchedBy(func(e *librarymodel.Ebook) bool {
		return e.AuthorName == "Escritor Fantástico"
	})).Return(nil)
	s.mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "purchase/sales-page", mock.Anything, "guest")

	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_abc"))

	assert.Equal(s.T(), http.StatusOK, w.Code)
	s.mockEbookService.AssertExpectations(s.T())
}

func (s *SalesPageViewSuite) TestUpdateError_LogsAndStillRenders() {
	creator := &accountmodel.Creator{Name: "Ana Lima"}
	creator.ID = 3
	ebook := &librarymodel.Ebook{Status: true, AuthorName: "Ana Lima", CreatorID: 3}
	ebook.ID = 5

	s.mockEbookService.On("FindByPublicID", "ebk_xyz").Return(ebook, nil)
	s.mockCreatorService.On("FindByID", uint(3)).Return(creator, nil)
	s.mockEbookService.On("Update", mock.Anything).Return(errors.New("db error"))
	s.mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "purchase/sales-page", mock.Anything, "guest")

	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_xyz"))

	assert.Equal(s.T(), http.StatusOK, w.Code)
}

func TestSalesPageViewSuite(t *testing.T) {
	suite.Run(t, new(SalesPageViewSuite))
}

// ─── SalesPagePreviewView ────────────────────────────────────────────────────

type SalesPagePreviewSuite struct {
	suite.Suite
	handler              *SalesPageHandler
	mockEbookService     *mocks.MockEbookService
	mockCreatorService   *mocks.MockCreatorService
	mockTemplateRenderer *mocks.MockTemplateRenderer
	testUser             *authmodel.User
}

func (s *SalesPagePreviewSuite) SetupSuite() {
	database.Connect()
	userRepo := authrepo.NewGormUserRepository(database.DB)
	database.DB.Exec("DELETE FROM users WHERE email = 'preview_test@sptest.com'")

	user := authmodel.NewUser("Preview Test", "hashed", "preview_test@sptest.com")
	err := userRepo.Create(user)
	s.Require().NoError(err)
	database.DB.Exec("UPDATE users SET email_verified_at = datetime('now') WHERE email = 'preview_test@sptest.com'")

	s.testUser = userRepo.FindByEmail("preview_test@sptest.com")
	s.Require().NotNil(s.testUser)
}

func (s *SalesPagePreviewSuite) SetupTest() {
	s.mockEbookService = new(mocks.MockEbookService)
	s.mockCreatorService = new(mocks.MockCreatorService)
	s.mockTemplateRenderer = new(mocks.MockTemplateRenderer)
	s.handler = NewSalesPageHandler(s.mockEbookService, s.mockCreatorService, s.mockTemplateRenderer)
}

func (s *SalesPagePreviewSuite) reqWithAuth(publicID string) *http.Request {
	req := httptest.NewRequest("GET", "/ebook/preview/"+publicID, nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, s.testUser.Email)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", publicID)
	return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
}

func (s *SalesPagePreviewSuite) reqWithAuthNoID() *http.Request {
	req := httptest.NewRequest("GET", "/ebook/preview/", nil)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, s.testUser.Email)
	rctx := chi.NewRouteContext()
	// "id" param not added → URLParam returns ""
	return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
}

func (s *SalesPagePreviewSuite) TestUnauthorized_NoContext_Returns401() {
	req := httptest.NewRequest("GET", "/ebook/preview/ebk_abc", nil)
	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, req)
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *SalesPagePreviewSuite) TestMissingPublicID_Returns400() {
	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, s.reqWithAuthNoID())
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *SalesPagePreviewSuite) TestEbookNotFound_Returns404() {
	s.mockEbookService.On("FindByPublicID", "ebk_notfound").Return(nil, errors.New("not found"))
	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, s.reqWithAuth("ebk_notfound"))
	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

func (s *SalesPagePreviewSuite) TestCreatorNotFound_Returns401() {
	ebook := &librarymodel.Ebook{}
	ebook.ID = 1
	ebook.CreatorID = 99
	s.mockEbookService.On("FindByPublicID", "ebk_ok").Return(ebook, nil)
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, s.reqWithAuth("ebk_ok"))
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *SalesPagePreviewSuite) TestCreatorMismatch_Returns401() {
	ebook := &librarymodel.Ebook{}
	ebook.ID = 1
	ebook.CreatorID = 99

	creator := &accountmodel.Creator{}
	creator.ID = 50 // differs from ebook.CreatorID
	s.mockEbookService.On("FindByPublicID", "ebk_ok").Return(ebook, nil)
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)

	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, s.reqWithAuth("ebk_ok"))
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *SalesPagePreviewSuite) TestAuthorNameEmpty_UsesCreatorDisplayName() {
	creator := &accountmodel.Creator{Name: "Maria Aparecida Ferreira", SocialName: "Mari F."}
	creator.ID = 5
	ebook := &librarymodel.Ebook{AuthorName: "", CreatorID: 5}
	ebook.ID = 2

	s.mockEbookService.On("FindByPublicID", "ebk_prev").Return(ebook, nil)
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)
	s.mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "purchase/sales-page",
		mock.MatchedBy(func(d map[string]any) bool {
			e, ok := d["Ebook"].(*librarymodel.Ebook)
			return ok && e.AuthorName == "Mari F."
		}), "guest")

	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, s.reqWithAuth("ebk_prev"))
	assert.Equal(s.T(), http.StatusOK, w.Code)
	s.mockTemplateRenderer.AssertExpectations(s.T())
}

func (s *SalesPagePreviewSuite) TestAuthorNameSet_KeepsExistingValue() {
	creator := &accountmodel.Creator{Name: "Carlos Pereira", SocialName: "Carlão"}
	creator.ID = 6
	ebook := &librarymodel.Ebook{AuthorName: "Escritor Fantástico", CreatorID: 6}
	ebook.ID = 3

	s.mockEbookService.On("FindByPublicID", "ebk_prev2").Return(ebook, nil)
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)
	s.mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "purchase/sales-page",
		mock.MatchedBy(func(d map[string]any) bool {
			e, ok := d["Ebook"].(*librarymodel.Ebook)
			return ok && e.AuthorName == "Escritor Fantástico"
		}), "guest")

	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, s.reqWithAuth("ebk_prev2"))
	assert.Equal(s.T(), http.StatusOK, w.Code)
	s.mockTemplateRenderer.AssertExpectations(s.T())
}

func TestSalesPagePreviewSuite(t *testing.T) {
	suite.Run(t, new(SalesPagePreviewSuite))
}
