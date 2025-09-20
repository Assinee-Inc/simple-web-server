package repository

import (
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/anglesson/simple-web-server/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	Save(user *models.User) error
	FindByUserEmail(emailUser string) *models.User
	FindByEmail(emailUser string) *models.User
	FindBySessionToken(token string) *models.User
	FindByStripeCustomerID(customerID string) *models.User
	FindByPasswordResetToken(token string) *models.User
	UpdatePasswordResetToken(user *models.User, token string) error
}

type GormUserRepositoryImpl struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepositoryImpl {
	return &GormUserRepositoryImpl{
		db: db,
	}
}

func (r *GormUserRepositoryImpl) Save(user *models.User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		log.Printf("Erro ao salvar usuário: %v", result.Error)
		return result.Error
	}

	log.Printf("Usuário atualizado com sucesso. ID: %d, EMAIL: %s", user.ID, user.Email)
	return nil
}

// TODO: add error handler
func (r *GormUserRepositoryImpl) FindByEmail(emailUser string) *models.User {
	var user models.User
	result := r.db.Preload("Subscription").Where("email = ?", emailUser).First(&user)
	if result.Error != nil {
		log.Printf("Erro ao buscar usuário por email: %v", result.Error)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) FindBySessionToken(token string) *models.User {
	var user models.User
	result := r.db.Preload("Subscription").Where("session_token = ?", token).First(&user)
	if result.Error != nil {
		log.Printf("Erro ao buscar usuário por session token: %v", result.Error)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) FindByStripeCustomerID(customerID string) *models.User {
	var user models.User
	err := r.db.Where("stripe_customer_id = ?", customerID).First(&user).Error
	if err != nil {
		log.Printf("Error finding user by Stripe customer ID: %v", err)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) Create(user *models.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}
	return nil
}

func (r *GormUserRepositoryImpl) FindByUserEmail(emailUser string) *models.User {
	var user models.User
	err := r.db.Preload("Subscription").Where("email = ?", emailUser).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		log.Printf("Error finding user by email: %v", err)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) FindByPasswordResetToken(token string) *models.User {
	var user models.User
	err := r.db.Preload("Subscription").Where("password_reset_token = ?", token).First(&user).Error
	if err != nil {
		log.Printf("Error finding user by password reset token: %v", err)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) UpdatePasswordResetToken(user *models.User, token string) error {
	now := time.Now()
	user.PasswordResetToken = token
	user.PasswordResetAt = &now
	return r.Save(user)
}
