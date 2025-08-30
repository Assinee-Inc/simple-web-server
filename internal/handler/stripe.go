package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
)

type StripeHandler struct {
	userRepository      repository.UserRepository
	subscriptionService service.SubscriptionService
	purchaseRepository  *repository.PurchaseRepository
	emailService        *service.EmailService
}

func NewStripeHandler(
	userRepository repository.UserRepository,
	subscriptionService service.SubscriptionService,
	purchaseRepository *repository.PurchaseRepository,
	emailService *service.EmailService,
) *StripeHandler {
	return &StripeHandler{
		userRepository:      userRepository,
		subscriptionService: subscriptionService,
		purchaseRepository:  purchaseRepository,
		emailService:        emailService,
	}
}

func (h *StripeHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Initialize Stripe with API key
	stripe.Key = config.AppConfig.StripeSecretKey

	// Get session token from cookie
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		log.Printf("Session token not found in cookie: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Não autorizado",
		})
		return
	}

	log.Printf("Session token found for user")

	// Find user by session token
	user := h.userRepository.FindBySessionToken(sessionCookie.Value)
	if user == nil {
		log.Printf("User not found for session token")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Não autorizado",
		})
		return
	}

	log.Printf("User found: %s", user.Email)

	// Validate CSRF token
	csrfToken := r.Header.Get("X-CSRF-Token")
	log.Printf("CSRF token received from header")
	log.Printf("User CSRF token validated")

	if csrfToken == "" {
		log.Printf("Token CSRF não encontrado no header")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Token CSRF não encontrado",
		})
		return
	}

	if csrfToken != user.CSRFToken {
		log.Printf("CSRF token mismatch for user: %s", user.Email)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Token CSRF inválido",
		})
		return
	}

	// Get user's subscription
	subscription, err := h.subscriptionService.FindByUserID(user.ID)
	if err != nil {
		log.Printf("Erro ao buscar subscription do usuário: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao processar pagamento",
		})
		return
	}

	if subscription == nil || subscription.StripeCustomerID == "" {
		log.Printf("Usuário %s não possui subscription ou ID do Stripe", user.Email)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao processar pagamento: Cliente não encontrado",
		})
		return
	}

	// Log Stripe configuration
	log.Printf("Stripe Secret Key: %s", config.AppConfig.StripeSecretKey)
	log.Printf("Stripe Price ID: %s", config.AppConfig.StripePriceID)

	if config.AppConfig.StripeSecretKey == "" {
		log.Printf("Stripe Secret Key não configurada")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro de configuração do Stripe",
		})
		return
	}

	if config.AppConfig.StripePriceID == "" {
		log.Printf("Stripe Price ID não configurado")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro de configuração do Stripe",
		})
		return
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(subscription.StripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(config.AppConfig.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("http://" + r.Host + "/settings?success=true"),
		CancelURL:  stripe.String("http://" + r.Host + "/settings?canceled=true"),
	}

	log.Printf("Criando sessão do Stripe com os parâmetros: %+v", params)

	s, err := session.New(params)
	if err != nil {
		log.Printf("Erro ao criar sessão do Stripe: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao processar pagamento: " + err.Error(),
		})
		return
	}

	response := struct {
		URL string `json:"url"`
	}{
		URL: s.URL,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Erro ao codificar resposta: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erro ao processar resposta",
		})
		return
	}
}

func (h *StripeHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Configure webhook options to ignore API version mismatch
	opts := webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	}

	event, err := webhook.ConstructEventWithOptions(payload, r.Header.Get("Stripe-Signature"), config.AppConfig.StripeWebhookSecret, opts)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing checkout session: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verificar se é um pagamento de ebook ou assinatura
		if session.Mode == stripe.CheckoutSessionModePayment {
			// É um pagamento de ebook
			err = h.handleEbookPayment(session)
			if err != nil {
				log.Printf("Error handling ebook payment: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if session.Mode == stripe.CheckoutSessionModeSubscription {
			// É uma assinatura
			err = h.handleSubscriptionPayment(session)
			if err != nil {
				log.Printf("Error handling subscription payment: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

	case "customer.subscription.updated":
		var stripeSubscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &stripeSubscription)
		if err != nil {
			log.Printf("Error parsing subscription: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Find subscription by Stripe customer ID
		subscription, err := h.subscriptionService.FindByStripeCustomerID(stripeSubscription.Customer.ID)
		if err != nil {
			log.Printf("Error finding subscription: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if subscription == nil {
			log.Printf("Subscription not found for Stripe customer ID: %s", stripeSubscription.Customer.ID)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Update subscription status
		var endDate *time.Time
		if stripeSubscription.CurrentPeriodEnd > 0 {
			endDate = &time.Time{}
			*endDate = time.Unix(stripeSubscription.CurrentPeriodEnd, 0)
		}
		err = h.subscriptionService.UpdateSubscriptionStatus(subscription, string(stripeSubscription.Status), endDate)
		if err != nil {
			log.Printf("Error updating subscription: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "customer.subscription.deleted":
		var stripeSubscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &stripeSubscription)
		if err != nil {
			log.Printf("Error parsing subscription: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Find subscription by Stripe customer ID
		subscription, err := h.subscriptionService.FindByStripeCustomerID(stripeSubscription.Customer.ID)
		if err != nil {
			log.Printf("Error finding subscription: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if subscription == nil {
			log.Printf("Subscription not found for Stripe customer ID: %s", stripeSubscription.Customer.ID)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Update subscription status
		var endDate *time.Time
		if stripeSubscription.CurrentPeriodEnd > 0 {
			endDate = &time.Time{}
			*endDate = time.Unix(stripeSubscription.CurrentPeriodEnd, 0)
		}
		err = h.subscriptionService.UpdateSubscriptionStatus(subscription, "canceled", endDate)
		if err != nil {
			log.Printf("Error updating subscription: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// handleEbookPayment processa pagamento de ebook
func (h *StripeHandler) handleEbookPayment(session stripe.CheckoutSession) error {
	// Extrair dados da sessão
	ebookIDStr := session.Metadata["ebook_id"]
	clientIDStr := session.Metadata["client_id"]

	if ebookIDStr == "" || clientIDStr == "" {
		return fmt.Errorf("dados da compra inválidos")
	}

	// Converter IDs
	ebookID, err := strconv.ParseUint(ebookIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("ebook ID inválido: %v", err)
	}

	clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("client ID inválido: %v", err)
	}

	// Criar registro de compra
	purchase := models.NewPurchase(uint(ebookID), uint(clientID))
	purchase.ExpiresAt = time.Now().AddDate(0, 0, 30) // 30 dias de acesso

	err = h.purchaseRepository.CreateManyPurchases([]*models.Purchase{purchase})
	if err != nil {
		return fmt.Errorf("erro ao criar compra: %v", err)
	}

	// Enviar email com link de download
	if purchase.ID > 0 {
		log.Printf("Purchase criado com sucesso: ID=%d, EbookID=%d, ClientID=%d", purchase.ID, purchase.EbookID, purchase.ClientID)

		// Buscar purchase com relacionamentos do banco
		log.Printf("🔍 Buscando purchase com ID=%d para carregar relacionamentos", purchase.ID)
		log.Printf("🔍 Usando purchaseRepository: %+v", h.purchaseRepository)

		// Verificar se purchaseRepository não é nil
		if h.purchaseRepository == nil {
			log.Printf("❌ ERRO: purchaseRepository é nil!")
			return fmt.Errorf("purchaseRepository não inicializado")
		}

		purchaseWithRelations, err := h.purchaseRepository.FindByID(purchase.ID)
		if err != nil {
			log.Printf("❌ Erro ao buscar purchase com relacionamentos: %v", err)
			return fmt.Errorf("erro ao buscar dados da compra: %v", err)
		}

		if purchaseWithRelations == nil {
			log.Printf("❌ Purchase não encontrado após criação")
			return fmt.Errorf("purchase não encontrado após criação")
		}

		log.Printf("✅ Purchase carregado: ID=%d, ClientID=%d",
			purchaseWithRelations.ID, purchaseWithRelations.ClientID)

		// Verificar se o cliente foi carregado
		if purchaseWithRelations.Client.ID == 0 {
			log.Printf("❌ Cliente não foi carregado! Client.ID=0")
		} else {
			log.Printf("✅ Cliente carregado: ID=%d, Name='%s', Email='%s'",
				purchaseWithRelations.Client.ID,
				purchaseWithRelations.Client.Name,
				purchaseWithRelations.Client.Email)
		}

		// Verificar se o cliente tem email
		if purchaseWithRelations.Client.Email == "" {
			log.Printf("Cliente sem email: ClientID=%d", purchaseWithRelations.ClientID)
			return fmt.Errorf("cliente sem email válido")
		}

		log.Printf("Enviando email para: %s", purchaseWithRelations.Client.Email)

		go h.emailService.SendLinkToDownload([]*models.Purchase{purchaseWithRelations})
	}

	return nil
}

// handleSubscriptionPayment processa pagamento de assinatura
func (h *StripeHandler) handleSubscriptionPayment(session stripe.CheckoutSession) error {
	// Find subscription by Stripe customer ID
	subscription, err := h.subscriptionService.FindByStripeCustomerID(session.Customer.ID)
	if err != nil {
		return fmt.Errorf("error finding subscription: %v", err)
	}
	if subscription == nil {
		return fmt.Errorf("subscription not found for Stripe customer ID: %s", session.Customer.ID)
	}

	// Update subscription status
	err = h.subscriptionService.ActivateSubscription(subscription, session.Customer.ID, session.Subscription.ID)
	if err != nil {
		return fmt.Errorf("error updating subscription: %v", err)
	}

	return nil
}
