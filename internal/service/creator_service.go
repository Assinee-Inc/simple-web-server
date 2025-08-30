package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/pkg/gov"
)

type CreatorService interface {
	CreateCreator(input models.InputCreateCreator) (*models.Creator, error)
	FindCreatorByEmail(email string) (*models.Creator, error)
	FindCreatorByUserID(userID uint) (*models.Creator, error)
	FindByID(id uint) (*models.Creator, error)
	UpdateCreator(creator *models.Creator) error
}

type creatorServiceImpl struct {
	creatorRepo         repository.CreatorRepository
	rfService           gov.ReceitaFederalService
	userService         UserService
	subscriptionService SubscriptionService
	paymentGateway      PaymentGateway
}

func NewCreatorService(
	creatorRepo repository.CreatorRepository,
	receitaFederalService gov.ReceitaFederalService,
	userService UserService,
	subscriptionService SubscriptionService,
	paymentGateway PaymentGateway,
) CreatorService {
	return &creatorServiceImpl{
		creatorRepo:         creatorRepo,
		rfService:           receitaFederalService,
		userService:         userService,
		subscriptionService: subscriptionService,
		paymentGateway:      paymentGateway,
	}
}

func (cs *creatorServiceImpl) CreateCreator(input models.InputCreateCreator) (*models.Creator, error) {
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
	if config.AppConfig.AppMode == "PRODUCTION" {
		validatedName, err = cs.validateReceita(cleanCPF, birthDate)
		if err != nil {
			return nil, err
		}
	}

	// Create user
	inputCreateUser := models.InputCreateUser{
		Username:             validatedName,
		Email:                input.Email,
		Password:             input.Password,
		PasswordConfirmation: input.PasswordConfirmation,
	}

	user, err := cs.userService.CreateUser(inputCreateUser)
	if err != nil {
		return nil, err
	}

	// Create creator
	creator := models.NewCreator(
		validatedName,
		input.Email,
		cleanPhone(input.PhoneNumber),
		cleanCPF,
		birthDate,
		user.ID,
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
		// The customer can be created later
	} else {
		// Create subscription for the creator
		subscription, err := cs.subscriptionService.CreateSubscription(user.ID, "default_plan")
		if err != nil {
			log.Printf("Error creating subscription: %v", err)
		} else {
			// Activate subscription with customer ID
			err = cs.subscriptionService.ActivateSubscription(subscription, customerID, "")
			if err != nil {
				log.Printf("Error activating subscription: %v", err)
			}
		}
	}

	return creator, nil
}

func (cs *creatorServiceImpl) FindCreatorByUserID(userID uint) (*models.Creator, error) {
	creator, err := cs.creatorRepo.FindCreatorByUserID(userID)
	if err != nil {
		log.Printf("Erro ao buscar creator: %s", err)
		return nil, errors.New("criador não encontrado")
	}

	log.Printf("Infoprodutor encontrado! Nome: %v", creator.Name)

	return creator, nil
}

func (cs *creatorServiceImpl) FindCreatorByEmail(email string) (*models.Creator, error) {
	creator, err := cs.creatorRepo.FindCreatorByUserEmail(email)
	if err != nil {
		log.Printf("Erro ao buscar creator: %s", err)
		return nil, errors.New("criador não encontrado")
	}

	log.Printf("Infoprodutor encontrado! Email: %v", creator.Email)

	return creator, nil
}

func (cs *creatorServiceImpl) FindByID(id uint) (*models.Creator, error) {
	// Buscar criador pelo ID
	creator, err := cs.creatorRepo.FindByID(id)
	if err != nil {
		log.Printf("Erro ao buscar creator por ID %d: %v", id, err)
		return nil, errors.New("criador não encontrado")
	}

	return creator, nil
}

func (cs *creatorServiceImpl) UpdateCreator(creator *models.Creator) error {
	return cs.creatorRepo.Update(creator)
}

// validateReceita validates CPF with Receita Federal and returns the validated name
func (cs *creatorServiceImpl) validateReceita(cpf string, birthDate time.Time) (string, error) {
	if cs.rfService == nil {
		return "", errors.New("serviço da receita federal não está disponível")
	}

	response, err := cs.rfService.ConsultaCPF(cpf, birthDate.Format("02/01/2006"))
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
