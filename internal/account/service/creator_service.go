package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	accountrepo "github.com/anglesson/simple-web-server/internal/account/repository"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	subscriptionservice "github.com/anglesson/simple-web-server/internal/subscription/service"
	"github.com/anglesson/simple-web-server/pkg/gov"
)

type CreatorService interface {
	CreateCreator(input InputCreateCreator) (*accountmodel.Creator, error)
	FindCreatorByEmail(email string) (*accountmodel.Creator, error)
	FindCreatorByUserID(userID uint) (*accountmodel.Creator, error)
	FindByID(id uint) (*accountmodel.Creator, error)
	FindByPublicID(publicID string) (*accountmodel.Creator, error)
	UpdateCreator(creator *accountmodel.Creator) error
}

type creatorServiceImpl struct {
	creatorRepo         accountrepo.CreatorRepository
	rfService           gov.ReceitaFederalService
	userService         authsvc.UserService
	subscriptionService subscriptionservice.SubscriptionService
	paymentGateway      subscriptionservice.PaymentGateway
}

func NewCreatorService(
	creatorRepo accountrepo.CreatorRepository,
	receitaFederalService gov.ReceitaFederalService,
	userService authsvc.UserService,
	subscriptionService subscriptionservice.SubscriptionService,
	paymentGateway subscriptionservice.PaymentGateway,
) CreatorService {
	return &creatorServiceImpl{
		creatorRepo:         creatorRepo,
		rfService:           receitaFederalService,
		userService:         userService,
		subscriptionService: subscriptionService,
		paymentGateway:      paymentGateway,
	}
}

func (cs *creatorServiceImpl) CreateCreator(input InputCreateCreator) (*accountmodel.Creator, error) {
	// Validate input
	if err := validateCreatorInput(input); err != nil {
		return nil, err
	}

	// Parse birth date - try DD/MM/YYYY format first (from jmask), then YYYY-MM-DD
	var birthDate time.Time
	birthDate, err := time.Parse("02/01/2006", input.BirthDate)
	if err != nil {
		// If that fails, try YYYY-MM-DD format (from HTML date input)
		birthDate, err = time.Parse("2006-01-02", input.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("formato de data de nascimento inválido: %w", err)
		}
	}

	// Clean CPF (remove non-digits)
	cleanCPF := cleanCPF(input.CPF)

	// Check if creator already exists
	creatorExists, err := cs.creatorRepo.FindByCPF(cleanCPF)
	if err != nil {
		return nil, err
	}

	if creatorExists != nil {
		return nil, errors.New("criador já existe")
	}

	// Validate with Receita Federal
	validatedName := input.Name
	validatedName, err = cs.validateReceita(validatedName, cleanCPF, birthDate)
	if err != nil {
		return nil, err
	}

	// Create user
	inputCreateUser := authsvc.InputCreateUser{
		Username:             validatedName,
		Email:                input.Email,
		Password:             input.Password,
		PasswordConfirmation: input.PasswordConfirmation,
	}

	userID, err := cs.userService.CreateUser(inputCreateUser)
	if err != nil {
		return nil, err
	}

	// Create creator
	creator := accountmodel.NewCreator(
		validatedName,
		input.Email,
		cleanPhone(input.PhoneNumber),
		cleanCPF,
		birthDate,
		userID,
	)

	// Save creator
	err = cs.creatorRepo.Create(creator)
	if err != nil {
		return nil, err
	}

	// Create customer in payment gateway
	customerID, err := cs.paymentGateway.CreateCustomer(input.Email, validatedName)
	if err != nil {
		log.Printf("Error creating customer in payment gateway: %v", err)
		// Don't fail the creator creation if payment gateway fails
	} else {
		// Create subscription for the creator
		subscriptionID, err := cs.subscriptionService.CreateSubscription(userID, "default_plan")
		if err != nil {
			log.Printf("Error creating subscription: %v", err)
		} else {
			// Activate subscription with customer ID
			err = cs.subscriptionService.ActivateSubscription(subscriptionID, customerID, "")
			if err != nil {
				log.Printf("Error activating subscription: %v", err)
			}
		}
	}

	return creator, nil
}

func (cs *creatorServiceImpl) FindCreatorByUserID(userID uint) (*accountmodel.Creator, error) {
	creator, err := cs.creatorRepo.FindCreatorByUserID(userID)
	if err != nil {
		log.Printf("Erro ao buscar creator: %s", err)
		return nil, errors.New("criador não encontrado")
	}

	return creator, nil
}

func (cs *creatorServiceImpl) FindCreatorByEmail(email string) (*accountmodel.Creator, error) {
	creator, err := cs.creatorRepo.FindCreatorByUserEmail(email)
	if err != nil {
		log.Printf("Erro ao buscar creator: %s", err)
		return nil, errors.New("criador não encontrado")
	}

	return creator, nil
}

func (cs *creatorServiceImpl) FindByID(id uint) (*accountmodel.Creator, error) {
	creator, err := cs.creatorRepo.FindByID(id)
	if err != nil {
		log.Printf("Erro ao buscar creator por ID %d: %v", id, err)
		return nil, errors.New("criador não encontrado")
	}

	return creator, nil
}

func (cs *creatorServiceImpl) FindByPublicID(publicID string) (*accountmodel.Creator, error) {
	creator, err := cs.creatorRepo.FindByPublicID(publicID)
	if err != nil {
		log.Printf("Erro ao buscar creator por PublicID %s: %v", publicID, err)
		return nil, errors.New("criador não encontrado")
	}
	return creator, nil
}

func (cs *creatorServiceImpl) UpdateCreator(creator *accountmodel.Creator) error {
	log.Printf("Salvando status atualizado para creator %s: onboardingCompleted=%v, chargesEnabled=%v, payoutsEnabled=%v, link=%s",
		creator.Email, creator.OnboardingCompleted, creator.ChargesEnabled, creator.PayoutsEnabled, creator.OnboardingRefreshURL)
	return cs.creatorRepo.Update(creator)
}

// validateReceita validates CPF with Receita Federal and returns the validated name
func (cs *creatorServiceImpl) validateReceita(name, cpf string, birthDate time.Time) (string, error) {
	if cs.rfService == nil {
		return "", errors.New("serviço da receita federal não está disponível")
	}

	response, err := cs.rfService.ConsultaCPF(name, cpf, birthDate.Format("02/01/2006"))
	if err != nil {
		return "", err
	}

	if !response.Status {
		return "", errors.New("CPF inválido ou não encontrado na receita federal")
	}

	if response.Result.NomeDaPF == "" {
		return "", errors.New("nome não encontrado na receita federal")
	}

	return response.Result.NomeDaPF, nil
}
