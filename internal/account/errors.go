package account

import "errors"

var (
	InternalError          = errors.New("Internal server error")
	ErrInvalidAccount      = errors.New("Invalid account data")
	StripeIntegrationError = errors.New("Stripe integration error")
)
