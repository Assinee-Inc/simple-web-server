package model

// GoogleAuthInput representa o payload de autenticação via Google OIDC.
type GoogleAuthInput struct {
	IDToken string `json:"id_token"`
}

// GoogleIdentity representa as claims validadas do ID Token do Google.
type GoogleIdentity struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

// UserSummaryDTO resume os dados do usuário autenticado retornados nos contratos JSON.
type UserSummaryDTO struct {
	PublicID           string `json:"public_id"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	GoogleID           string `json:"google_id,omitempty"`
	EmailVerified      bool   `json:"email_verified"`
	TermsAccepted      bool   `json:"terms_accepted"`
	SubscriptionStatus string `json:"subscription_status,omitempty"`
	DaysLeft           int    `json:"days_left,omitempty"`
}

// AuthResponse representa a resposta de sucesso de POST /api/auth/google.
type AuthResponse struct {
	Success            bool           `json:"success"`
	AccessToken        string         `json:"access_token"`
	TokenType          string         `json:"token_type"`
	ExpiresIn          int64          `json:"expires_in"`
	OnboardingRequired bool           `json:"onboarding_required"`
	User               UserSummaryDTO `json:"user"`
}

// InputCreateUser mantido temporariamente como alias para evitar quebra de compilação fora do pacote auth antes da Fase 4.
type InputCreateUser struct {
	Username             string
	Email                string
	Password             string
	PasswordConfirmation string
}

// InputLogin mantido temporariamente para compatibilidade.
type InputLogin struct {
	Email    string
	Password string
}
