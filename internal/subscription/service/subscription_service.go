package service

import (
	"errors"
	"time"

	"github.com/anglesson/simple-web-server/internal/subscription/model"
	"github.com/anglesson/simple-web-server/internal/subscription/repository"
	"github.com/anglesson/simple-web-server/pkg/gov"
)

type SubscriptionService interface {
	CreateSubscription(userID uint, planID string) (uint, error)
	FindByUserID(userID uint) (*model.Subscription, error)
	FindByStripeCustomerID(customerID string) (*model.Subscription, error)
	FindByStripeSubscriptionID(subscriptionID string) (*model.Subscription, error)
	ActivateSubscription(subscriptionID uint, stripeCustomerID, stripeSubscriptionID string) error
	UpdateSubscriptionStatus(subscription *model.Subscription, status string, endDate *time.Time) error
	CancelSubscription(subscription *model.Subscription) error
	EndTrial(subscription *model.Subscription) error
	GetUserSubscriptionStatus(userID uint) (string, int, error)
}

type subscriptionServiceImpl struct {
	subscriptionRepository repository.SubscriptionRepository
	receitaFederalService  gov.ReceitaFederalService
}

func NewSubscriptionService(
	subscriptionRepository repository.SubscriptionRepository,
	receitaFederalService gov.ReceitaFederalService,
) SubscriptionService {
	return &subscriptionServiceImpl{
		subscriptionRepository: subscriptionRepository,
		receitaFederalService:  receitaFederalService,
	}
}

func (ss *subscriptionServiceImpl) CreateSubscription(userID uint, planID string) (uint, error) {
	if userID == 0 {
		return 0, errors.New("ID do usuário é obrigatório")
	}
	if planID == "" {
		return 0, errors.New("ID do plano é obrigatório")
	}

	subscription := model.NewSubscription(userID, planID)

	err := ss.subscriptionRepository.Create(subscription)
	if err != nil {
		return 0, err
	}

	return subscription.ID, nil
}

func (ss *subscriptionServiceImpl) FindByUserID(userID uint) (*model.Subscription, error) {
	if userID == 0 {
		return nil, errors.New("ID do usuário é obrigatório")
	}

	return ss.subscriptionRepository.FindByUserID(userID)
}

func (ss *subscriptionServiceImpl) FindByStripeCustomerID(customerID string) (*model.Subscription, error) {
	if customerID == "" {
		return nil, errors.New("ID do cliente é obrigatório")
	}

	return ss.subscriptionRepository.FindByStripeCustomerID(customerID)
}

func (ss *subscriptionServiceImpl) FindByStripeSubscriptionID(subscriptionID string) (*model.Subscription, error) {
	if subscriptionID == "" {
		return nil, errors.New("ID da assinatura é obrigatório")
	}

	return ss.subscriptionRepository.FindByStripeSubscriptionID(subscriptionID)
}

func (ss *subscriptionServiceImpl) ActivateSubscription(subscriptionID uint, stripeCustomerID, stripeSubscriptionID string) error {
	if subscriptionID == 0 {
		return errors.New("ID da assinatura é obrigatório")
	}
	if stripeCustomerID == "" {
		return errors.New("ID do cliente Stripe é obrigatório")
	}
	if stripeSubscriptionID == "" {
		return errors.New("ID da assinatura Stripe é obrigatório")
	}

	subscription, err := ss.subscriptionRepository.FindByID(subscriptionID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return errors.New("assinatura não encontrada")
	}

	subscription.ActivateSubscription(stripeCustomerID, stripeSubscriptionID)

	return ss.subscriptionRepository.Save(subscription)
}

func (ss *subscriptionServiceImpl) UpdateSubscriptionStatus(subscription *model.Subscription, status string, endDate *time.Time) error {
	if subscription == nil {
		return errors.New("assinatura é obrigatória")
	}
	if status == "" {
		return errors.New("status é obrigatório")
	}

	subscription.UpdateSubscriptionStatus(status, endDate)

	return ss.subscriptionRepository.Save(subscription)
}

func (ss *subscriptionServiceImpl) CancelSubscription(subscription *model.Subscription) error {
	if subscription == nil {
		return errors.New("assinatura é obrigatória")
	}

	subscription.CancelSubscription()

	return ss.subscriptionRepository.Save(subscription)
}

func (ss *subscriptionServiceImpl) EndTrial(subscription *model.Subscription) error {
	if subscription == nil {
		return errors.New("assinatura é obrigatória")
	}

	subscription.EndTrial()

	return ss.subscriptionRepository.Save(subscription)
}

func (ss *subscriptionServiceImpl) GetUserSubscriptionStatus(userID uint) (string, int, error) {
	if userID == 0 {
		return "inactive", 0, errors.New("ID do usuário é obrigatório")
	}

	subscription, err := ss.subscriptionRepository.FindByUserID(userID)
	if err != nil {
		return "inactive", 0, err
	}

	if subscription == nil {
		return "inactive", 0, nil
	}

	status := subscription.GetSubscriptionStatus()
	daysLeft := 0

	if subscription.IsInTrialPeriod() {
		daysLeft = subscription.DaysLeftInTrial()
	} else if subscription.IsSubscribed() {
		daysLeft = subscription.DaysLeftInSubscription()
	}

	return status, daysLeft, nil
}
