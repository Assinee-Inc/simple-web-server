package gorm

import (
	"time"

	"github.com/anglesson/simple-web-server/internal/account"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *Model) BeforeCreate() error {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	m.ID = newUUID
	return nil
}

type AccountModel struct {
	Model
	Name                 string    `json:"name"`
	CPF                  string    `json:"cpf"`
	Email                string    `json:"email"`
	Phone                string    `json:"phone"`
	BirthDate            time.Time `json:"birth_date"`
	UserID               uuid.UUID `json:"user_id"`
	Origin               string    `json:"origin"`
	ExternalAccountID    string    `json:"external_account_id"`
	OnboardingCompleted  bool      `json:"onboarding_completed" gorm:"default:false"`
	PayoutsEnabled       bool      `json:"payouts_enabled" gorm:"default:false"`
	ChargesEnabled       bool      `json:"charges_enabled" gorm:"default:false"`
	OnboardingRefreshURL string    `json:"onboarding_refresh_url"`
	OnboardingReturnURL  string    `json:"onboarding_return_url"`
}

func (a *AccountModel) ToDomain() *account.Account {
	return &account.Account{
		ID:                   string(a.ID.String()),
		Name:                 a.Name,
		CPF:                  a.CPF,
		Email:                a.Email,
		Phone:                a.Phone,
		BirthDate:            a.BirthDate,
		UserID:               string(a.UserID.String()),
		Origin:               a.Origin,
		ExternalAccountID:    a.ExternalAccountID,
		OnboardingCompleted:  a.OnboardingCompleted,
		PayoutsEnabled:       a.PayoutsEnabled,
		ChargesEnabled:       a.ChargesEnabled,
		OnboardingRefreshURL: a.OnboardingRefreshURL,
		OnboardingReturnURL:  a.OnboardingReturnURL,
	}
}
