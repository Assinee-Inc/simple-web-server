package middleware

import (
	"context"
	"net/http"

	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/service"
)

// SubscriptionDataKey is the context key for subscription data
type SubscriptionDataKey string

const SubscriptionDataKeyValue SubscriptionDataKey = "subscription_data"

// SubscriptionData contains subscription status information
type SubscriptionData struct {
	Status   string
	DaysLeft int
}

// SubscriptionMiddleware adds subscription status data to the request context
func SubscriptionMiddleware(subscriptionService service.SubscriptionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user := authmw.Auth(r)
			if user == nil || user.ID == 0 {
				// If no user, continue without subscription data
				next.ServeHTTP(w, r)
				return
			}

			// Get subscription status
			status, daysLeft, err := subscriptionService.GetUserSubscriptionStatus(user.ID)
			if err != nil {
				// If error, use default values
				status = "inactive"
				daysLeft = 0
			}

			// Create subscription data
			subscriptionData := SubscriptionData{
				Status:   status,
				DaysLeft: daysLeft,
			}

			// Add to context
			ctx := context.WithValue(r.Context(), SubscriptionDataKeyValue, subscriptionData)
			*r = *r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetSubscriptionData retrieves subscription data from request context
func GetSubscriptionData(r *http.Request) *SubscriptionData {
	if data, ok := r.Context().Value(SubscriptionDataKeyValue).(SubscriptionData); ok {
		return &data
	}
	return &SubscriptionData{
		Status:   "inactive",
		DaysLeft: 0,
	}
}
