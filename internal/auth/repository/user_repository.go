package repository

import (
	"errors"
	"log"

	"gorm.io/gorm"

	"github.com/Assinee-Inc/docffy/internal/auth/model"
)

type UserRepository interface {
	Create(user *model.User) error
	Save(user *model.User) error
	Update(user *model.User) error
	Delete(id uint) error
	FindByID(id uint) *model.User
	FindByGoogleID(googleID string) (*model.User, error)
	FindByUserEmail(emailUser string) *model.User
	FindByEmail(emailUser string) *model.User
	FindBySessionToken(token string) *model.User
	FindByStripeCustomerID(customerID string) *model.User
	FindByPublicID(publicID string) *model.User
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

func (r *GormUserRepositoryImpl) Update(user *model.User) error {
	return r.Save(user)
}

func (r *GormUserRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *GormUserRepositoryImpl) FindByID(id uint) *model.User {
	if id == 0 {
		return nil
	}
	var user model.User
	err := r.db.Preload("Subscription").First(&user, id).Error
	if err != nil {
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) FindByGoogleID(googleID string) (*model.User, error) {
	if googleID == "" {
		return nil, nil
	}
	var user model.User
	err := r.db.Preload("Subscription").Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("Error finding user by Google ID: %v", err)
		return nil, err
	}
	return &user, nil
}

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
	return nil
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		log.Printf("Error finding user by email: %v", err)
		return nil
	}
	return &user
}

func (r *GormUserRepositoryImpl) FindByPublicID(publicID string) *model.User {
	var user model.User
	err := r.db.Preload("Subscription").Where("public_id = ?", publicID).First(&user).Error
	if err != nil {
		log.Printf("Error finding user by public ID: %v", err)
		return nil
	}
	return &user
}
