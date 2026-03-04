package service

import (
	"errors"
	"log"
)

type StripePaymentGateway struct {
	stripeService *StripeService
}

func NewStripePaymentGateway(stripeService *StripeService) PaymentGateway {
	return &StripePaymentGateway{
		stripeService: stripeService,
	}
}

func (spg *StripePaymentGateway) CreateCustomer(email, name string) (string, error) {
	if email == "" {
		return "", errors.New("e-mail é obrigatório")
	}
	if name == "" {
		return "", errors.New("nome é obrigatório")
	}

	customerID, err := spg.stripeService.CreateCustomer(email, name)
	if err != nil {
		log.Printf("Error creating customer in Stripe: %v", err)
		return "", err
	}

	return customerID, nil
}

func (spg *StripePaymentGateway) CreateSubscription(customerID, priceID string) (string, error) {
	if customerID == "" {
		return "", errors.New("ID do cliente é obrigatório")
	}
	if priceID == "" {
		return "", errors.New("ID do preço é obrigatório")
	}

	err := spg.stripeService.CreateSubscription(customerID, priceID)
	if err != nil {
		log.Printf("Error creating subscription in Stripe: %v", err)
		return "", err
	}

	// Note: The current StripeService.CreateSubscription doesn't return subscription ID
	// This would need to be updated to return the subscription ID
	return "", nil
}

func (spg *StripePaymentGateway) CancelSubscription(subscriptionID string) error {
	if subscriptionID == "" {
		return errors.New("ID da assinatura é obrigatório")
	}

	err := spg.stripeService.CancelSubscription(subscriptionID)
	if err != nil {
		log.Printf("Error canceling subscription in Stripe: %v", err)
		return err
	}

	return nil
}
