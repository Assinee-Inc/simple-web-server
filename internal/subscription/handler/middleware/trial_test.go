package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUser interface defines the contract for user-related operations in tests
type TestUser interface {
	IsInTrialPeriod() bool
	IsSubscribed() bool
}

// MockUser represents a test user with configurable trial and subscription status
type MockUser struct {
	inTrial    bool
	subscribed bool
}

func (u *MockUser) IsInTrialPeriod() bool {
	return u.inTrial
}

func (u *MockUser) IsSubscribed() bool {
	return u.subscribed
}

// TestAuthProvider interface defines the contract for authentication in tests
type TestAuthProvider interface {
	GetUser(r *http.Request) TestUser
}

// MockAuthProvider implements TestAuthProvider for testing
type MockAuthProvider struct {
	user TestUser
}

func (m *MockAuthProvider) GetUser(r *http.Request) TestUser {
	return m.user
}

// testTrialMiddleware is a wrapper that allows us to inject the creator provider for testing
func testTrialMiddleware(next http.Handler, authProvider TestAuthProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := authProvider.GetUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Skip trial check for these paths
		excludedPaths := map[string]bool{
			"/settings": true,
			"/logout":   true,
		}

		if excludedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		if !user.IsInTrialPeriod() && !user.IsSubscribed() {
			http.Redirect(w, r, "/settings", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func TestTrialMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		user           TestUser
		expectedStatus int
		expectedPath   string
	}{
		{
			name:           "No user - redirect to login",
			path:           "/any-path",
			user:           nil,
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/login",
		},
		{
			name:           "User in trial - allow access",
			path:           "/any-path",
			user:           &MockUser{inTrial: true, subscribed: false},
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
		{
			name:           "Subscribed user - allow access",
			path:           "/any-path",
			user:           &MockUser{inTrial: false, subscribed: true},
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
		{
			name:           "No trial, no subscription - redirect to settings",
			path:           "/any-path",
			user:           &MockUser{inTrial: false, subscribed: false},
			expectedStatus: http.StatusSeeOther,
			expectedPath:   "/settings",
		},
		{
			name:           "Settings path - allow access regardless of trial status",
			path:           "/settings",
			user:           &MockUser{inTrial: false, subscribed: false},
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
		{
			name:           "Logout path - allow access regardless of trial status",
			path:           "/logout",
			user:           &MockUser{inTrial: false, subscribed: false},
			expectedStatus: http.StatusOK,
			expectedPath:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			// Create a test handler that will be called if middleware passes
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create creator provider with our test user
			authProvider := &MockAuthProvider{user: tt.user}

			// Create and call the middleware with our test creator provider
			handler := testTrialMiddleware(nextHandler, authProvider)
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check redirect location if applicable
			if tt.expectedPath != "" {
				location := w.Header().Get("Location")
				if location != tt.expectedPath {
					t.Errorf("expected redirect to %s, got %s", tt.expectedPath, location)
				}
			}
		})
	}
}
