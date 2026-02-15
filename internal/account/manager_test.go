package account_test

import (
	"errors"
	"testing"

	"github.com/anglesson/simple-web-server/internal/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAccountRepo struct {
	mock.Mock
}

func (m *mockAccountRepo) Create(account *account.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *mockAccountRepo) Update(account *account.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

type mockStripeService struct {
	mock.Mock
}

func (m *mockStripeService) CreateSellerAccount(account *account.Account) (string, error) {
	args := m.Called(account)
	return args.String(0), args.Error(1)
}

func TestAccountManager_CreateAccount(t *testing.T) {
	tests := []struct {
		name          string
		input         *account.Account
		expectedID    string
		setupMocks    func(m *mockAccountRepo, s *mockStripeService)
		expectedError error
	}{
		{
			name:       "success",
			input:      &account.Account{},
			expectedID: "any_connected_account_id",
			setupMocks: func(m *mockAccountRepo, s *mockStripeService) {
				m.On("Create", mock.Anything).Return(nil)
				s.On("CreateSellerAccount", mock.Anything).Return("any_connected_account_id", nil)
				m.On("Update", mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:       "database create error",
			input:      &account.Account{},
			expectedID: "",
			setupMocks: func(m *mockAccountRepo, s *mockStripeService) {
				m.On("Create", mock.Anything).Return(errors.New("repository fail"))
			},
			expectedError: account.InternalError,
		},
		{
			name:       "stripe error",
			input:      &account.Account{},
			expectedID: "",
			setupMocks: func(m *mockAccountRepo, s *mockStripeService) {
				m.On("Create", mock.Anything).Return(nil)
				s.On("CreateSellerAccount", mock.Anything).Return("", errors.New("stripe fail"))
			},
			expectedError: account.StripeIntegrationError,
		},
		{
			name:       "database update error",
			input:      &account.Account{},
			expectedID: "any_connected_account_id",
			setupMocks: func(m *mockAccountRepo, s *mockStripeService) {
				m.On("Create", mock.Anything).Return(nil)
				s.On("CreateSellerAccount", mock.Anything).Return("any_connected_account_id", nil)
				m.On("Update", mock.Anything).Return(errors.New("repository fail"))
			},
			expectedError: account.InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockAccountRepo)
			mockStripe := new(mockStripeService)
			tt.setupMocks(mockRepo, mockStripe)

			uc := account.NewManager(mockRepo, mockStripe)
			err := uc.CreateAccount(tt.input)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedID, tt.input.ExternalAccountID)
			mockRepo.AssertExpectations(t)
			mockStripe.AssertExpectations(t)
		})
	}
}
