package handler

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type StripeConnectHandler struct {
	stripeConnectService service.StripeConnectService
	creatorService       service.CreatorService
	sessionService       service.SessionService
	templateRenderer     template.TemplateRenderer
}

func NewStripeConnectHandler(
	stripeConnectService service.StripeConnectService,
	creatorService service.CreatorService,
	sessionService service.SessionService,
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
	// Get current user from session
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	// Find creator
	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		http.Error(w, "Creator não encontrado", http.StatusNotFound)
		return
	}

	// Check if onboarding is already complete
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
	// Get current user from session
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	// Find creator
	slog.Debug("Buscando dados do creator")
	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		http.Error(w, "Creator não encontrado", http.StatusNotFound)
		return
	}

	slog.Debug("Verifica se o infoprodutor já possui uma conta no Stripe")
	// Check if creator already has a Stripe Connect account AND is fully enabled
	if creator.StripeConnectAccountID != "" &&
		creator.OnboardingCompleted &&
		creator.ChargesEnabled &&
		creator.PayoutsEnabled {
		http.Redirect(w, r, "/dashboard?message=onboarding_completed", http.StatusSeeOther)
		return
	}

	var accountID string
	if creator.StripeConnectAccountID == "" {
		// Create new Stripe Connect account
		slog.Debug("Criando nova conta no Stripe")
		accountID, err = h.stripeConnectService.CreateConnectAccount(creator)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao criar conta: %v", err), http.StatusInternalServerError)
			return
		}

		// Update creator with account ID
		creator.StripeConnectAccountID = accountID
		// We'll update this through the creator service later
		h.creatorService.UpdateCreator(creator)
	} else {
		accountID = creator.StripeConnectAccountID
	}

	// Create onboarding URLs
	refreshURL := fmt.Sprintf("%s/stripe-connect/onboard", config.AppConfig.Host)
	returnURL := fmt.Sprintf("%s/stripe-connect/complete", config.AppConfig.Host)
	// Create onboarding link
	slog.Debug("Protocolo: " + r.Proto)
	slog.Debug("Criando link de onboarding: URL de retorno: " + returnURL)
	onboardingURL, err := h.stripeConnectService.CreateOnboardingLink(accountID, refreshURL, returnURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao criar link de onboarding: %v", err), http.StatusInternalServerError)
		return
	}

	// Store URLs for later use
	creator.OnboardingRefreshURL = refreshURL
	creator.OnboardingReturnURL = returnURL

	// Here we would typically update the creator in the database
	// For now, redirect to Stripe
	http.Redirect(w, r, onboardingURL, http.StatusSeeOther)
}

// CompleteOnboarding handles the return from Stripe after onboarding
func (h *StripeConnectHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	log.Printf("Completando Onboarding do Infoprodutor na stripe")
	// Get current user from session
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	// Find creator
	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		http.Error(w, "Creator não encontrado", http.StatusNotFound)
		return
	}

	if creator.StripeConnectAccountID == "" {
		http.Error(w, "Conta Stripe não encontrada", http.StatusBadRequest)
		return
	}

	// Get account details from Stripe
	account, err := h.stripeConnectService.GetAccountDetails(creator.StripeConnectAccountID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao verificar conta: %v", err), http.StatusInternalServerError)
		return
	}

	// Adicione logs
	slog.Info("Atualizando creator após onboarding",
		"creatorID", creator.ID,
		"stripeAccountID", creator.StripeConnectAccountID,
		"detailsSubmitted", account.DetailsSubmitted)

	// Update creator with account status
	err = h.stripeConnectService.UpdateCreatorFromAccount(creator, account)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao atualizar creator: %v", err), http.StatusInternalServerError)
		return
	}

	if account.DetailsSubmitted {
		// Onboarding completed successfully
		h.templateRenderer.View(w, r, "stripe-connect/success", map[string]any{
			"Creator":          creator,
			"OnboardingStatus": "completed",
		}, "admin")
	} else {
		// Onboarding still pending
		h.templateRenderer.View(w, r, "stripe-connect/pending", map[string]any{
			"Creator":          creator,
			"OnboardingStatus": "pending",
			"RefreshURL":       creator.OnboardingRefreshURL,
		}, "admin")
	}
}

// OnboardingStatus shows the current onboarding status
func (h *StripeConnectHandler) OnboardingStatus(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	// Find creator
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
		"NeedsOnboarding":     creator.StripeConnectAccountID != "" && (!creator.OnboardingCompleted || !creator.ChargesEnabled || !creator.PayoutsEnabled),
	}

	h.templateRenderer.View(w, r, "stripe-connect/status", data, "admin")
}
