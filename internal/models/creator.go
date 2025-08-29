package models

import (
	"time"

	"gorm.io/gorm"
)

type Creator struct {
	gorm.Model
	Name                    string    `json:"name"`
	CPF                     string    `json:"cpf"`
	Email                   string    `json:"email"`
	Phone                   string    `json:"phone"`
	BirthDate               time.Time `json:"birth_date"`
	UserID                  uint      `json:"user_id"`
	User                    User      `gorm:"foreignKey:UserID"`
	Ebooks                  []Ebook
	Clients                 []*Client `gorm:"many2many:client_creators"`
	StripeConnectAccountID  string    `json:"stripe_connect_account_id"`
	OnboardingCompleted     bool      `json:"onboarding_completed" gorm:"default:false"`
	PayoutsEnabled          bool      `json:"payouts_enabled" gorm:"default:false"`
	ChargesEnabled          bool      `json:"charges_enabled" gorm:"default:false"`
	OnboardingRefreshURL    string    `json:"onboarding_refresh_url"`
	OnboardingReturnURL     string    `json:"onboarding_return_url"`
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

func (c *Creator) GetEbooks() []Ebook {
	return c.Ebooks
}

func (c *Creator) AddClient(client *Client) {
	c.Clients = append(c.Clients, client)
}

// IsAdult returns true if the creator is 18 years or older
func (c *Creator) IsAdult() bool {
	now := time.Now()
	age := now.Year() - c.BirthDate.Year()

	// Adjust age if birthday hasn't occurred yet this year
	if now.Month() < c.BirthDate.Month() || (now.Month() == c.BirthDate.Month() && now.Day() < c.BirthDate.Day()) {
		age--
	}

	return age >= 18
}
