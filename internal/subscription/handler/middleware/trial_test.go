package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	"github.com/Assinee-Inc/docffy/internal/subscription/handler/middleware"
	submodel "github.com/Assinee-Inc/docffy/internal/subscription/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildUserWithSub(inTrial, subscribed bool) *authmodel.User {
	user := authmodel.NewGoogleUser("Trial User", "trial@example.com", "google-trial")
	user.ID = 100
	sub := submodel.NewSubscription(user.ID, "default_plan")
	if !inTrial {
		sub.IsTrialActive = false
		sub.TrialEndDate = time.Now().AddDate(0, 0, -2)
	}
	if subscribed {
		sub.SubscriptionStatus = "active"
	} else {
		sub.SubscriptionStatus = "inactive"
	}
	user.Subscription = sub
	return user
}

func TestTrialMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		user           *authmodel.User
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "No user - returns 401 JSON",
			path:           "/any-path",
			user:           nil,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "User in trial - allow access",
			path:           "/any-path",
			user:           buildUserWithSub(true, false),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Subscribed user - allow access",
			path:           "/any-path",
			user:           buildUserWithSub(false, true),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No trial, no subscription - returns 403 JSON SUBSCRIPTION_REQUIRED",
			path:           "/any-path",
			user:           buildUserWithSub(false, false),
			expectedStatus: http.StatusForbidden,
			expectedCode:   "SUBSCRIPTION_REQUIRED",
		},
		{
			name:           "Settings path - allow access regardless of trial status",
			path:           "/settings",
			user:           buildUserWithSub(false, false),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Logout path - allow access regardless of trial status",
			path:           "/logout",
			user:           buildUserWithSub(false, false),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.user != nil {
				req = req.WithContext(authmw.WithAuthenticatedUser(req.Context(), tt.user))
			}
			w := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.TrialMiddleware(nextHandler)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				var body map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				assert.Equal(t, false, body["success"])
				assert.Equal(t, tt.expectedCode, body["code"])
			}
		})
	}
}
