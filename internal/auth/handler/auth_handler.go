package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/config"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

// CreatorFinder define a interface mínima consumida pelo AuthHandler para verificar se o usuário possui perfil Creator.
type CreatorFinder interface {
	FindCreatorByUserID(userID uint) (*accountmodel.Creator, error)
}

type AuthHandler struct {
	userService    authsvc.UserService
	tokenService   authsvc.TokenService
	creatorFinder  CreatorFinder
	sessionService authsvc.SessionService
}

// NewAuthHandler cria uma instância do AuthHandler suportando injeção de TokenService e CreatorFinder.
func NewAuthHandler(userService authsvc.UserService, deps ...any) *AuthHandler {
	h := &AuthHandler{
		userService: userService,
	}

	for _, dep := range deps {
		switch d := dep.(type) {
		case authsvc.TokenService:
			h.tokenService = d
		case CreatorFinder:
			h.creatorFinder = d
		case authsvc.SessionService:
			h.sessionService = d
		}
	}

	if h.tokenService == nil {
		expHours := config.AppConfig.JWTExpirationHours
		if expHours <= 0 {
			expHours = 24
		}
		h.tokenService = authsvc.NewTokenService(config.AppConfig.JWTSecret, time.Duration(expHours)*time.Hour)
	}

	return h
}

// GoogleLogin autentica ou cria um usuário via ID Token do Google (`POST /api/auth/google`).
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var input authmodel.GoogleAuthInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_JSON", "Corpo da requisição JSON inválido", nil)
		return
	}

	input.IDToken = strings.TrimSpace(input.IDToken)
	if input.IDToken == "" {
		web.RespondValidationError(w, map[string]string{
			"id_token": "O campo id_token é obrigatório",
		})
		return
	}

	user, isNewUser, err := h.userService.AuthenticateWithGoogle(r.Context(), input.IDToken)
	if err != nil || user == nil {
		web.RespondError(w, http.StatusUnauthorized, "INVALID_GOOGLE_TOKEN", "Credencial Google inválida ou e-mail não verificado", nil)
		return
	}

	accessToken, expiresIn, err := h.tokenService.GenerateToken(user)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Erro ao emitir token de acesso", nil)
		return
	}

	onboardingRequired := true
	if h.creatorFinder != nil && user.ID > 0 {
		if creator, err := h.creatorFinder.FindCreatorByUserID(user.ID); err == nil && creator != nil {
			onboardingRequired = false
		}
	}

	status := http.StatusOK
	if isNewUser {
		status = http.StatusCreated
	}

	web.RespondJSON(w, status, authmodel.AuthResponse{
		Success:            true,
		AccessToken:        accessToken,
		TokenType:          "Bearer",
		ExpiresIn:          expiresIn,
		OnboardingRequired: onboardingRequired,
		User: authmodel.UserSummaryDTO{
			PublicID:      user.PublicID,
			Username:      user.Username,
			Email:         user.Email,
			GoogleID:      user.GoogleID,
			EmailVerified: user.IsEmailVerified(),
			TermsAccepted: user.HasAcceptedTerms(),
		},
	})
}

// Me retorna os dados do usuário autenticado, o estado do perfil Creator e o status da assinatura (`GET /api/auth/me`).
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := authmw.GetAuthenticatedUser(r.Context())
	if user == nil {
		userID, ok := authmw.GetAuthenticatedUserID(r.Context())
		if !ok || userID == 0 {
			web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
			return
		}
		user = h.userService.FindByID(userID)
		if user == nil {
			web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não encontrado", nil)
			return
		}
	} else if h.userService != nil && user.ID > 0 {
		if fullUser := h.userService.FindByID(user.ID); fullUser != nil {
			user = fullUser
		}
	}

	var creatorPayload any
	onboardingRequired := true
	if h.creatorFinder != nil && user.ID > 0 {
		if creator, err := h.creatorFinder.FindCreatorByUserID(user.ID); err == nil && creator != nil {
			onboardingRequired = false
			creatorPayload = map[string]any{
				"public_id":               creator.PublicID,
				"name":                    creator.Name,
				"social_name":             creator.SocialName,
				"email":                   creator.Email,
				"phone":                   creator.Phone,
				"onboarding_completed":    creator.OnboardingCompleted,
				"charges_enabled":         creator.ChargesEnabled,
				"payouts_enabled":         creator.PayoutsEnabled,
				"needs_stripe_onboarding": creator.NeedsOnboarding(),
			}
		}
	}

	daysLeft := user.DaysLeftInSubscription()
	if daysLeft == 0 && user.IsInTrialPeriod() {
		daysLeft = user.DaysLeftInTrial()
	}

	resp := map[string]any{
		"success":             true,
		"onboarding_required": onboardingRequired,
		"user": authmodel.UserSummaryDTO{
			PublicID:           user.PublicID,
			Username:           user.Username,
			Email:              user.Email,
			GoogleID:           user.GoogleID,
			EmailVerified:      user.IsEmailVerified(),
			TermsAccepted:      user.HasAcceptedTerms(),
			SubscriptionStatus: user.GetSubscriptionStatus(),
			DaysLeft:           daysLeft,
		},
	}
	if creatorPayload != nil {
		resp["creator"] = creatorPayload
	}

	web.RespondJSON(w, http.StatusOK, resp)
}

// Logout encerra a sessão atual (`POST /api/auth/logout`).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if h.sessionService != nil {
		_ = h.sessionService.Destroy(r, w)
	}
	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Sessão encerrada com sucesso",
	})
}

// LogoutSubmit é um alias para Logout garantindo compatibilidade com referências existentes.
func (h *AuthHandler) LogoutSubmit(w http.ResponseWriter, r *http.Request) {
	h.Logout(w, r)
}
