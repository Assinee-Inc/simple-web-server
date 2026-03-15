package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	accounthandler "github.com/anglesson/simple-web-server/internal/account/handler"
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// errReader é um io.Reader que sempre retorna erro, usado para forçar ParseForm a falhar.
type errReader struct{}

func (e errReader) Read(p []byte) (int, error) { return 0, errors.New("forced read error") }

// ─── Suite ───────────────────────────────────────────────────────────────────

type SettingsHandlerSuite struct {
	suite.Suite
	handler              *accounthandler.SettingsHandler
	mockSessionService   *mocks.MockSessionService
	mockCreatorService   *mocks.MockCreatorService
	mockTemplateRenderer *mocks.MockTemplateRenderer
	testUser             *authmodel.User
}

func (s *SettingsHandlerSuite) SetupSuite() {
	database.Connect()
	userRepo := authrepo.NewGormUserRepository(database.DB)
	database.DB.Exec("DELETE FROM users WHERE email = 'settings_test@shtest.com'")

	user := authmodel.NewUser("Settings Test", "hashed", "settings_test@shtest.com")
	err := userRepo.Create(user)
	s.Require().NoError(err)
	database.DB.Exec("UPDATE users SET email_verified_at = datetime('now') WHERE email = 'settings_test@shtest.com'")

	s.testUser = userRepo.FindByEmail("settings_test@shtest.com")
	s.Require().NotNil(s.testUser)
}

func (s *SettingsHandlerSuite) SetupTest() {
	s.mockSessionService = new(mocks.MockSessionService)
	s.mockCreatorService = new(mocks.MockCreatorService)
	s.mockTemplateRenderer = new(mocks.MockTemplateRenderer)
	s.handler = accounthandler.NewSettingsHandler(s.mockSessionService, s.mockCreatorService, s.mockTemplateRenderer)
}

func (s *SettingsHandlerSuite) reqWithAuth(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	ctx := context.WithValue(req.Context(), authmw.UserEmailKey, s.testUser.Email)
	return req.WithContext(ctx)
}

// ─── SettingsView ─────────────────────────────────────────────────────────────

func (s *SettingsHandlerSuite) TestSettingsView_NoAuth_RedirectsToLogin() {
	req := httptest.NewRequest("GET", "/settings", nil)
	// UserEmailKey not in context → Auth returns nil
	w := httptest.NewRecorder()
	s.handler.SettingsView(w, req)
	assert.Equal(s.T(), http.StatusSeeOther, w.Code)
	assert.Equal(s.T(), "/login", w.Header().Get("Location"))
}

func (s *SettingsHandlerSuite) TestSettingsView_CreatorNotFound_StillRenders() {
	req := s.reqWithAuth("GET", "/settings", nil)
	w := httptest.NewRecorder()

	s.mockSessionService.On("RegenerateCSRFToken", mock.Anything, mock.Anything).Return("csrf-tok", nil)
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(nil, errors.New("not found"))
	s.mockSessionService.On("GetFlashes", mock.Anything, mock.Anything, "success").Return(nil)
	s.mockSessionService.On("GetFlashes", mock.Anything, mock.Anything, "error").Return(nil)
	s.mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "settings", mock.Anything, "admin-daisy")

	s.handler.SettingsView(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	s.mockTemplateRenderer.AssertExpectations(s.T())
}

func (s *SettingsHandlerSuite) TestSettingsView_WithCreator_RendersCreatorInContext() {
	req := s.reqWithAuth("GET", "/settings", nil)
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{SocialName: "João das Letras"}
	creator.ID = 42

	s.mockSessionService.On("RegenerateCSRFToken", mock.Anything, mock.Anything).Return("csrf-tok", nil)
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)
	s.mockSessionService.On("GetFlashes", mock.Anything, mock.Anything, "success").Return([]string{"ok"})
	s.mockSessionService.On("GetFlashes", mock.Anything, mock.Anything, "error").Return(nil)
	s.mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "settings",
		mock.MatchedBy(func(d map[string]interface{}) bool {
			c, ok := d["Creator"].(*accountmodel.Creator)
			return ok && c.SocialName == "João das Letras"
		}), "admin-daisy")

	s.handler.SettingsView(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	s.mockTemplateRenderer.AssertExpectations(s.T())
}

// ─── UpdateSocialName ─────────────────────────────────────────────────────────

func (s *SettingsHandlerSuite) TestUpdateSocialName_NoAuth_RedirectsToLogin() {
	req := httptest.NewRequest("POST", "/settings/social-name", nil)
	w := httptest.NewRecorder()
	s.handler.UpdateSocialName(w, req)
	assert.Equal(s.T(), http.StatusSeeOther, w.Code)
	assert.Equal(s.T(), "/login", w.Header().Get("Location"))
}

func (s *SettingsHandlerSuite) TestUpdateSocialName_ParseFormError_Returns400() {
	req := s.reqWithAuth("POST", "/settings/social-name", errReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.handler.UpdateSocialName(w, req)
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *SettingsHandlerSuite) TestUpdateSocialName_CreatorNotFound_FlashesAndRedirects() {
	form := strings.NewReader("social_name=João+das+Letras")
	req := s.reqWithAuth("POST", "/settings/social-name", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(nil, errors.New("not found"))
	s.mockSessionService.On("AddFlash", mock.Anything, mock.Anything, "Creator não encontrado", "error").Return(nil)

	s.handler.UpdateSocialName(w, req)

	assert.Equal(s.T(), http.StatusSeeOther, w.Code)
	assert.Equal(s.T(), "/settings", w.Header().Get("Location"))
	s.mockSessionService.AssertExpectations(s.T())
}

func (s *SettingsHandlerSuite) TestUpdateSocialName_UpdateCreatorError_FlashesAndRedirects() {
	form := strings.NewReader("social_name=João+das+Letras")
	req := s.reqWithAuth("POST", "/settings/social-name", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{}
	creator.ID = 1
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)
	s.mockCreatorService.On("UpdateCreator", mock.Anything).Return(errors.New("db error"))
	s.mockSessionService.On("AddFlash", mock.Anything, mock.Anything, "Erro ao salvar nome social", "error").Return(nil)

	s.handler.UpdateSocialName(w, req)

	assert.Equal(s.T(), http.StatusSeeOther, w.Code)
	assert.Equal(s.T(), "/settings", w.Header().Get("Location"))
	s.mockSessionService.AssertExpectations(s.T())
}

func (s *SettingsHandlerSuite) TestUpdateSocialName_Success_SavesSocialNameAndRedirects() {
	form := strings.NewReader("social_name=Escritor+Fant%C3%A1stico")
	req := s.reqWithAuth("POST", "/settings/social-name", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{}
	creator.ID = 1
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)
	s.mockCreatorService.On("UpdateCreator", mock.MatchedBy(func(c *accountmodel.Creator) bool {
		return c.SocialName == "Escritor Fantástico"
	})).Return(nil)
	s.mockSessionService.On("AddFlash", mock.Anything, mock.Anything, "Nome social atualizado com sucesso!", "success").Return(nil)

	s.handler.UpdateSocialName(w, req)

	assert.Equal(s.T(), http.StatusSeeOther, w.Code)
	assert.Equal(s.T(), "/settings", w.Header().Get("Location"))
	s.mockCreatorService.AssertExpectations(s.T())
	s.mockSessionService.AssertExpectations(s.T())
}

func TestSettingsHandlerSuite(t *testing.T) {
	suite.Run(t, new(SettingsHandlerSuite))
}
