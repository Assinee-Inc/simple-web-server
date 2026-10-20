package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	accounthandler "github.com/Assinee-Inc/docffy/internal/account/handler"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	"github.com/Assinee-Inc/docffy/internal/mocks"
	"github.com/Assinee-Inc/docffy/internal/subscription/model"
	"github.com/Assinee-Inc/docffy/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/stripe/stripe-go/v76"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type errReader struct{}

func (e errReader) Read(p []byte) (int, error) { return 0, errors.New("forced read error") }

type SettingsHandlerSuite struct {
	suite.Suite
	db                 *gorm.DB
	handler            *accounthandler.SettingsHandler
	mockSessionService *mocks.MockSessionService
	mockCreatorService *mocks.MockCreatorService
	testUser           *authmodel.User
}

func (s *SettingsHandlerSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.AutoMigrate(&authmodel.User{}, &model.Subscription{}, &accountmodel.Creator{}))
	s.db = db
	database.DB = db

	userRepo := authrepo.NewGormUserRepository(db)
	user := authmodel.NewGoogleUser("Settings Test", "settings_test@shtest.com", "google-settings")
	require.NoError(s.T(), userRepo.Create(user))

	s.testUser = userRepo.FindByEmail("settings_test@shtest.com")
	s.Require().NotNil(s.testUser)
}

func (s *SettingsHandlerSuite) SetupTest() {
	s.mockSessionService = new(mocks.MockSessionService)
	s.mockCreatorService = new(mocks.MockCreatorService)
	s.handler = accounthandler.NewSettingsHandler(s.mockSessionService, s.mockCreatorService)
}

func (s *SettingsHandlerSuite) reqWithAuth(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	ctx := authmw.WithAuthenticatedUser(req.Context(), s.testUser)
	ctx = context.WithValue(ctx, authmw.UserEmailKey, s.testUser.Email)
	return req.WithContext(ctx)
}

func (s *SettingsHandlerSuite) TestSettingsView_NoAuth_Returns401JSON() {
	req := httptest.NewRequest("GET", "/settings", nil)
	w := httptest.NewRecorder()
	s.handler.SettingsView(w, req)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
	var resp map[string]any
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), false, resp["success"])
}

func (s *SettingsHandlerSuite) TestSettingsView_WithCreator_Returns200JSON() {
	req := s.reqWithAuth("GET", "/settings", nil)
	w := httptest.NewRecorder()

	creator := &accountmodel.Creator{
		PublicID:        "crt_42",
		SocialName:      "João das Letras",
		FacebookPixelID: "1234567890",
	}
	creator.ID = 42

	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)

	s.handler.SettingsView(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), true, resp["success"])
	assert.Equal(s.T(), "João das Letras", resp["social_name"])
	assert.Equal(s.T(), "1234567890", resp["facebook_pixel_id"])
}

func (s *SettingsHandlerSuite) TestUpdateSocialName_NoAuth_Returns401JSON() {
	req := httptest.NewRequest("POST", "/settings/social-name", nil)
	w := httptest.NewRecorder()
	s.handler.UpdateSocialName(w, req)
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
}

func (s *SettingsHandlerSuite) TestUpdateSocialName_CreatorNotFound_Returns404JSON() {
	form := strings.NewReader("social_name=João+das+Letras")
	req := s.reqWithAuth("POST", "/settings/social-name", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(nil, errors.New("not found"))

	s.handler.UpdateSocialName(w, req)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

func (s *SettingsHandlerSuite) TestUpdateSocialName_Success_JSONAndForm() {
	creator := &accountmodel.Creator{}
	creator.ID = 1
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)
	s.mockCreatorService.On("UpdateCreator", mock.MatchedBy(func(c *accountmodel.Creator) bool {
		return c.SocialName == "Escritor Fantástico"
	})).Return(nil)

	body, _ := json.Marshal(map[string]string{"social_name": "Escritor Fantástico"})
	req := s.reqWithAuth("POST", "/settings/social-name", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.UpdateSocialName(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), true, resp["success"])
	assert.Equal(s.T(), "Escritor Fantástico", resp["social_name"])
}

func (s *SettingsHandlerSuite) TestUpdateFacebookPixel_InvalidPixelID_Returns400JSON() {
	form := strings.NewReader("facebook_pixel_id=not-a-pixel-id")
	req := s.reqWithAuth("POST", "/settings/facebook-pixel", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(&accountmodel.Creator{}, nil)

	s.handler.UpdateFacebookPixel(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), false, resp["success"])
}

func (s *SettingsHandlerSuite) TestUpdateFacebookPixel_Success_Returns200JSON() {
	creator := &accountmodel.Creator{}
	creator.ID = 1
	s.mockCreatorService.On("FindCreatorByUserID", s.testUser.ID).Return(creator, nil)
	s.mockCreatorService.On("UpdateCreator", mock.MatchedBy(func(c *accountmodel.Creator) bool {
		return c.FacebookPixelID == "987654321012345"
	})).Return(nil)

	body, _ := json.Marshal(map[string]string{"facebook_pixel_id": "987654321012345"})
	req := s.reqWithAuth("POST", "/settings/facebook-pixel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.UpdateFacebookPixel(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), true, resp["success"])
	assert.Equal(s.T(), "987654321012345", resp["facebook_pixel_id"])
}

func TestSettingsHandlerSuite(t *testing.T) {
	suite.Run(t, new(SettingsHandlerSuite))
}

func TestStripeConnectHandler_JSONResponses(t *testing.T) {
	mockCreatorSvc := new(mocks.MockCreatorService)
	mockStripeSvc := new(mockStripeConnectService)
	h := accounthandler.NewStripeConnectHandler(mockStripeSvc, mockCreatorSvc)

	user := &authmodel.User{Email: "creator@example.com"}
	user.ID = 55

	creator := &accountmodel.Creator{
		PublicID:               "crt_55",
		UserID:                 55,
		Name:                   "Maria",
		Email:                  "creator@example.com",
		StripeConnectAccountID: "acct_55",
		OnboardingCompleted:    false,
		ChargesEnabled:         false,
		PayoutsEnabled:         false,
	}

	t.Run("GET /stripe-connect/status retorna JSON 200", func(t *testing.T) {
		mockCreatorSvc.On("FindCreatorByUserID", uint(55)).Return(creator, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/stripe-connect/status", nil)
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		h.OnboardingStatus(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, true, resp["success"])
		assert.Equal(t, true, resp["has_stripe_account"])
		assert.Equal(t, true, resp["needs_onboarding"])
	})

	t.Run("GET /stripe-connect/onboard retorna onboarding_url em JSON 200", func(t *testing.T) {
		mockCreatorSvc.On("FindCreatorByUserID", uint(55)).Return(creator, nil).Once()
		mockStripeSvc.On("CreateOnboardingLink", "acct_55", mock.Anything, mock.Anything).Return("https://connect.stripe.com/setup/s/123", nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/stripe-connect/onboard", nil)
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		h.StartOnboarding(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, true, resp["success"])
		assert.Equal(t, "https://connect.stripe.com/setup/s/123", resp["onboarding_url"])
	})

	t.Run("GET /stripe-connect/complete sincroniza conta e retorna status em JSON 200", func(t *testing.T) {
		mockCreatorSvc.On("FindCreatorByUserID", uint(55)).Return(creator, nil).Once()
		stripeAcc := &stripe.Account{
			ID:               "acct_55",
			DetailsSubmitted: true,
			ChargesEnabled:   true,
			PayoutsEnabled:   true,
		}
		mockStripeSvc.On("GetAccountDetails", "acct_55").Return(stripeAcc, nil).Once()
		mockStripeSvc.On("UpdateCreatorFromAccount", creator, stripeAcc).Return(nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/stripe-connect/complete", nil)
		req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), user))
		rec := httptest.NewRecorder()

		h.CompleteOnboarding(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, true, resp["success"])
		assert.Equal(t, "completed", resp["onboarding_status"])
	})
}
