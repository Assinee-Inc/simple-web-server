package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/handler/auth"
	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Create mock services
var mockUserService = new(mocks.MockUserService)
var mockSessionService = new(mocks.MockSessionService)
var mockEmailService = new(mocks.MockEmailService)
var mockTemplateRenderer = new(mocks.MockTemplateRenderer)

func TestLoginView(t *testing.T) {
	// Setup mock expectations
	mockSessionService.On("GenerateCSRFToken").Return("test-csrf-token")
	mockSessionService.On("SetCSRFToken", mock.AnythingOfType("*httptest.ResponseRecorder")).Return()
	mockTemplateRenderer.On("View", mock.AnythingOfType("*httptest.ResponseRecorder"), mock.AnythingOfType("*http.Request"), "auth/login", mock.AnythingOfType("map[string]interface {}"), "guest").Return()

	// Create auth handler
	authHandler := auth.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	authHandler.LoginView(w, req)

	// Verify template was called
	mockTemplateRenderer.AssertExpectations(t)
	mockSessionService.AssertExpectations(t)
}

func TestLoginSubmit_EmptyFields(t *testing.T) {
	// Create auth handler
	authHandler := auth.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	// Create form data with empty fields
	formData := url.Values{}
	formData.Set("email", "")
	formData.Set("password", "")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	authHandler.LoginSubmit(w, req)

	// Verify redirect back to login
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login")

	// Verify error cookies are set
	cookies := w.Result().Cookies()
	formCookie := findCookie(cookies, "form")
	errorsCookie := findCookie(cookies, "errors")
	assert.NotNil(t, formCookie)
	assert.NotNil(t, errorsCookie)

	// Verify error messages
	var errors map[string]string
	errorsJSON, _ := url.QueryUnescape(errorsCookie.Value)
	json.Unmarshal([]byte(errorsJSON), &errors)
	assert.Equal(t, "Email é obrigatório.", errors["email"])
	assert.Equal(t, "Senha é obrigatória.", errors["password"])
}

func TestLoginSubmit_InvalidCredentials(t *testing.T) {
	// Setup mock expectations for invalid credentials
	mockUserService.On("AuthenticateUser", models.InputLogin{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}).Return(nil, service.ErrInvalidCredentials)

	// Create auth handler
	authHandler := auth.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	// Create form data
	formData := url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("password", "wrongpassword")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	authHandler.LoginSubmit(w, req)

	// Verify redirect back to login
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login")

	// Verify error cookies are set
	cookies := w.Result().Cookies()
	formCookie := findCookie(cookies, "form")
	errorsCookie := findCookie(cookies, "errors")
	assert.NotNil(t, formCookie)
	assert.NotNil(t, errorsCookie)

	// Verify error message
	var errors map[string]string
	errorsJSON, _ := url.QueryUnescape(errorsCookie.Value)
	json.Unmarshal([]byte(errorsJSON), &errors)
	assert.Equal(t, "Email ou senha inválidos", errors["password"])

	mockUserService.AssertExpectations(t)
}

func TestLogoutSubmit(t *testing.T) {
	// Create mock services
	mockUserService := new(mocks.MockUserService)
	mockSessionService := new(mocks.MockSessionService)
	mockEmailService := new(mocks.MockEmailService)
	mockTemplateRenderer := new(mocks.MockTemplateRenderer)

	// Setup mock expectations for logout
	mockSessionService.On("ClearSession", mock.AnythingOfType("*httptest.ResponseRecorder")).Return()

	// Create auth handler
	authHandler := auth.NewAuthHandler(mockUserService, mockSessionService, mockEmailService, mockTemplateRenderer)

	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	authHandler.LogoutSubmit(w, req)

	// Verify redirect to home
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/")

	// Verify that ClearSession was called
	mockSessionService.AssertExpectations(t)
}

// Helper function to find a cookie by name
func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
