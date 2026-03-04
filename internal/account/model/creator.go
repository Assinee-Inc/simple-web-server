package model

import (
	"time"

	"gorm.io/gorm"
)

type Creator struct {
	gorm.Model
	Name                   string    `json:"name"`
	CPF                    string    `json:"cpf"`
	Email                  string    `json:"email"`
	Phone                  string    `json:"phone"`
	BirthDate              time.Time `json:"birth_date"`
	UserID                 uint      `json:"user_id"`
	StripeConnectAccountID string    `json:"stripe_connect_account_id"`
	OnboardingCompleted    bool      `json:"onboarding_completed" gorm:"default:false"`
	PayoutsEnabled         bool      `json:"payouts_enabled" gorm:"default:false"`
	ChargesEnabled         bool      `json:"charges_enabled" gorm:"default:false"`
	OnboardingRefreshURL   string    `json:"onboarding_refresh_url"`
	OnboardingReturnURL    string    `json:"onboarding_return_url"`
}

func NewCreator(name, email, phone, cpf string, birthDate time.Time, userID uint) *Creator {
	return &Creator{
		Name:      name,
		Email:     email,
		Phone:     phone,
		CPF:       cpf,
		BirthDate: birthDate,
		UserID:    userID,
	}
}

// IsAdult returns true if the creator is 18 years or older
func (c *Creator) IsAdult() bool {
	now := time.Now()
	age := now.Year() - c.BirthDate.Year()
	if now.Month() < c.BirthDate.Month() || (now.Month() == c.BirthDate.Month() && now.Day() < c.BirthDate.Day()) {
		age--
	}
	return age >= 18
}

func (c *Creator) NeedsOnboarding() bool {
	return c.StripeConnectAccountID != "" && (!c.OnboardingCompleted || !c.ChargesEnabled || !c.PayoutsEnabled)
}
