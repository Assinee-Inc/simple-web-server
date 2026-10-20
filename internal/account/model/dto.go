package model

// InputOnboardingCreator representa os dados da Etapa 2 do onboarding de criador (sem senha).
type InputOnboardingCreator struct {
	Name          string `json:"name"`
	SocialName    string `json:"social_name"`
	CPF           string `json:"cpf"`
	BirthDate     string `json:"birthdate"`
	PhoneNumber   string `json:"phone"`
	TermsAccepted bool   `json:"terms_accepted"`
}

// CreatorDTO representa os dados públicos/operacionais do Creator em respostas JSON.
type CreatorDTO struct {
	PublicID               string `json:"public_id"`
	Name                   string `json:"name"`
	SocialName             string `json:"social_name"`
	Email                  string `json:"email"`
	CPF                    string `json:"cpf,omitempty"`
	Phone                  string `json:"phone"`
	StripeConnectAccountID string `json:"stripe_connect_account_id,omitempty"`
	OnboardingCompleted    bool   `json:"onboarding_completed"`
	ChargesEnabled         bool   `json:"charges_enabled,omitempty"`
	PayoutsEnabled         bool   `json:"payouts_enabled,omitempty"`
	NeedsStripeOnboarding  bool   `json:"needs_stripe_onboarding,omitempty"`
}

// InputCreateCreator mantido por compatibilidade legada.
type InputCreateCreator struct {
	Name                 string
	SocialName           string
	CPF                  string
	BirthDate            string
	PhoneNumber          string
	Email                string
	Password             string
	PasswordConfirmation string
	TermsAccepted        string
}
