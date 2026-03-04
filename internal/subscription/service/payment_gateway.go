package service

// PaymentGateway interface for payment operations
type PaymentGateway interface {
	CreateCustomer(email, name string) (string, error)
	CreateSubscription(customerID, priceID string) (string, error)
	CancelSubscription(subscriptionID string) error
}
