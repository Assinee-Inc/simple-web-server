package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/pkg/database"
)

// First, define a custom type for context keys (typically at package level)
type contextKey string

// Define a constant for your key
const UserEmailKey contextKey = "user_email"
const CSRFTokenKey contextKey = "csrf_token"
const UserKey contextKey = "user"

var ErrUnauthorized = errors.New("Unauthorized")

func authorizer(r *http.Request, sessionService authsvc.SessionService) (string, error) {
	// Get user email from session
	userEmail := sessionService.Get(r, authsvc.UserEmailKey)
	if userEmail == nil {
		log.Printf("User email not found in session")
		return "", ErrUnauthorized
	}

	email, ok := userEmail.(string)
	if !ok || email == "" {
		log.Printf("Invalid user email in session")
		return "", ErrUnauthorized
	}

	// Find user by email
	userRepository := authrepo.NewGormUserRepository(database.DB)
	user := userRepository.FindByUserEmail(email)
	if user == nil {
		log.Printf("User not found for email: %s", email)
		return "", ErrUnauthorized
	}

	// Get CSRF token from session
	csrfTokenFromSession := sessionService.Get(r, authsvc.CSRFTokenKey)
	if csrfTokenFromSession == nil {
		log.Printf("CSRF token not found in session for user: %s", user.Email)
		return "", ErrUnauthorized
	}

	csrfToken, ok := csrfTokenFromSession.(string)
	if !ok || csrfToken == "" {
		log.Printf("Invalid CSRF token in session for user: %s", user.Email)
		return "", ErrUnauthorized
	}

	// Store the email and CSRF token in request context
	ctx := context.WithValue(r.Context(), UserEmailKey, user.Email)
	ctx = context.WithValue(ctx, CSRFTokenKey, csrfToken)
	ctx = context.WithValue(ctx, UserKey, user)
	*r = *r.WithContext(ctx)

	return csrfToken, nil
}

func AuthMiddleware(sessionService authsvc.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Authentication logic
			csrfToken, err := authorizer(r, sessionService)
			if err != nil {
				log.Printf("Unauthorized access attempt: %v", err)

				// Check if it's an API request
				if strings.HasPrefix(r.URL.Path, "/api/") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Não autorizado",
					})
					return
				}

				// For regular page requests, redirect to login
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Email verification gate
			user := r.Context().Value(UserKey).(*model.User)
			if !user.IsEmailVerified() {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{"error": "E-mail não verificado"})
					return
				}
				http.Redirect(w, r, "/email-not-verified", http.StatusSeeOther)
				return
			}

			// Store CSRF token in a header that your templates can access
			w.Header().Set("X-CSRF-Token", csrfToken)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetCSRFToken retrieves the CSRF token from the request context
func GetCSRFToken(r *http.Request) string {
	if token, ok := r.Context().Value(CSRFTokenKey).(string); ok {
		return token
	}
	return ""
}

func Auth(r *http.Request) *model.User {
	user_email, ok := r.Context().Value(UserEmailKey).(string)
	if !ok {
		return nil
	}

	userRepository := authrepo.NewGormUserRepository(database.DB)
	user := userRepository.FindByEmail(user_email)

	return user
}
