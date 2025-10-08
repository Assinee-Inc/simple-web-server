package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/pkg/utils"
)

var ErrUserAlreadyExists = errors.New("usuário já existe")
var ErrInvalidCredentials = errors.New("email ou senha inválidos")
var ErrUserNotFound = errors.New("usuário não encontrado")
var ErrInvalidResetToken = errors.New("token de reset inválido ou expirado")

type UserService interface {
	CreateUser(input models.InputCreateUser) (*models.User, error)
	AuthenticateUser(input models.InputLogin) (*models.User, error)
	RequestPasswordReset(email string) error
	ResetPassword(token, newPassword string) error
	FindByEmail(email string) *models.User
}

type UserServiceImpl struct {
	userRepository repository.UserRepository
	encrypter      utils.Encrypter
}

func NewUserService(userRepository repository.UserRepository, encrypter utils.Encrypter) UserService {
	return &UserServiceImpl{
		userRepository: userRepository,
		encrypter:      encrypter,
	}
}

func (us *UserServiceImpl) CreateUser(input models.InputCreateUser) (*models.User, error) {
	// Validate input
	if err := validateUserInput(input); err != nil {
		return nil, err
	}

	// Clean username
	input.Username = strings.TrimSpace(input.Username)

	existingUser := us.userRepository.FindByUserEmail(input.Email)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword := us.encrypter.HashPassword(input.Password)
	user := models.NewUser(input.Username, hashedPassword, input.Email)

	err := us.userRepository.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (us *UserServiceImpl) AuthenticateUser(input models.InputLogin) (*models.User, error) {
	// Validate input
	if input.Email == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	// Find user by email
	user := us.userRepository.FindByUserEmail(input.Email)
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check password
	if !us.encrypter.CheckPasswordHash(user.Password, input.Password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (us *UserServiceImpl) RequestPasswordReset(email string) error {
	user := us.userRepository.FindByUserEmail(email)
	if user == nil {
		// Não retornamos erro para não revelar se o email existe ou não
		return nil
	}

	// Gerar token único
	token, err := generateResetToken()
	if err != nil {
		return err
	}

	// Salvar token no usuário
	err = us.userRepository.UpdatePasswordResetToken(user, token)
	if err != nil {
		return err
	}

	return nil
}

func (us *UserServiceImpl) ResetPassword(token, newPassword string) error {
	user := us.userRepository.FindByPasswordResetToken(token)
	if user == nil {
		return ErrInvalidResetToken
	}

	// Verificar se o token não expirou (24 horas)
	if user.PasswordResetAt != nil && time.Since(*user.PasswordResetAt) > 24*time.Hour {
		return ErrInvalidResetToken
	}

	// Hash da nova senha
	hashedPassword := us.encrypter.HashPassword(newPassword)
	user.Password = hashedPassword
	user.PasswordResetToken = ""
	user.PasswordResetAt = nil

	// Salvar usuário
	err := us.userRepository.Save(user)
	if err != nil {
		return err
	}

	return nil
}

func (us *UserServiceImpl) FindByEmail(email string) *models.User {
	return us.userRepository.FindByUserEmail(email)
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
