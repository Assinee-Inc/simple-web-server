package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestSalesPageHandler_Creation(t *testing.T) {
	mockEbookService := new(mocks.MockEbookService)
	mockCreatorService := new(mocks.MockCreatorService)

	handler := NewSalesPageHandler(mockEbookService, mockCreatorService)

	assert.NotNil(t, handler, "Handler deve ser criado")
}

// ─── SalesPageView ──────────────────────────────────────────────────────────

type SalesPageViewSuite struct {
	suite.Suite
	handler            *SalesPageHandler
	mockEbookService   *mocks.MockEbookService
	mockCreatorService *mocks.MockCreatorService
}

func (s *SalesPageViewSuite) SetupTest() {
	s.mockEbookService = new(mocks.MockEbookService)
	s.mockCreatorService = new(mocks.MockCreatorService)
	s.handler = NewSalesPageHandler(s.mockEbookService, s.mockCreatorService)
}

func (s *SalesPageViewSuite) reqWithID(id string) *http.Request {
	req := httptest.NewRequest("GET", "/sales/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func (s *SalesPageViewSuite) TestMissingID_Returns404() {
	req := httptest.NewRequest("GET", "/sales/", nil)
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

	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_abc"))

	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Contains(s.T(), w.Header().Get("Content-Type"), "application/json")
	var body map[string]any
	err := json.NewDecoder(w.Body).Decode(&body)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), true, body["success"])
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

	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_xyz"))

	assert.Equal(s.T(), http.StatusOK, w.Code)
}

func (s *SalesPageViewSuite) TestFirstVisit_IncrementsViewsAndSetsCookie() {
	creator := &accountmodel.Creator{Name: "Autora"}
	creator.ID = 1
	ebook := &librarymodel.Ebook{Status: true, AuthorName: "Autora", CreatorID: 1}
	ebook.ID = 10

	s.mockEbookService.On("FindByPublicID", "ebk_new").Return(ebook, nil)
	s.mockCreatorService.On("FindByID", uint(1)).Return(creator, nil)
	s.mockEbookService.On("Update", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, s.reqWithID("ebk_new"))

	assert.Equal(s.T(), http.StatusOK, w.Code)
	s.mockEbookService.AssertCalled(s.T(), "Update", mock.Anything)

	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "ebook_view_ebk_new" {
			found = true
			assert.Equal(s.T(), "1", c.Value)
			assert.Equal(s.T(), 86400, c.MaxAge)
			assert.True(s.T(), c.HttpOnly)
		}
	}
	assert.True(s.T(), found, "cookie de deduplicação deve estar presente na resposta")
}

func (s *SalesPageViewSuite) TestRepeatVisit_CookiePresentSkipsIncrement() {
	creator := &accountmodel.Creator{Name: "Autora"}
	creator.ID = 1
	ebook := &librarymodel.Ebook{Status: true, AuthorName: "Autora", CreatorID: 1}
	ebook.ID = 10

	s.mockEbookService.On("FindByPublicID", "ebk_repeat").Return(ebook, nil)
	s.mockCreatorService.On("FindByID", uint(1)).Return(creator, nil)

	req := s.reqWithID("ebk_repeat")
	req.AddCookie(&http.Cookie{Name: "ebook_view_ebk_repeat", Value: "1"})

	w := httptest.NewRecorder()
	s.handler.SalesPageView(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	s.mockEbookService.AssertNotCalled(s.T(), "Update", mock.Anything)
}

func TestSalesPageViewSuite(t *testing.T) {
	suite.Run(t, new(SalesPageViewSuite))
}

// ─── SalesPagePreviewView ────────────────────────────────────────────────────

type SalesPagePreviewSuite struct {
	suite.Suite
	handler            *SalesPageHandler
	mockEbookService   *mocks.MockEbookService
	mockCreatorService *mocks.MockCreatorService
	testUser           *authmodel.User
}

func (s *SalesPagePreviewSuite) SetupTest() {
	s.mockEbookService = new(mocks.MockEbookService)
	s.mockCreatorService = new(mocks.MockCreatorService)
	s.handler = NewSalesPageHandler(s.mockEbookService, s.mockCreatorService)

	user := authmodel.NewGoogleUser("Preview Test", "preview_test@sptest.com", "gid_prev")
	user.ID = 1
	s.testUser = user
}

func (s *SalesPagePreviewSuite) reqWithAuth(publicID string) *http.Request {
	req := httptest.NewRequest("GET", "/ebook/preview/"+publicID, nil)
	ctx := authmw.WithAuthenticatedUser(req.Context(), s.testUser)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", publicID)
	return req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
}

func (s *SalesPagePreviewSuite) reqWithAuthNoID() *http.Request {
	req := httptest.NewRequest("GET", "/ebook/preview/", nil)
	ctx := authmw.WithAuthenticatedUser(req.Context(), s.testUser)
	rctx := chi.NewRouteContext()
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
	creator.ID = 50
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

	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, s.reqWithAuth("ebk_prev"))
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var body map[string]any
	err := json.NewDecoder(w.Body).Decode(&body)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), true, body["is_preview"])
	ebookMap, ok := body["ebook"].(map[string]any)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), "Mari F.", ebookMap["author_name"])
}

func (s *SalesPagePreviewSuite) TestAuthorNameSet_KeepsExistingValue() {
	creator := &accountmodel.Creator{Name: "Carlos Pereira", SocialName: "Carlão"}
	creator.ID = 6
	ebook := &librarymodel.Ebook{AuthorName: "Escritor Fantástico", CreatorID: 6}
	ebook.ID = 3

	s.mockEbookService.On("FindByPublicID", "ebk_prev2").Return(ebook, nil)
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)

	w := httptest.NewRecorder()
	s.handler.SalesPagePreviewView(w, s.reqWithAuth("ebk_prev2"))
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var body map[string]any
	err := json.NewDecoder(w.Body).Decode(&body)
	assert.NoError(s.T(), err)
	ebookMap, ok := body["ebook"].(map[string]any)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), "Escritor Fantástico", ebookMap["author_name"])
}

func TestSalesPagePreviewSuite(t *testing.T) {
	suite.Run(t, new(SalesPagePreviewSuite))
}
