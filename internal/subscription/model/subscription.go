package model

import (
	"math"
	"time"

	"github.com/anglesson/simple-web-server/pkg/utils"
	"gorm.io/gorm"
)

type Subscription struct {
	gorm.Model
	PublicID             string     `json:"public_id" gorm:"type:varchar(40);uniqueIndex"`
	UserID               uint       `json:"user_id" gorm:"not null"`
	PlanID               string     `json:"plan_id"`
	TrialStartDate       time.Time  `json:"trial_start_date"`
	TrialEndDate         time.Time  `json:"trial_end_date"`
	IsTrialActive        bool       `json:"is_trial_active" gorm:"default:true"`
	StripeCustomerID     string     `json:"stripe_customer_id"`
	StripeSubscriptionID string     `json:"stripe_subscription_id"`
	SubscriptionStatus   string     `json:"subscription_status" gorm:"default:'inactive'"`
	SubscriptionEndDate  *time.Time `json:"subscription_end_date"`
	Origin               string     `json:"origin" gorm:"default:'web'"`
}

func NewSubscription(userID uint, planID string) *Subscription {
	now := time.Now()
	trialEndDate := now.AddDate(0, 0, 7) // 7 days trial

	return &Subscription{
		UserID:         userID,
		PlanID:         planID,
		TrialStartDate: now,
		TrialEndDate:   trialEndDate,
		IsTrialActive:  true,
		Origin:         "web",
	}
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.PublicID == "" {
		s.PublicID = utils.GeneratePublicID("sub_")
	}
	return nil
}

func (s *Subscription) IsInTrialPeriod() bool {
	if s.TrialEndDate.IsZero() {
		return false
	}
	return s.IsTrialActive && time.Now().Before(s.TrialEndDate)
}

func (s *Subscription) DaysLeftInTrial() int {
	if s.TrialEndDate.IsZero() {
		return 0
	}
	days := time.Until(s.TrialEndDate).Hours() / 24
	if math.Ceil(days) < 0 {
		return 0
	}
	return int(math.Ceil(days))
}

func (s *Subscription) IsSubscribed() bool {
	if s.SubscriptionStatus == "active" {
		return true
	}
	if s.SubscriptionEndDate != nil && time.Now().Before(*s.SubscriptionEndDate) {
		return true
	}
	return false
}

func (s *Subscription) CancelSubscription() {
	s.SubscriptionStatus = "canceled"
	s.UpdatedAt = time.Now()
}

func (s *Subscription) EndTrial() {
	s.IsTrialActive = false
	s.UpdatedAt = time.Now()
}

func (s *Subscription) ActivateSubscription(stripeCustomerID, stripeSubscriptionID string) {
	s.StripeCustomerID = stripeCustomerID
	s.StripeSubscriptionID = stripeSubscriptionID
	s.SubscriptionStatus = "active"
	s.UpdatedAt = time.Now()
}

func (s *Subscription) UpdateSubscriptionStatus(status string, endDate *time.Time) {
	s.SubscriptionStatus = status
	s.SubscriptionEndDate = endDate
	s.UpdatedAt = time.Now()
}

// DaysLeftInSubscription returns the number of days left in the subscription
func (s *Subscription) DaysLeftInSubscription() int {
	if s.SubscriptionEndDate == nil {
		return 0
	}
	days := time.Until(*s.SubscriptionEndDate).Hours() / 24
	if math.Ceil(days) < 0 {
		return 0
	}
	return int(math.Ceil(days))
}

// IsExpiringSoon returns true if subscription expires in 10 days or less
func (s *Subscription) IsExpiringSoon() bool {
	daysLeft := s.DaysLeftInSubscription()
	return daysLeft > 0 && daysLeft <= 10
}

// GetSubscriptionStatus returns a string describing the current subscription status
func (s *Subscription) GetSubscriptionStatus() string {
	if s.IsInTrialPeriod() {
		return "trial"
	}
	if s.IsSubscribed() {
		if s.IsExpiringSoon() {
			return "expiring"
		}
		return "active"
	}
	return "inactive"
}
