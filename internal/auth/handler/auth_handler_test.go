package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	authhandler "github.com/anglesson/simple-web-server/internal/auth/handler"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginView(t *testing.T) {
	// Arrange
	mockUserService := new(mocks.MockUserService)
	mockSessionService := new(mocks.MockSessionService)
	mockEmailService := new(mocks.MockEmailService)
	mockTemplateRenderer := new(mocks.MockTemplateRenderer)

	h := authhandler.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockSessionService.On("RegenerateCSRFToken", mock.Anything, mock.Anything).Return("test-csrf-token", nil)
	mockSessionService.On("GetFlashes", mock.Anything, mock.Anything, "form-error").Return([]string{})
	mockSessionService.On("GetFlashes", mock.Anything, mock.Anything, "success").Return([]string{})
	mockSessionService.On("Get", mock.Anything, "form").Return(nil)
	mockSessionService.On("Pop", mock.Anything, mock.Anything, "form").Return(nil)
	mockTemplateRenderer.On("View", mock.Anything, mock.Anything, "auth/login", mock.Anything, "guest").Return()

	// Act
	h.LoginView(w, req)

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

	h := authhandler.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	formData := url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("password", "wrongpassword")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockUserService.On("AuthenticateUser", mock.Anything).Return(nil, authsvc.ErrInvalidCredentials)
	mockSessionService.On("Set", mock.Anything, mock.Anything, "form", mock.Anything).Return(nil)
	mockSessionService.On("AddFlash", mock.Anything, mock.Anything, "email ou senha inválidos", "form-error").Return(nil)

	// Act
	h.LoginSubmit(w, req)

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

	h := authhandler.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations
	mockSessionService.On("Destroy", mock.Anything, mock.Anything).Return(nil)

	// Act
	h.LogoutSubmit(w, req)

	// Assert
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
	mockSessionService.AssertExpectations(t)
}
