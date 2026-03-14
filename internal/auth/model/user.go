package model

import (
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/internal/subscription/model"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	PublicID           string `json:"public_id" gorm:"type:varchar(40);uniqueIndex"`
	Username           string `json:"username" validate:"required"`
	Password           string `json:"password" validate:"required"`
	Email              string `json:"email" validate:"required,email" gorm:"unique"`
	CSRFToken          string
	SessionToken       string
	PasswordResetToken    string
	PasswordResetAt       *time.Time
	EmailVerifiedAt       *time.Time
	EmailVerificationToken string
	TermsAcceptedAt       *time.Time `json:"terms_accepted_at"`
	Subscription       *model.Subscription `json:"subscription" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.PublicID == "" {
		u.PublicID = utils.GeneratePublicID("usr_")
	}
	return nil
}

func NewUser(username, password, email string) *User {
	now := time.Now()
	return &User{
		Username:        username,
		Password:        password,
		Email:           email,
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
