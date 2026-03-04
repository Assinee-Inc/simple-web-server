package repository

import (
	"github.com/anglesson/simple-web-server/internal/subscription/model"
)

type SubscriptionRepository interface {
	Create(subscription *model.Subscription) error
	FindByUserID(userID uint) (*model.Subscription, error)
	FindByStripeCustomerID(customerID string) (*model.Subscription, error)
	FindByStripeSubscriptionID(subscriptionID string) (*model.Subscription, error)
	Update(subscription *model.Subscription) error
	Save(subscription *model.Subscription) error
}
