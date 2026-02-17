package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/account"
	"github.com/stripe/stripe-go/v76/accountlink"
)

type StripeConnectService interface {
	CreateConnectAccount(creator *models.Creator) (string, error)
	CreateOnboardingLink(accountID string, refreshURL string, returnURL string) (string, error)
	GetAccountDetails(accountID string) (*stripe.Account, error)
	UpdateCreatorFromAccount(creator *models.Creator, account *stripe.Account) error
	GerarLinkRemediacao(creator *models.Creator) (string, error)
}

type stripeConnectServiceImpl struct {
	creatorSvc CreatorService
}

func NewStripeConnectService(creatorSvc CreatorService) StripeConnectService {
	stripe.Key = config.AppConfig.StripeSecretKey
	return &stripeConnectServiceImpl{
		creatorSvc,
	}
}

// CreateConnectAccount creates a new Stripe Connect account for the creator
func (s *stripeConnectServiceImpl) CreateConnectAccount(creator *models.Creator) (string, error) {
	// Split name into first and last name (basic approach)
	names := strings.Fields(creator.Name)
	firstName := names[0]
	lastName := ""
	if len(names) > 1 {
		lastName = strings.Join(names[1:], " ")
	}

	params := &stripe.AccountParams{
		Type:         stripe.String("express"),
		Country:      stripe.String("BR"), // Brazil
		Email:        stripe.String(creator.Email),
		BusinessType: stripe.String("individual"),
		Individual: &stripe.PersonParams{
			FirstName: stripe.String(firstName),
			LastName:  stripe.String(lastName),
			Email:     stripe.String(creator.Email),
			Phone:     stripe.String("+55" + creator.Phone),
			IDNumber:  stripe.String(creator.CPF),
			DOB: &stripe.PersonDOBParams{
				Day:   stripe.Int64(int64(creator.BirthDate.Day())),
				Month: stripe.Int64(int64(creator.BirthDate.Month())),
				Year:  stripe.Int64(int64(creator.BirthDate.Year())),
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

	acc, err := account.New(params)
	if err != nil {
		log.Printf("Error creating Stripe Connect account: %v", err)
		return "", fmt.Errorf("erro ao criar conta no Stripe: %v", err)
	}

	return acc.ID, nil
}

// CreateOnboardingLink creates an onboarding link for the creator to complete their Stripe setup
func (s *stripeConnectServiceImpl) CreateOnboardingLink(accountID string, refreshURL string, returnURL string) (string, error) {
	params := &stripe.AccountLinkParams{
		Account:    stripe.String(accountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(params)
	if err != nil {
		log.Printf("Error creating onboarding link: %v", err)
		return "", fmt.Errorf("erro ao criar link de onboarding: %v", err)
	}

	return link.URL, nil
}

// GetAccountDetails retrieves account details from Stripe
func (s *stripeConnectServiceImpl) GetAccountDetails(accountID string) (*stripe.Account, error) {
	acc, err := account.GetByID(accountID, nil)
	if err != nil {
		log.Printf("Error retrieving account details: %v", err)
		return nil, fmt.Errorf("erro ao buscar detalhes da conta: %v", err)
	}

	return acc, nil
}

// UpdateCreatorFromAccount updates creator with account status from Stripe
func (s *stripeConnectServiceImpl) UpdateCreatorFromAccount(creator *models.Creator, account *stripe.Account) error {
	log.Printf("Atualizando creator %s com status da conta Stripe: detailsSubmitted=%v, chargesEnabled=%v, payoutsEnabled=%v",
		creator.Email, account.DetailsSubmitted, account.ChargesEnabled, account.PayoutsEnabled)

	if account.Requirements.Errors != nil && len(account.Requirements.Errors) > 0 {
		log.Printf("Conta Stripe do creator %s tem pendências: %v", creator.Email, account.Requirements.Errors)
		urlRefresh, err := s.GerarLinkRemediacao(creator)
		if urlRefresh != "" && err == nil {
			log.Printf("Gerado link de remediação para creator %s: %s", creator.Email, urlRefresh)
			creator.OnboardingRefreshURL = urlRefresh
		}
	}

	creator.OnboardingCompleted = account.DetailsSubmitted
	creator.PayoutsEnabled = account.PayoutsEnabled
	creator.ChargesEnabled = account.ChargesEnabled

	err := s.creatorSvc.UpdateCreator(creator)
	if err != nil {
		log.Printf("Error updating creator: %v", err)
		return fmt.Errorf("erro ao atualizar criador: %v", err)
	}

	return nil
}

func (s *stripeConnectServiceImpl) GerarLinkRemediacao(creator *models.Creator) (string, error) {
	if creator.StripeConnectAccountID == "" {
		return "", fmt.Errorf("creator não possui conta Stripe Connect")
	}

	params := &stripe.AccountLinkParams{
		Account:    stripe.String(creator.StripeConnectAccountID),
		Type:       stripe.String("account_onboarding"),
		ReturnURL:  stripe.String(config.AppConfig.Host + "/stripe-connect/status"),
		RefreshURL: stripe.String(config.AppConfig.Host + "/stripe-connect/status"),
	}
	result, err := accountlink.New(params)
	if err != nil {
		log.Printf("Error creating remediation link: %v", err)
		return "", fmt.Errorf("erro ao criar link de remediação: %v", err)
	}

	return result.URL, err
}
