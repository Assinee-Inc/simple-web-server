package service

import (
	"log"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/subscription"
)

type StripeService struct {
	apiKey string
}

func NewStripeService() *StripeService {
	stripe.Key = config.AppConfig.StripeSecretKey
	return &StripeService{
		apiKey: config.AppConfig.StripeSecretKey,
	}
}

func (s *StripeService) CreateCustomer(email, name string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	c, err := customer.New(params)
	if err != nil {
		log.Printf("Error creating Stripe customer: %v", err)
		return "", err
	}

	return c.ID, nil
}

func (s *StripeService) CreateSubscription(customerID string, priceID string) error {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
	}

	_, err := subscription.New(params)
	if err != nil {
		log.Printf("Error creating subscription: %v", err)
		return err
	}

	return nil
}

func (s *StripeService) CancelSubscription(subscriptionID string) error {
	_, err := subscription.Cancel(subscriptionID, nil)
	if err != nil {
		log.Printf("Error canceling subscription: %v", err)
		return err
	}

	return nil
}
