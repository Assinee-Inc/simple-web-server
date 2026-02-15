package account

import "errors"

var InternalError = errors.New("Internal server error")
var StripeIntegrationError = errors.New("Stripe integration error")
