package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Assinee-Inc/docffy/internal/auth/model"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	"github.com/Assinee-Inc/docffy/internal/config"
)

var (
	ErrUserAlreadyExists  = errors.New("usuário já existe")
	ErrInvalidCredentials = errors.New("credencial inválida")
	ErrUserNotFound       = errors.New("usuário não encontrado")
)

type UserService interface {
	AuthenticateWithGoogle(ctx context.Context, idToken string) (*model.User, bool, error)
	FindByID(id uint) *model.User
	FindByEmail(email string) *model.User
	Update(user *model.User) error
	CreateUser(input InputCreateUser) (uint, error)
	GenerateVerificationToken(email string) (string, error)
}

type UserServiceImpl struct {
	userRepository authrepo.UserRepository
	googleVerifier GoogleTokenVerifier
}

func NewUserService(userRepository authrepo.UserRepository, deps ...any) UserService {
	var verifier GoogleTokenVerifier
	for _, dep := range deps {
		if v, ok := dep.(GoogleTokenVerifier); ok {
			verifier = v
		}
	}
	if verifier == nil {
		verifier = NewGoogleOIDCVerifier(config.AppConfig.GoogleClientID)
	}
	return &UserServiceImpl{
		userRepository: userRepository,
		googleVerifier: verifier,
	}
}

func (us *UserServiceImpl) AuthenticateWithGoogle(ctx context.Context, idToken string) (*model.User, bool, error) {
	if strings.TrimSpace(idToken) == "" {
		return nil, false, ErrInvalidGoogleToken
	}

	identity, err := us.googleVerifier.VerifyIDToken(ctx, idToken)
	if err != nil || identity == nil || !identity.EmailVerified {
		return nil, false, ErrInvalidGoogleToken
	}

	email := strings.TrimSpace(strings.ToLower(identity.Email))
	sub := strings.TrimSpace(identity.Sub)
	if email == "" || sub == "" {
		return nil, false, ErrInvalidGoogleToken
	}

	// 1. Busca pelo GoogleID
	user, err := us.userRepository.FindByGoogleID(sub)
	if err != nil {
		return nil, false, err
	}
	if user != nil {
		if user.EmailVerifiedAt == nil {
			now := time.Now()
			user.EmailVerifiedAt = &now
			_ = us.userRepository.Update(user)
		}
		return user, false, nil
	}

	// 2. Busca por e-mail existente para vínculo automático sem senha
	existingByEmail := us.userRepository.FindByEmail(email)
	if existingByEmail != nil {
		existingByEmail.GoogleID = sub
		if existingByEmail.EmailVerifiedAt == nil {
			now := time.Now()
			existingByEmail.EmailVerifiedAt = &now
		}
		if strings.TrimSpace(existingByEmail.Username) == "" && strings.TrimSpace(identity.Name) != "" {
			existingByEmail.Username = strings.TrimSpace(identity.Name)
		}
		if err := us.userRepository.Update(existingByEmail); err != nil {
			return nil, false, err
		}
		return existingByEmail, false, nil
	}

	// 3. Novo usuário (Etapa 1 do onboarding)
	username := strings.TrimSpace(identity.Name)
	if username == "" {
		username = strings.Split(email, "@")[0]
	}

	newUser := model.NewGoogleUser(username, email, sub)
	if err := us.userRepository.Create(newUser); err != nil {
		return nil, false, err
	}

	return newUser, true, nil
}

func (us *UserServiceImpl) FindByID(id uint) *model.User {
	return us.userRepository.FindByID(id)
}

func (us *UserServiceImpl) FindByEmail(email string) *model.User {
	return us.userRepository.FindByUserEmail(email)
}

func (us *UserServiceImpl) Update(user *model.User) error {
	return us.userRepository.Update(user)
}

// CreateUser mantido apenas para compatibilidade temporária antes da Fase 4.
func (us *UserServiceImpl) CreateUser(input InputCreateUser) (uint, error) {
	if err := validateUserInput(input); err != nil {
		return 0, err
	}

	input.Username = strings.TrimSpace(input.Username)
	existingUser := us.userRepository.FindByUserEmail(input.Email)
	if existingUser != nil {
		return 0, ErrUserAlreadyExists
	}

	user := model.NewGoogleUser(input.Username, input.Email, "")
	if err := us.userRepository.Create(user); err != nil {
		return 0, err
	}
	return user.ID, nil
}

// GenerateVerificationToken mantido apenas como no-op compatível até a Fase 4.
func (us *UserServiceImpl) GenerateVerificationToken(email string) (string, error) {
	return "", nil
}
