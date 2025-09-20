package auth_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/handler/auth"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginView(t *testing.T) {
	// Arrange
	mockUserService := new(mocks.MockUserService)
	mockSessionService := new(mocks.MockSessionService)
	mockEmailService := new(mocks.MockEmailService)
	mockTemplateRenderer := new(mocks.MockTemplateRenderer)

	authHandler := auth.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockSessionService.On("RegenerateCSRFToken", mock.Anything, mock.Anything).Return("test-csrf-token", nil)
	mockSessionService.On("GetFlashes", mock.Anything, mock.Anything, "error").Return([]string{})
	mockSessionService.On("GetFlashes", mock.Anything, mock.Anything, "success").Return([]string{})
	mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "auth/login", mock.Anything, "guest").Return()

	// Act
	authHandler.LoginView(w, req)

	// Assert
	mockTemplateRenderer.AssertExpectations(t)
	mockSessionService.AssertExpectations(t)
}

func TestLoginSubmit_InvalidCredentials(t *testing.T) {
	// Arrange
	mockUserService := new(mocks.MockUserService)
	mockSessionService := new(mocks.MockSessionService)
	mockEmailService := new(mocks.MockEmailService)
	mockTemplateRenderer := new(mocks.MockTemplateRenderer)

	authHandler := auth.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	formData := url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("password", "wrongpassword")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockUserService.On("AuthenticateUser", mock.Anything).Return(nil, service.ErrInvalidCredentials)
	mockSessionService.On("AddFlash", mock.Anything, mock.Anything, "Email ou senha inválidos", "error").Return(nil)

	// Act
	authHandler.LoginSubmit(w, req)

	// Assert
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))
	mockUserService.AssertExpectations(t)
	mockSessionService.AssertExpectations(t)
}

func TestLogoutSubmit(t *testing.T) {
	// Arrange
	mockUserService := new(mocks.MockUserService)
	mockSessionService := new(mocks.MockSessionService)
	mockEmailService := new(mocks.MockEmailService)
	mockTemplateRenderer := new(mocks.MockTemplateRenderer)

	authHandler := auth.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockSessionService.On("Destroy", mock.Anything, mock.Anything).Return(nil)

	// Act
	authHandler.LogoutSubmit(w, req)

	// Assert
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
	mockSessionService.AssertExpectations(t)
}
