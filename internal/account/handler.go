package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type CreateAccountRequest struct {
	Name                 string `json:"name" validate:"required"`
	CPF                  string `json:"cpf" validate:"required"`
	Email                string `json:"email" validate:"required,email"`
	Phone                string `json:"phone" validate:"required"`
	BirthDate            int64  `json:"birth_date" validate:"required"`
	Origin               string `json:"origin"`
	OnboardingRefreshURL string `json:"onboarding_refresh_url,omitempty" validate:"omitempty,url"`
	OnboardingReturnURL  string `json:"onboarding_return_url,omitempty" validate:"omitempty,url"`
}

func (r *CreateAccountRequest) ToDomain() *Account {
	return &Account{
		Name:                 r.Name,
		CPF:                  r.CPF,
		Email:                r.Email,
		Phone:                r.Phone,
		BirthDate:            time.Unix(r.BirthDate, 0),
		Origin:               r.Origin,
		OnboardingRefreshURL: r.OnboardingRefreshURL,
		OnboardingReturnURL:  r.OnboardingReturnURL,
	}
}

type Manager interface {
	CreateAccount(account *Account) error
}

type Handler struct {
	manager Manager
}

func NewHandler(manager Manager) *Handler {
	return &Handler{manager: manager}
}

func (h *Handler) PostAccount(w http.ResponseWriter, r *http.Request) {
	// Parse request body into Account struct
	var accountData CreateAccountRequest
	err := json.NewDecoder(r.Body).Decode(&accountData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields (basic validation example)
	if accountData.Name == "" || accountData.Email == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create account using the manager
	accountDomain := accountData.ToDomain()
	err = h.manager.CreateAccount(accountDomain)
	if err != nil {
		if errors.Is(err, ErrInvalidAccount) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if errors.Is(err, StripeIntegrationError) {
			http.Error(w, err.Error(), http.StatusBadGateway)
		} else {
			http.Error(w, InternalError.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(accountDomain)
}
