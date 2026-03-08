package handler

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	"github.com/anglesson/simple-web-server/internal/config"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type StripeConnectHandler struct {
	stripeConnectService accountsvc.StripeConnectService
	creatorService       accountsvc.CreatorService
	sessionService       authsvc.SessionService
	templateRenderer     template.TemplateRenderer
}

func NewStripeConnectHandler(
	stripeConnectService accountsvc.StripeConnectService,
	creatorService accountsvc.CreatorService,
	sessionService authsvc.SessionService,
	templateRenderer template.TemplateRenderer,
) *StripeConnectHandler {
	return &StripeConnectHandler{
		stripeConnectService: stripeConnectService,
		creatorService:       creatorService,
		sessionService:       sessionService,
		templateRenderer:     templateRenderer,
	}
}

// OnboardingWelcome shows welcome page for new users
func (h *StripeConnectHandler) OnboardingWelcome(w http.ResponseWriter, r *http.Request) {
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		http.Error(w, "Creator não encontrado", http.StatusNotFound)
		return
	}

	if creator.OnboardingCompleted {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"Creator": creator,
	}

	h.templateRenderer.View(w, r, "stripe-connect/onboard-welcome", data, "guest")
}

// StartOnboarding initiates the Stripe Connect onboarding process
func (h *StripeConnectHandler) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Iniciando Onboarding do Infoprodutor na stripe")
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	slog.Debug("Buscando dados do creator")
	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		http.Error(w, "Creator não encontrado", http.StatusNotFound)
		return
	}

	slog.Debug("Verifica se o infoprodutor já possui uma conta no Stripe")
	if creator.StripeConnectAccountID != "" &&
		creator.OnboardingCompleted &&
		creator.ChargesEnabled &&
		creator.PayoutsEnabled {
		http.Redirect(w, r, "/dashboard?message=onboarding_completed", http.StatusSeeOther)
		return
	}

	var accountID string
	if creator.StripeConnectAccountID == "" {
		slog.Debug("Criando nova conta no Stripe")
		accountID, err = h.stripeConnectService.CreateConnectAccount(creator)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao criar conta: %v", err), http.StatusInternalServerError)
			return
		}

		creator.StripeConnectAccountID = accountID
		h.creatorService.UpdateCreator(creator)
	} else {
		accountID = creator.StripeConnectAccountID
	}

	refreshURL := fmt.Sprintf("%s/stripe-connect/onboard", config.AppConfig.Host)
	returnURL := fmt.Sprintf("%s/stripe-connect/complete", config.AppConfig.Host)

	slog.Debug("Protocolo: " + r.Proto)
	slog.Debug("Criando link de onboarding: URL de retorno: " + returnURL)
	onboardingURL, err := h.stripeConnectService.CreateOnboardingLink(accountID, refreshURL, returnURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao criar link de onboarding: %v", err), http.StatusInternalServerError)
		return
	}

	creator.OnboardingRefreshURL = refreshURL
	creator.OnboardingReturnURL = returnURL

	http.Redirect(w, r, onboardingURL, http.StatusSeeOther)
}

// CompleteOnboarding handles the return from Stripe after onboarding
func (h *StripeConnectHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	log.Printf("Completando Onboarding do Infoprodutor na stripe")
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		http.Error(w, "Creator não encontrado", http.StatusNotFound)
		return
	}

	if creator.StripeConnectAccountID == "" {
		http.Error(w, "Conta Stripe não encontrada", http.StatusBadRequest)
		return
	}

	account, err := h.stripeConnectService.GetAccountDetails(creator.StripeConnectAccountID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao verificar conta: %v", err), http.StatusInternalServerError)
		return
	}

	slog.Info("Atualizando creator após onboarding",
		"creatorID", creator.ID,
		"stripeAccountID", creator.StripeConnectAccountID,
		"detailsSubmitted", account.DetailsSubmitted)

	err = h.stripeConnectService.UpdateCreatorFromAccount(creator, account)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao atualizar creator: %v", err), http.StatusInternalServerError)
		return
	}

	if account.DetailsSubmitted {
		h.templateRenderer.View(w, r, "stripe-connect/success", map[string]any{
			"Creator":          creator,
			"OnboardingStatus": "completed",
		}, "admin-daisy")
	} else {
		h.templateRenderer.View(w, r, "stripe-connect/pending", map[string]any{
			"Creator":          creator,
			"OnboardingStatus": "pending",
			"RefreshURL":       creator.OnboardingRefreshURL,
		}, "admin-daisy")
	}
}

// OnboardingStatus shows the current onboarding status
func (h *StripeConnectHandler) OnboardingStatus(w http.ResponseWriter, r *http.Request) {
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	slog.Debug("Buscando dados do creator")
	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		http.Error(w, "Creator não encontrado", http.StatusNotFound)
		return
	}

	data := map[string]any{
		"Creator":             creator,
		"HasStripeAccount":    creator.StripeConnectAccountID != "",
		"OnboardingCompleted": creator.OnboardingCompleted,
		"PayoutsEnabled":      creator.PayoutsEnabled,
		"ChargesEnabled":      creator.ChargesEnabled,
		"CanStartOnboarding":  creator.StripeConnectAccountID == "",
		"NeedsOnboarding":     creator.NeedsOnboarding(),
		"RemediationLink":     creator.OnboardingRefreshURL,
	}

	h.templateRenderer.View(w, r, "stripe-connect/status", data, "admin-daisy")
}
