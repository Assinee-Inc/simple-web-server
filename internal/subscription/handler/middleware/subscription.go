package middleware

import (
	"context"
	"net/http"

	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	"github.com/Assinee-Inc/docffy/internal/subscription/service"
)

// SubscriptionDataKey is the context key for subscription data
type SubscriptionDataKey string

const SubscriptionDataKeyValue SubscriptionDataKey = "subscription_data"

// SubscriptionData contains subscription status information
type SubscriptionData struct {
	Status   string `json:"status"`
	DaysLeft int    `json:"days_left"`
}

// SubscriptionMiddleware adds subscription status data to the request context
func SubscriptionMiddleware(subscriptionService service.SubscriptionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID uint
			if uid, ok := authmw.GetAuthenticatedUserID(r.Context()); ok && uid > 0 {
				userID = uid
			} else if user := authmw.Auth(r); user != nil && user.ID > 0 {
				userID = user.ID
			}

			if userID == 0 || subscriptionService == nil {
				next.ServeHTTP(w, r)
				return
			}

			status, daysLeft, err := subscriptionService.GetUserSubscriptionStatus(userID)
			if err != nil {
				status = "inactive"
				daysLeft = 0
			}

			subscriptionData := SubscriptionData{
				Status:   status,
				DaysLeft: daysLeft,
			}

			ctx := context.WithValue(r.Context(), SubscriptionDataKeyValue, subscriptionData)
			next.ServeHTTP(w, r.WithContext(ctx))
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
