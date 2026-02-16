package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

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

func (h *Handler) PostAccountSSR(w http.ResponseWriter, r *http.Request) {
	account := h.parseFormDataToDomain(r)

	err := h.manager.CreateAccount(account)
	if err != nil {
		if errors.Is(err, ErrInvalidAccount) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, StripeIntegrationError) {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		} else {
			http.Error(w, InternalError.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) parseFormDataToDomain(r *http.Request) *Account {
	// Parse form data into Account struct
	accountData := CreateAccountRequest{
		Name:      r.FormValue("name"),
		Email:     r.FormValue("email"),
		CPF:       r.FormValue("cpf"),
		Phone:     r.FormValue("phone"),
		BirthDate: 0,
	}

	if birthDateStr := r.FormValue("birth_date"); birthDateStr != "" {
		if birthDate, err := time.Parse("2006-01-02", birthDateStr); err == nil {
			accountData.BirthDate = birthDate.Unix()
		}
	}

	if origin := r.FormValue("origin"); origin != "" {
		accountData.Origin = origin
	}

	if onboardingRefreshURL := r.FormValue("onboarding_refresh_url"); onboardingRefreshURL != "" {
		accountData.OnboardingRefreshURL = onboardingRefreshURL
	}

	if onboardingReturnURL := r.FormValue("onboarding_return_url"); onboardingReturnURL != "" {
		accountData.OnboardingReturnURL = onboardingReturnURL
	}

	return accountData.ToDomain()
}
