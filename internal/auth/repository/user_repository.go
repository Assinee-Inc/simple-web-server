package repository

import (
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/anglesson/simple-web-server/internal/auth/model"
)

type UserRepository interface {
	Create(user *model.User) error
	Save(user *model.User) error
	FindByUserEmail(emailUser string) *model.User
	FindByEmail(emailUser string) *model.User
	FindBySessionToken(token string) *model.User
	FindByStripeCustomerID(customerID string) *model.User
	FindByPasswordResetToken(token string) *model.User
	UpdatePasswordResetToken(user *model.User, token string) error
}

type GormUserRepositoryImpl struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepositoryImpl {
	return &GormUserRepositoryImpl{
		db: db,
	}
}

func (r *GormUserRepositoryImpl) Save(user *model.User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		log.Printf("Erro ao salvar usuário: %v", result.Error)
		return result.Error
	}

	log.Printf("Usuário atualizado com sucesso. ID: %d, EMAIL: %s", user.ID, user.Email)
	return nil
}

// TODO: add error handler
func (r *GormUserRepositoryImpl) FindByEmail(emailUser string) *model.User {
	var user model.User
	result := r.db.Preload("Subscription").Where("email = ?", emailUser).First(&user)
	if result.Error != nil {
		log.Printf("Erro ao buscar usuário por email: %v", result.Error)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) FindBySessionToken(token string) *model.User {
	var user model.User
	result := r.db.Preload("Subscription").Where("session_token = ?", token).First(&user)
	if result.Error != nil {
		log.Printf("Erro ao buscar usuário por session token: %v", result.Error)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) FindByStripeCustomerID(customerID string) *model.User {
	var user model.User
	err := r.db.Where("stripe_customer_id = ?", customerID).First(&user).Error
	if err != nil {
		log.Printf("Error finding user by Stripe customer ID: %v", err)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) Create(user *model.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}
	return nil
}

func (r *GormUserRepositoryImpl) FindByUserEmail(emailUser string) *model.User {
	var user model.User
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

func (r *GormUserRepositoryImpl) FindByPasswordResetToken(token string) *model.User {
	var user model.User
	err := r.db.Preload("Subscription").Where("password_reset_token = ?", token).First(&user).Error
	if err != nil {
		log.Printf("Error finding user by password reset token: %v", err)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) UpdatePasswordResetToken(user *model.User, token string) error {
	now := time.Now()
	user.PasswordResetToken = token
	user.PasswordResetAt = &now
	return r.Save(user)
}
