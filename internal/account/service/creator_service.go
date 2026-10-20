package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountrepo "github.com/Assinee-Inc/docffy/internal/account/repository"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	subscriptionservice "github.com/Assinee-Inc/docffy/internal/subscription/service"
	"github.com/Assinee-Inc/docffy/pkg/gov"
)

var (
	ErrCreatorAlreadyExists = errors.New("criador já existe")
	ErrCreatorNotFound      = errors.New("criador não encontrado")
)

type CreatorService interface {
	OnboardCreator(userID uint, input InputOnboardingCreator) (*accountmodel.Creator, error)
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

func (cs *creatorServiceImpl) OnboardCreator(userID uint, input InputOnboardingCreator) (*accountmodel.Creator, error) {
	if userID == 0 {
		return nil, errors.New("usuário não autenticado")
	}

	if err := validateOnboardingInput(input); err != nil {
		return nil, err
	}

	birthDate, err := parseBirthDate(input.BirthDate)
	if err != nil {
		return nil, err
	}

	user := cs.userService.FindByID(userID)
	if user == nil {
		return nil, authsvc.ErrUserNotFound
	}

	existingForUser, _ := cs.creatorRepo.FindCreatorByUserID(userID)
	if existingForUser != nil {
		return nil, ErrCreatorAlreadyExists
	}

	cleanCPFStr := cleanCPF(input.CPF)
	creatorExists, err := cs.creatorRepo.FindByCPF(cleanCPFStr)
	if err != nil {
		return nil, err
	}
	if creatorExists != nil {
		return nil, ErrCreatorAlreadyExists
	}

	validatedName := strings.TrimSpace(input.Name)
	if cs.rfService != nil {
		validatedName, err = cs.validateReceita(validatedName, cleanCPFStr, birthDate)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	user.TermsAcceptedAt = &now
	if err := cs.userService.Update(user); err != nil {
		return nil, err
	}

	creator := accountmodel.NewCreator(
		validatedName,
		strings.TrimSpace(input.SocialName),
		user.Email,
		cleanPhone(input.PhoneNumber),
		cleanCPFStr,
		birthDate,
		userID,
	)

	if err := cs.creatorRepo.Create(creator); err != nil {
		return nil, err
	}

	cs.provisionSubscription(userID, user.Email, validatedName)

	return creator, nil
}

func (cs *creatorServiceImpl) CreateCreator(input InputCreateCreator) (*accountmodel.Creator, error) {
	if err := validateCreatorInput(input); err != nil {
		return nil, err
	}

	birthDate, err := parseBirthDate(input.BirthDate)
	if err != nil {
		return nil, err
	}

	cleanCPFStr := cleanCPF(input.CPF)
	creatorExists, err := cs.creatorRepo.FindByCPF(cleanCPFStr)
	if err != nil {
		return nil, err
	}
	if creatorExists != nil {
		return nil, ErrCreatorAlreadyExists
	}

	validatedName := input.Name
	if cs.rfService != nil {
		validatedName, err = cs.validateReceita(validatedName, cleanCPFStr, birthDate)
		if err != nil {
			return nil, err
		}
	}

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

	creator := accountmodel.NewCreator(
		validatedName,
		input.SocialName,
		input.Email,
		cleanPhone(input.PhoneNumber),
		cleanCPFStr,
		birthDate,
		userID,
	)

	if err := cs.creatorRepo.Create(creator); err != nil {
		return nil, err
	}

	cs.provisionSubscription(userID, input.Email, validatedName)

	return creator, nil
}

func (cs *creatorServiceImpl) provisionSubscription(userID uint, email, name string) {
	if cs.paymentGateway == nil || cs.subscriptionService == nil {
		return
	}
	customerID, err := cs.paymentGateway.CreateCustomer(email, name)
	if err != nil {
		log.Printf("Error creating customer in payment gateway: %v", err)
		return
	}
	subscriptionID, err := cs.subscriptionService.CreateSubscription(userID, "default_plan")
	if err != nil {
		log.Printf("Error creating subscription: %v", err)
		return
	}
	if err := cs.subscriptionService.ActivateSubscription(subscriptionID, customerID, ""); err != nil {
		log.Printf("Error activating subscription: %v", err)
	}
}

func parseBirthDate(birthDateStr string) (time.Time, error) {
	birthDate, err := time.Parse("02/01/2006", birthDateStr)
	if err != nil {
		birthDate, err = time.Parse("2006-01-02", birthDateStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("formato de data de nascimento inválido: %w", err)
		}
	}
	return birthDate, nil
}

func (cs *creatorServiceImpl) FindCreatorByUserID(userID uint) (*accountmodel.Creator, error) {
	creator, err := cs.creatorRepo.FindCreatorByUserID(userID)
	if err != nil {
		log.Printf("Erro ao buscar creator: %s", err)
		return nil, ErrCreatorNotFound
	}

	return creator, nil
}

func (cs *creatorServiceImpl) FindCreatorByEmail(email string) (*accountmodel.Creator, error) {
	creator, err := cs.creatorRepo.FindCreatorByUserEmail(email)
	if err != nil {
		log.Printf("Erro ao buscar creator: %s", err)
		return nil, ErrCreatorNotFound
	}

	return creator, nil
}

func (cs *creatorServiceImpl) FindByID(id uint) (*accountmodel.Creator, error) {
	creator, err := cs.creatorRepo.FindByID(id)
	if err != nil {
		log.Printf("Erro ao buscar creator por ID %d: %v", id, err)
		return nil, ErrCreatorNotFound
	}

	return creator, nil
}

func (cs *creatorServiceImpl) FindByPublicID(publicID string) (*accountmodel.Creator, error) {
	creator, err := cs.creatorRepo.FindByPublicID(publicID)
	if err != nil {
		log.Printf("Erro ao buscar creator por PublicID %s: %v", publicID, err)
		return nil, ErrCreatorNotFound
	}
	return creator, nil
}

func (cs *creatorServiceImpl) UpdateCreator(creator *accountmodel.Creator) error {
	log.Printf("Salvando status atualizado para creator %s: onboardingCompleted=%v, chargesEnabled=%v, payoutsEnabled=%v, link=%s",
		creator.Email, creator.OnboardingCompleted, creator.ChargesEnabled, creator.PayoutsEnabled, creator.OnboardingRefreshURL)
	return cs.creatorRepo.Update(creator)
}

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
