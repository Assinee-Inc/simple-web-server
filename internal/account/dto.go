package account

import "time"

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
