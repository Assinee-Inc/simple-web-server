package repository

import (
	"github.com/anglesson/simple-web-server/internal/account"
	"gorm.io/gorm"
)

type gormAccountRepository struct {
	db *gorm.DB
}

func NewGormAccountRepository(db *gorm.DB) *gormAccountRepository {
	return &gormAccountRepository{db: db}
}
func (r *gormAccountRepository) Create(accountData *account.Account) error {
	if accountData == nil {
		return account.ErrInvalidAccount
	}

	return r.db.Create(accountData).Error
}
