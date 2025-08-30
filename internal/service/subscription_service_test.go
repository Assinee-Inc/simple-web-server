package service

import (
	"errors"
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubscriptionService_CreateSubscription(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		planID         string
		setupMocks     func(*mocks.MockSubscriptionRepository)
		expectedError  bool
		expectedResult *models.Subscription
	}{
		{
			name:   "should create subscription successfully",
			userID: 1,
			planID: "test_plan",
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				mockRepo.On("Create", mock.AnythingOfType("*models.Subscription")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:           "should return error when user ID is zero",
			userID:         0,
			planID:         "test_plan",
			setupMocks:     func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name:           "should return error when plan ID is empty",
			userID:         1,
			planID:         "",
			setupMocks:     func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name:   "should return error when repository fails",
			userID: 1,
			planID: "test_plan",
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				mockRepo.On("Create", mock.AnythingOfType("*models.Subscription")).Return(errors.New("database error"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSubscriptionRepository)
			mockRF := new(mocks.MockRFService)
			tt.setupMocks(mockRepo)

			service := NewSubscriptionService(mockRepo, mockRF)

			result, err := service.CreateSubscription(tt.userID, tt.planID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.UserID)
				assert.Equal(t, tt.planID, result.PlanID)
				assert.True(t, result.IsTrialActive)
				assert.Equal(t, "web", result.Origin)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_FindByUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		setupMocks     func(*mocks.MockSubscriptionRepository)
		expectedError  bool
		expectedResult *models.Subscription
	}{
		{
			name:   "should find subscription successfully",
			userID: 1,
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				subscription := &models.Subscription{
					UserID: 1,
					PlanID: "test_plan",
				}
				mockRepo.On("FindByUserID", uint(1)).Return(subscription, nil)
			},
			expectedError: false,
		},
		{
			name:           "should return error when user ID is zero",
			userID:         0,
			setupMocks:     func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name:   "should return error when repository fails",
			userID: 1,
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				mockRepo.On("FindByUserID", uint(1)).Return(nil, errors.New("database error"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSubscriptionRepository)
			mockRF := new(mocks.MockRFService)
			tt.setupMocks(mockRepo)

			service := NewSubscriptionService(mockRepo, mockRF)

			result, err := service.FindByUserID(tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_ActivateSubscription(t *testing.T) {
	tests := []struct {
		name                 string
		subscription         *models.Subscription
		stripeCustomerID     string
		stripeSubscriptionID string
		setupMocks           func(*mocks.MockSubscriptionRepository)
		expectedError        bool
	}{
		{
			name: "should activate subscription successfully",
			subscription: &models.Subscription{
				UserID: 1,
				PlanID: "test_plan",
			},
			stripeCustomerID:     "cus_123",
			stripeSubscriptionID: "sub_123",
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				mockRepo.On("Save", mock.AnythingOfType("*models.Subscription")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:                 "should return error when subscription is nil",
			subscription:         nil,
			stripeCustomerID:     "cus_123",
			stripeSubscriptionID: "sub_123",
			setupMocks:           func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedError:        true,
		},
		{
			name: "should return error when customer ID is empty",
			subscription: &models.Subscription{
				UserID: 1,
				PlanID: "test_plan",
			},
			stripeCustomerID:     "",
			stripeSubscriptionID: "sub_123",
			setupMocks:           func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedError:        true,
		},
		{
			name: "should return error when subscription ID is empty",
			subscription: &models.Subscription{
				UserID: 1,
				PlanID: "test_plan",
			},
			stripeCustomerID:     "cus_123",
			stripeSubscriptionID: "",
			setupMocks:           func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSubscriptionRepository)
			mockRF := new(mocks.MockRFService)
			tt.setupMocks(mockRepo)

			service := NewSubscriptionService(mockRepo, mockRF)

			err := service.ActivateSubscription(tt.subscription, tt.stripeCustomerID, tt.stripeSubscriptionID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.stripeCustomerID, tt.subscription.StripeCustomerID)
				assert.Equal(t, tt.stripeSubscriptionID, tt.subscription.StripeSubscriptionID)
				assert.Equal(t, "active", tt.subscription.SubscriptionStatus)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_CancelSubscription(t *testing.T) {
	tests := []struct {
		name          string
		subscription  *models.Subscription
		setupMocks    func(*mocks.MockSubscriptionRepository)
		expectedError bool
	}{
		{
			name: "should cancel subscription successfully",
			subscription: &models.Subscription{
				UserID: 1,
				PlanID: "test_plan",
			},
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				mockRepo.On("Save", mock.AnythingOfType("*models.Subscription")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "should return error when subscription is nil",
			subscription:  nil,
			setupMocks:    func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSubscriptionRepository)
			mockRF := new(mocks.MockRFService)
			tt.setupMocks(mockRepo)

			service := NewSubscriptionService(mockRepo, mockRF)

			err := service.CancelSubscription(tt.subscription)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "canceled", tt.subscription.SubscriptionStatus)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_EndTrial(t *testing.T) {
	tests := []struct {
		name          string
		subscription  *models.Subscription
		setupMocks    func(*mocks.MockSubscriptionRepository)
		expectedError bool
	}{
		{
			name: "should end trial successfully",
			subscription: &models.Subscription{
				UserID:        1,
				PlanID:        "test_plan",
				IsTrialActive: true,
			},
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				mockRepo.On("Save", mock.AnythingOfType("*models.Subscription")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "should return error when subscription is nil",
			subscription:  nil,
			setupMocks:    func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSubscriptionRepository)
			mockRF := new(mocks.MockRFService)
			tt.setupMocks(mockRepo)

			service := NewSubscriptionService(mockRepo, mockRF)

			err := service.EndTrial(tt.subscription)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, tt.subscription.IsTrialActive)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionService_GetUserSubscriptionStatus(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		setupMocks     func(*mocks.MockSubscriptionRepository)
		expectedStatus string
		expectedDays   int
		expectedError  bool
	}{
		{
			name:   "should return trial status when user is in trial period",
			userID: 1,
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				subscription := &models.Subscription{
					UserID:        1,
					PlanID:        "test_plan",
					IsTrialActive: true,
					TrialEndDate:  time.Now().AddDate(0, 0, 5), // 5 days from now
				}
				mockRepo.On("FindByUserID", uint(1)).Return(subscription, nil)
			},
			expectedStatus: "trial",
			expectedDays:   5,
			expectedError:  false,
		},
		{
			name:   "should return active status when user has active subscription",
			userID: 1,
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				subscription := &models.Subscription{
					UserID:              1,
					PlanID:              "test_plan",
					IsTrialActive:       false,
					SubscriptionStatus:  "active",
					SubscriptionEndDate: &time.Time{}, // Will be set in test
				}
				endDate := time.Now().AddDate(0, 1, 0) // 1 month from now
				subscription.SubscriptionEndDate = &endDate
				mockRepo.On("FindByUserID", uint(1)).Return(subscription, nil)
			},
			expectedStatus: "active",
			expectedDays:   30,
			expectedError:  false,
		},
		{
			name:   "should return expiring status when subscription expires in 10 days",
			userID: 1,
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				subscription := &models.Subscription{
					UserID:              1,
					PlanID:              "test_plan",
					IsTrialActive:       false,
					SubscriptionStatus:  "active",
					SubscriptionEndDate: &time.Time{}, // Will be set in test
				}
				endDate := time.Now().AddDate(0, 0, 8) // 8 days from now
				subscription.SubscriptionEndDate = &endDate
				mockRepo.On("FindByUserID", uint(1)).Return(subscription, nil)
			},
			expectedStatus: "expiring",
			expectedDays:   8,
			expectedError:  false,
		},
		{
			name:   "should return inactive status when no subscription found",
			userID: 1,
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				mockRepo.On("FindByUserID", uint(1)).Return(nil, nil)
			},
			expectedStatus: "inactive",
			expectedDays:   0,
			expectedError:  false,
		},
		{
			name:           "should return error when user ID is zero",
			userID:         0,
			setupMocks:     func(mockRepo *mocks.MockSubscriptionRepository) {},
			expectedStatus: "inactive",
			expectedDays:   0,
			expectedError:  true,
		},
		{
			name:   "should return error when repository fails",
			userID: 1,
			setupMocks: func(mockRepo *mocks.MockSubscriptionRepository) {
				mockRepo.On("FindByUserID", uint(1)).Return(nil, errors.New("database error"))
			},
			expectedStatus: "inactive",
			expectedDays:   0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockSubscriptionRepository)
			mockRF := new(mocks.MockRFService)
			tt.setupMocks(mockRepo)

			service := NewSubscriptionService(mockRepo, mockRF)

			status, days, err := service.GetUserSubscriptionStatus(tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedStatus, status)

			// For days calculation, allow some flexibility due to time-based calculations
			if tt.expectedDays > 0 {
				// Allow ±1 day tolerance for time-based calculations
				assert.GreaterOrEqual(t, days, tt.expectedDays-1)
				assert.LessOrEqual(t, days, tt.expectedDays+1)
			} else {
				assert.Equal(t, tt.expectedDays, days)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
