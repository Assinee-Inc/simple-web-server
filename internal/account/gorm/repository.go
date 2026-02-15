package gorm

import (
	"github.com/anglesson/simple-web-server/internal/account"
	"github.com/google/uuid"
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

	userID, err := uuid.Parse(accountData.UserID)
	if err != nil {
		return err
	}

	modelAccount := &AccountModel{
		Name:                 accountData.Name,
		CPF:                  accountData.CPF,
		Email:                accountData.Email,
		Phone:                accountData.Phone,
		BirthDate:            accountData.BirthDate,
		UserID:               userID,
		Origin:               accountData.Origin,
		ExternalAccountID:    accountData.ExternalAccountID,
		OnboardingCompleted:  accountData.OnboardingCompleted,
		OnboardingRefreshURL: accountData.OnboardingRefreshURL,
		OnboardingReturnURL:  accountData.OnboardingReturnURL,
		PayoutsEnabled:       accountData.PayoutsEnabled,
		ChargesEnabled:       accountData.ChargesEnabled,
	}

	accountData = modelAccount.ToDomain()

	return r.db.Create(modelAccount).Error
}
