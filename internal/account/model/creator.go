package model

import (
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/pkg/utils"
	"gorm.io/gorm"
)

type Creator struct {
	gorm.Model
	PublicID               string    `json:"public_id" gorm:"type:varchar(40);uniqueIndex"`
	Name                   string    `json:"name"`
	SocialName             string    `json:"social_name"`
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

func NewCreator(name, socialName, email, phone, cpf string, birthDate time.Time, userID uint) *Creator {
	return &Creator{
		Name:       name,
		SocialName: socialName,
		Email:      email,
		Phone:      phone,
		CPF:        cpf,
		BirthDate:  birthDate,
		UserID:     userID,
	}
}

func (c *Creator) GetDisplayName() string {
	if c.SocialName != "" {
		return c.SocialName
	}
	parts := strings.Fields(c.Name)
	if len(parts) <= 1 {
		return c.Name
	}
	return parts[0] + " " + parts[len(parts)-1]
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

func (c *Creator) BeforeCreate(tx *gorm.DB) error {
	if c.PublicID == "" {
		c.PublicID = utils.GeneratePublicID("crt_")
	}
	return nil
}

func (c *Creator) NeedsOnboarding() bool {
	return c.StripeConnectAccountID != "" && (!c.OnboardingCompleted || !c.ChargesEnabled || !c.PayoutsEnabled)
}
