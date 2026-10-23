package model

import (
	"strings"
	"time"

	"github.com/Assinee-Inc/docffy/internal/subscription/model"
	"github.com/Assinee-Inc/docffy/pkg/utils"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	PublicID        string              `json:"public_id" gorm:"type:varchar(40);uniqueIndex"`
	Username        string              `json:"username" validate:"required"`
	Email           string              `json:"email" validate:"required,email" gorm:"uniqueIndex"`
	GoogleID        string              `json:"google_id" gorm:"type:varchar(128);uniqueIndex"`
	EmailVerifiedAt *time.Time          `json:"email_verified_at"`
	TermsAcceptedAt *time.Time          `json:"terms_accepted_at"`
	Subscription    *model.Subscription `json:"subscription,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.PublicID == "" {
		u.PublicID = utils.GeneratePublicID("usr_")
	}
	return nil
}

func NewGoogleUser(username, email, googleID string) *User {
	now := time.Now()
	return &User{
		Username:        strings.TrimSpace(username),
		Email:           strings.TrimSpace(strings.ToLower(email)),
		GoogleID:        googleID,
		EmailVerifiedAt: &now,
	}
}

// NewUser cria uma instância de User compatível com testes existentes ignorando qualquer parâmetro legado de senha.
func NewUser(username string, args ...string) *User {
	now := time.Now()
	var email, googleID string
	if len(args) == 1 {
		email = args[0]
	} else if len(args) >= 2 {
		// Compatibilidade com chamadas legadas NewUser(username, password, email)
		email = args[1]
		if len(args) >= 3 {
			googleID = args[2]
		}
	}
	return &User{
		Username:        strings.TrimSpace(username),
		Email:           strings.TrimSpace(strings.ToLower(email)),
		GoogleID:        googleID,
		EmailVerifiedAt: &now,
		TermsAcceptedAt: &now,
	}
}

func (u *User) GetInitials() string {
	words := strings.Fields(u.Username)
	if len(words) == 0 {
		return ""
	}
	initials := strings.ToUpper(string(words[0][0]))
	if len(words) > 1 {
		lastWord := words[len(words)-1]
		if len(lastWord) > 0 {
			initials += strings.ToUpper(string(lastWord[0]))
		}
	}
	return initials
}

func (u *User) IsInTrialPeriod() bool {
	if u.Subscription == nil {
		return false
	}
	return u.Subscription.IsInTrialPeriod()
}

func (u *User) DaysLeftInTrial() int {
	if u.Subscription == nil {
		return 0
	}
	return u.Subscription.DaysLeftInTrial()
}

func (u *User) IsSubscribed() bool {
	if u.Subscription == nil {
		return false
	}
	return u.Subscription.IsSubscribed()
}

func (u *User) HasAcceptedTerms() bool {
	return u.TermsAcceptedAt != nil
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

func (u *User) IsConfirmed() bool {
	return u.EmailVerifiedAt != nil
}

// GetSubscriptionStatus returns the subscription status for the user
func (u *User) GetSubscriptionStatus() string {
	if u.Subscription == nil {
		return "inactive"
	}
	return u.Subscription.GetSubscriptionStatus()
}

// DaysLeftInSubscription returns days left in subscription
func (u *User) DaysLeftInSubscription() int {
	if u.Subscription == nil {
		return 0
	}
	return u.Subscription.DaysLeftInSubscription()
}

// IsExpiringSoon returns true if subscription expires in 10 days or less
func (u *User) IsExpiringSoon() bool {
	if u.Subscription == nil {
		return false
	}
	return u.Subscription.IsExpiringSoon()
}
