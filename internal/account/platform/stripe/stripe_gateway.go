package stripe

import (
	"fmt"
	"log"

	"github.com/anglesson/simple-web-server/internal/account"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/client"
)

type StripeGateway struct {
	client *client.API
}

func NewStripeGateway(client *client.API) account.AccountGateway {
	return &StripeGateway{
		client: client,
	}
}

func (s *StripeGateway) CreateSellerAccount(account *account.Account) (string, error) {
	firstName, lastName := account.SplitName()

	params := &stripe.AccountParams{
		Type:         stripe.String("express"),
		Country:      stripe.String("BR"), // Brazil
		Email:        stripe.String(account.Email),
		BusinessType: stripe.String("individual"),
		Individual: &stripe.PersonParams{
			FirstName: stripe.String(firstName),
			LastName:  stripe.String(lastName),
			Email:     stripe.String(account.Email),
			Phone:     stripe.String("+55" + account.Phone),
			IDNumber:  stripe.String(account.CPF),
			DOB: &stripe.PersonDOBParams{
				Day:   stripe.Int64(int64(account.BirthDate.Day())),
				Month: stripe.Int64(int64(account.BirthDate.Month())),
				Year:  stripe.Int64(int64(account.BirthDate.Year())),
			},
		},
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
	}

	acc, err := s.client.Accounts.New(params)
	if err != nil {
		log.Printf("Error creating Stripe Connect account: %v", err)
		return "", fmt.Errorf("erro ao criar conta no Stripe: %v", err)
	}

	return acc.ID, nil
}
