package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	"github.com/anglesson/simple-web-server/internal/config"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	subscriptionservice "github.com/anglesson/simple-web-server/internal/subscription/service"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
)

type StripeHandler struct {
	userRepository      authrepo.UserRepository
	subscriptionService subscriptionservice.SubscriptionService
	purchaseRepository  *salesrepo.PurchaseRepository
	purchaseService     salesvc.PurchaseService
	emailService        salesvc.IEmailService
	transactionService  salesvc.TransactionService
	creatorService      accountsvc.CreatorService
}

func NewStripeHandler(
	userRepository authrepo.UserRepository,
	subscriptionService subscriptionservice.SubscriptionService,
	purchaseRepository *salesrepo.PurchaseRepository,
	purchaseService salesvc.PurchaseService,
	emailService salesvc.IEmailService,
	transactionService salesvc.TransactionService,
	creatorService accountsvc.CreatorService,
) *StripeHandler {
	return &StripeHandler{
		userRepository:      userRepository,
		subscriptionService: subscriptionService,
		purchaseRepository:  purchaseRepository,
		purchaseService:     purchaseService,
		emailService:        emailService,
		transactionService:  transactionService,
		creatorService:      creatorService,
	}
}

func (h *StripeHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stripe.Key = config.AppConfig.StripeSecretKey

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

	var event stripe.Event

	if config.AppConfig.StripeWebhookSecret == "" {
		log.Printf("Warning: STRIPE_WEBHOOK_SECRET not set, skipping signature verification")
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("Error parsing webhook payload: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		opts := webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		}
		var err error
		event, err = webhook.ConstructEventWithOptions(payload, r.Header.Get("Stripe-Signature"), config.AppConfig.StripeWebhookSecret, opts)
		if err != nil {
			log.Printf("Error verifying webhook signature: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	switch event.Type {
	case "checkout.session.completed":
		var stripeSession stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &stripeSession)
		if err != nil {
			log.Printf("Error parsing checkout session: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if stripeSession.Mode == stripe.CheckoutSessionModePayment {
			err = h.handleEbookPayment(stripeSession)
			if err != nil {
				log.Printf("Error handling ebook payment: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if stripeSession.Mode == stripe.CheckoutSessionModeSubscription {
			err = h.handleSubscriptionPayment(stripeSession)
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
func (h *StripeHandler) handleEbookPayment(stripeSession stripe.CheckoutSession) error {
	ebookIDStr := stripeSession.Metadata["ebook_id"]
	clientIDStr := stripeSession.Metadata["client_id"]

	if ebookIDStr == "" || clientIDStr == "" {
		return fmt.Errorf("dados da compra inválidos")
	}

	ebookID, err := strconv.ParseUint(ebookIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("ebook ID inválido: %v", err)
	}

	clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("client ID inválido: %v", err)
	}

	purchase, err := h.purchaseService.CreatePurchaseWithResult(uint(ebookID), uint(clientID))
	if err != nil {
		return fmt.Errorf("erro ao criar/buscar compra: %v", err)
	}

	var purchaseWithRelations *salesmodel.Purchase
	if purchase.Ebook.ID == 0 || purchase.Client.ID == 0 {
		purchaseWithRelations, err = h.purchaseRepository.FindByID(purchase.ID)
		if err != nil {
			log.Printf("Erro ao buscar purchase com relacionamentos: %v", err)
			return fmt.Errorf("erro ao buscar dados da compra: %v", err)
		}
	} else {
		purchaseWithRelations = purchase
	}

	if purchaseWithRelations == nil {
		log.Printf("Purchase não encontrado após criação")
		return fmt.Errorf("purchase não encontrado após criação")
	}

	paymentIntentID := stripeSession.PaymentIntent.ID

	// A transação pendente foi criada durante o checkout (CreateEbookCheckout).
	// O webhook apenas a confirma — nunca cria uma segunda transação para a mesma purchase.
	if err := h.transactionService.UpdateTransactionToCompleted(purchase.ID, paymentIntentID); err != nil {
		log.Printf("Aviso: não foi possível atualizar transação para purchase_id=%d: %v", purchase.ID, err)
	} else {
		log.Printf("Transação atualizada para completed: purchase_id=%d", purchase.ID)
	}

	if purchaseWithRelations.Client.ID == 0 {
		log.Printf("Cliente não foi carregado! Client.ID=0")
	} else {
		log.Printf("Cliente carregado: ID=%d, Name='%s', Email='%s'",
			purchaseWithRelations.Client.ID,
			purchaseWithRelations.Client.Name,
			purchaseWithRelations.Client.Email)
	}

	if err := h.purchaseService.ConfirmPayment(purchase.ID); err != nil {
		log.Printf("Erro ao confirmar pagamento para purchase_id=%d: %v", purchase.ID, err)
	} else {
		log.Printf("Pagamento confirmado para purchase_id=%d", purchase.ID)
	}

	if purchaseWithRelations.Client.Email == "" {
		log.Printf("Cliente sem email: ClientID=%d", purchaseWithRelations.ClientID)
		return fmt.Errorf("cliente sem email válido")
	}

	log.Printf("Enviando email para: %s", purchaseWithRelations.Client.Email)

	go h.emailService.SendLinkToDownload([]*salesmodel.Purchase{purchaseWithRelations})

	return nil
}

// handleSubscriptionPayment processa pagamento de assinatura
func (h *StripeHandler) handleSubscriptionPayment(stripeSession stripe.CheckoutSession) error {
	subscription, err := h.subscriptionService.FindByStripeCustomerID(stripeSession.Customer.ID)
	if err != nil {
		return fmt.Errorf("error finding subscription: %v", err)
	}
	if subscription == nil {
		return fmt.Errorf("subscription not found for Stripe customer ID: %s", stripeSession.Customer.ID)
	}

	err = h.subscriptionService.ActivateSubscription(subscription.ID, stripeSession.Customer.ID, stripeSession.Subscription.ID)
	if err != nil {
		return fmt.Errorf("error updating subscription: %v", err)
	}

	return nil
}
