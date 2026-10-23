package handler

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/config"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

type StripeConnectHandler struct {
	stripeConnectService accountsvc.StripeConnectService
	creatorService       accountsvc.CreatorService
	sessionService       authsvc.SessionService
}

func NewStripeConnectHandler(
	stripeConnectService accountsvc.StripeConnectService,
	creatorService accountsvc.CreatorService,
	deps ...any,
) *StripeConnectHandler {
	h := &StripeConnectHandler{
		stripeConnectService: stripeConnectService,
		creatorService:       creatorService,
	}
	for _, dep := range deps {
		if s, ok := dep.(authsvc.SessionService); ok {
			h.sessionService = s
		}
	}
	return h
}

func (h *StripeConnectHandler) resolveCreator(r *http.Request) (*accountmodel.Creator, int, string) {
	if c := accountmw.CreatorFromContext(r.Context()); c != nil {
		return c, http.StatusOK, ""
	}
	if uid, ok := authmw.GetAuthenticatedUserID(r.Context()); ok && uid > 0 {
		creator, err := h.creatorService.FindCreatorByUserID(uid)
		if err != nil || creator == nil {
			return nil, http.StatusNotFound, "Creator não encontrado"
		}
		return creator, http.StatusOK, ""
	}
	if email := authmw.GetAuthenticatedUserEmail(r.Context()); email != "" {
		creator, err := h.creatorService.FindCreatorByEmail(email)
		if err != nil || creator == nil {
			return nil, http.StatusNotFound, "Creator não encontrado"
		}
		return creator, http.StatusOK, ""
	}
	if h.sessionService != nil {
		if email, err := h.sessionService.GetUserEmailFromSession(r); err == nil && email != "" {
			creator, err := h.creatorService.FindCreatorByEmail(email)
			if err != nil || creator == nil {
				return nil, http.StatusNotFound, "Creator não encontrado"
			}
			return creator, http.StatusOK, ""
		}
	}
	return nil, http.StatusUnauthorized, "Não autorizado"
}

// OnboardingWelcome retorna o status de onboarding Stripe em JSON (`GET /stripe-connect/welcome`).
func (h *StripeConnectHandler) OnboardingWelcome(w http.ResponseWriter, r *http.Request) {
	h.OnboardingStatus(w, r)
}

// StartOnboarding cria a conta Stripe Connect (se necessário) e retorna a URL de onboarding em JSON (`GET /stripe-connect/onboard`).
func (h *StripeConnectHandler) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	creator, status, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, status, "STRIPE_ONBOARDING_ERROR", errMsg, nil)
		return
	}

	if creator.StripeConnectAccountID != "" &&
		creator.OnboardingCompleted &&
		creator.ChargesEnabled &&
		creator.PayoutsEnabled {
		web.RespondJSON(w, http.StatusOK, map[string]any{
			"success":              true,
			"onboarding_completed": true,
			"creator":              creator,
		})
		return
	}

	var accountID string
	var err error
	if creator.StripeConnectAccountID == "" {
		accountID, err = h.stripeConnectService.CreateConnectAccount(creator)
		if err != nil {
			web.RespondError(w, http.StatusInternalServerError, "STRIPE_ACCOUNT_CREATE_FAILED", fmt.Sprintf("Erro ao criar conta: %v", err), nil)
			return
		}
		creator.StripeConnectAccountID = accountID
		_ = h.creatorService.UpdateCreator(creator)
	} else {
		accountID = creator.StripeConnectAccountID
	}

	frontendURL := strings.TrimRight(config.AppConfig.FrontendURL, "/")
	fallbackBase := resolveFallbackHost(config.AppConfig.Host, config.AppConfig.Port)

	returnURL := r.URL.Query().Get("return_url")
	if returnURL == "" {
		if frontendURL != "" {
			returnURL = fmt.Sprintf("%s/stripe-connect?status=complete", frontendURL)
		} else {
			returnURL = fmt.Sprintf("%s/stripe-connect/complete", fallbackBase)
		}
	}

	refreshURL := r.URL.Query().Get("refresh_url")
	if refreshURL == "" {
		if frontendURL != "" {
			refreshURL = fmt.Sprintf("%s/stripe-connect?status=refresh", frontendURL)
		} else {
			refreshURL = fmt.Sprintf("%s/stripe-connect/onboard", fallbackBase)
		}
	}

	onboardingURL, err := h.stripeConnectService.CreateOnboardingLink(accountID, refreshURL, returnURL)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "STRIPE_ONBOARDING_LINK_FAILED", fmt.Sprintf("Erro ao criar link de onboarding: %v", err), nil)
		return
	}

	creator.OnboardingRefreshURL = refreshURL
	creator.OnboardingReturnURL = returnURL

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":        true,
		"onboarding_url": onboardingURL,
	})
}

// CompleteOnboarding sincroniza os dados da conta Stripe após retorno do onboarding e retorna o status em JSON (`GET /stripe-connect/complete`).
func (h *StripeConnectHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	creator, status, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, status, "STRIPE_ONBOARDING_ERROR", errMsg, nil)
		return
	}

	if creator.StripeConnectAccountID == "" {
		web.RespondError(w, http.StatusBadRequest, "STRIPE_ACCOUNT_MISSING", "Conta Stripe não encontrada", nil)
		return
	}

	account, err := h.stripeConnectService.GetAccountDetails(creator.StripeConnectAccountID)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "STRIPE_ACCOUNT_FETCH_FAILED", fmt.Sprintf("Erro ao verificar conta: %v", err), nil)
		return
	}

	slog.Info("Atualizando creator após onboarding",
		"creatorID", creator.ID,
		"stripeAccountID", creator.StripeConnectAccountID,
		"detailsSubmitted", account.DetailsSubmitted)

	if err := h.stripeConnectService.UpdateCreatorFromAccount(creator, account); err != nil {
		web.RespondError(w, http.StatusInternalServerError, "CREATOR_UPDATE_FAILED", fmt.Sprintf("Erro ao atualizar creator: %v", err), nil)
		return
	}

	onboardingStatus := "pending"
	if account.DetailsSubmitted {
		onboardingStatus = "completed"
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":           true,
		"onboarding_status": onboardingStatus,
		"refresh_url":       creator.OnboardingRefreshURL,
		"creator":           creator,
	})
}

// OnboardingStatus retorna o status detalhado da conta Stripe Connect em JSON (`GET /stripe-connect/status`).
func (h *StripeConnectHandler) OnboardingStatus(w http.ResponseWriter, r *http.Request) {
	creator, status, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, status, "STRIPE_ONBOARDING_ERROR", errMsg, nil)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":              true,
		"creator":              creator,
		"has_stripe_account":   creator.StripeConnectAccountID != "",
		"onboarding_completed": creator.OnboardingCompleted,
		"payouts_enabled":      creator.PayoutsEnabled,
		"charges_enabled":      creator.ChargesEnabled,
		"can_start_onboarding": creator.StripeConnectAccountID == "",
		"needs_onboarding":     creator.NeedsOnboarding(),
		"remediation_link":     creator.OnboardingRefreshURL,
	})
}

func resolveFallbackHost(host, port string) string {
	cleanHost := strings.TrimRight(host, "/")
	if cleanHost == "" {
		return ""
	}
	hasPort := false
	if u, err := url.Parse(cleanHost); err == nil && u.Host != "" {
		hasPort = (u.Port() != "")
	} else if _, _, err := net.SplitHostPort(cleanHost); err == nil {
		hasPort = true
	}

	if !hasPort && port != "" {
		return fmt.Sprintf("%s:%s", cleanHost, port)
	}
	return cleanHost
}
