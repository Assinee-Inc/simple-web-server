package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stripe/stripe-go/v76"
)

// Mock services for testing
type mockCreatorService struct {
	creator *models.Creator
	err     error
}

func (m *mockCreatorService) FindCreatorByEmail(email string) (*models.Creator, error) {
	return m.creator, m.err
}

func (m *mockCreatorService) CreateCreator(input models.InputCreateCreator) (*models.Creator, error) {
	return nil, nil
}

func (m *mockCreatorService) UpdateCreator(creator *models.Creator) error {
	return nil
}

func (m *mockCreatorService) FindByID(id uint) (*models.Creator, error) {
	return m.creator, m.err
}

func (m *mockCreatorService) FindCreatorByUserID(userID uint) (*models.Creator, error) {
	return m.creator, m.err
}

type mockStripeConnectService struct {
	account *stripe.Account
	err     error
}

func (m *mockStripeConnectService) GetAccountDetails(accountID string) (*stripe.Account, error) {
	if m.account == nil {
		return &stripe.Account{
			ID:               accountID,
			DetailsSubmitted: false,
			PayoutsEnabled:   false,
			ChargesEnabled:   false,
		}, m.err
	}
	return m.account, m.err
}

func (m *mockStripeConnectService) CreateConnectAccount(creator *models.Creator) (string, error) {
	return "acct_test", nil
}

func (m *mockStripeConnectService) CreateOnboardingLink(accountID, refreshURL, returnURL string) (string, error) {
	return "https://stripe.com/onboard", nil
}

func (m *mockStripeConnectService) UpdateCreatorFromAccount(creator *models.Creator, account *stripe.Account) error {
	return nil
}

type mockSessionService struct {
	email string
	err   error
}

func (m *mockSessionService) GenerateSessionToken() string {
	return "test_session_token"
}

func (m *mockSessionService) GenerateCSRFToken() string {
	return "test_csrf_token"
}

func (m *mockSessionService) SetSessionToken(w http.ResponseWriter) {}

func (m *mockSessionService) SetCSRFToken(w http.ResponseWriter) {}

func (m *mockSessionService) ClearSessionToken(w http.ResponseWriter) {}

func (m *mockSessionService) ClearCSRFToken(w http.ResponseWriter) {}

func (m *mockSessionService) GetSessionToken(r *http.Request) string {
	return "test_session_token"
}

func (m *mockSessionService) GetCSRFToken(r *http.Request) string {
	return "test_csrf_token"
}

func (m *mockSessionService) ClearSession(w http.ResponseWriter) {}

func (m *mockSessionService) SetSession(w http.ResponseWriter) {}

func (m *mockSessionService) GetSession(w http.ResponseWriter, r *http.Request) (string, string) {
	return "test_session_token", "test_csrf_token"
}

func (m *mockSessionService) InitSession(w http.ResponseWriter, email string) {}

func (m *mockSessionService) GetUserEmailFromSession(r *http.Request) (string, error) {
	return m.email, m.err
}

func TestStripeOnboardingMiddleware(t *testing.T) {
	tests := []struct {
		name                 string
		path                 string
		creator              *models.Creator
		expectedStatusCode   int
		expectedRedirectPath string
	}{
		{
			name: "Allows access to excluded paths",
			path: "/stripe-connect/status",
			creator: &models.Creator{
				Email:                  "test@example.com",
				StripeConnectAccountID: "",
				OnboardingCompleted:    false,
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Redirects when no Stripe account",
			path: "/dashboard",
			creator: &models.Creator{
				Email:                  "test@example.com",
				StripeConnectAccountID: "",
				OnboardingCompleted:    false,
			},
			expectedStatusCode:   http.StatusSeeOther,
			expectedRedirectPath: "/stripe-connect/welcome",
		},
		{
			name: "Redirects when onboarding incomplete",
			path: "/dashboard",
			creator: &models.Creator{
				Email:                  "test@example.com",
				StripeConnectAccountID: "acct_test",
				OnboardingCompleted:    false,
			},
			expectedStatusCode:   http.StatusSeeOther,
			expectedRedirectPath: "/stripe-connect/status",
		},
		{
			name: "Allows access when onboarding complete",
			path: "/dashboard",
			creator: &models.Creator{
				Email:                  "test@example.com",
				StripeConnectAccountID: "acct_test",
				OnboardingCompleted:    true,
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			creatorService := &mockCreatorService{creator: tt.creator}
			stripeService := &mockStripeConnectService{}
			sessionService := &mockSessionService{email: "test@example.com"}

			// Create middleware
			middleware := StripeOnboardingMiddleware(creatorService, stripeService, sessionService)

			// Create test handler
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			// Create request
			req := httptest.NewRequest("GET", tt.path, nil)
			req = req.WithContext(context.WithValue(req.Context(), UserEmailKey, "test@example.com"))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, rr.Code)
			}

			// Check redirect location if expected
			if tt.expectedRedirectPath != "" {
				location := rr.Header().Get("Location")
				if location != tt.expectedRedirectPath {
					t.Errorf("Expected redirect to %s, got %s", tt.expectedRedirectPath, location)
				}
			}
		})
	}
}
