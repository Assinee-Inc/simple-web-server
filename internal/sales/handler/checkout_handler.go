package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	accountrepo "github.com/anglesson/simple-web-server/internal/account/repository"
	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	"github.com/anglesson/simple-web-server/internal/config"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	librarysvc "github.com/anglesson/simple-web-server/internal/library/service"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesrepogorm "github.com/anglesson/simple-web-server/internal/sales/repository/gorm"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/anglesson/simple-web-server/pkg/gov"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

type CheckoutHandler struct {
	templateRenderer   template.TemplateRenderer
	ebookService       librarysvc.EbookService
	clientService      salesvc.ClientService
	clientRepo         salesrepo.ClientRepository
	creatorService     accountsvc.CreatorService
	rfService          gov.ReceitaFederalService
	emailService       salesvc.IEmailService
	transactionService salesvc.TransactionService
	purchaseService    salesvc.PurchaseService
}

func NewCheckoutHandler(
	templateRenderer template.TemplateRenderer,
	ebookService librarysvc.EbookService,
	clientService salesvc.ClientService,
	clientRepo salesrepo.ClientRepository,
	creatorService accountsvc.CreatorService,
	rfService gov.ReceitaFederalService,
	emailService salesvc.IEmailService,
	transactionService salesvc.TransactionService,
	purchaseService salesvc.PurchaseService,
) *CheckoutHandler {
	return &CheckoutHandler{
		templateRenderer:   templateRenderer,
		ebookService:       ebookService,
		clientService:      clientService,
		clientRepo:         clientRepo,
		creatorService:     creatorService,
		rfService:          rfService,
		emailService:       emailService,
		transactionService: transactionService,
		purchaseService:    purchaseService,
	}
}

// CheckoutView exibe a página de checkout para o ebook
func (h *CheckoutHandler) CheckoutView(w http.ResponseWriter, r *http.Request) {
	ebookPublicID := chi.URLParam(r, "id")
	if ebookPublicID == "" {
		http.Error(w, "ID do ebook não fornecido", http.StatusBadRequest)
		return
	}

	ebook, err := h.ebookService.FindByPublicID(ebookPublicID)
	if err != nil {
		log.Printf("Erro ao buscar ebook: %v", err)
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	if ebook == nil {
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	if !ebook.Status {
		http.Error(w, "Ebook não disponível", http.StatusNotFound)
		return
	}

	creator, err := h.creatorService.FindByID(ebook.CreatorID)
	if err != nil {
		log.Printf("Erro ao buscar criador do ebook: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Ebook":   ebook,
		"Creator": creator,
	}

	h.templateRenderer.View(w, r, "purchase/checkout", data, "guest")
}

// ValidateCustomer valida os dados do cliente com a Receita Federal
func (h *CheckoutHandler) ValidateCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request struct {
		Name      string `json:"name"`
		CPF       string `json:"cpf"`
		Birthdate string `json:"birthdate"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		EbookID   string `json:"ebookId"`
		CSRFToken string `json:"csrfToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Erro ao decodificar requisição: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Dados inválidos",
		})
		return
	}

	if request.Name == "" || request.CPF == "" || request.Birthdate == "" || request.Email == "" || request.Phone == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Todos os campos são obrigatórios",
		})
		return
	}

	if len(request.CPF) != 11 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "CPF inválido",
		})
		return
	}

	if !isValidEmail(request.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "E-mail inválido",
		})
		return
	}

	if len(request.Phone) != 11 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Telefone inválido",
		})
		return
	}

	if request.EbookID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Ebook inválido",
		})
		return
	}

	ebook, err := h.ebookService.FindByPublicID(request.EbookID)
	if err != nil || ebook == nil || !ebook.Status {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Ebook não encontrado ou indisponível",
		})
		return
	}

	existingClient, err := h.clientRepo.FindByCPF(request.CPF)
	if err == nil && existingClient != nil {
		existingPurchase, err := h.purchaseService.FindExistingPurchase(ebook.ID, existingClient.ID)
		if err == nil && existingPurchase != nil {
			creator, _ := h.creatorService.FindByID(ebook.CreatorID)
			creatorEmail := ""
			creatorName := ""
			if creator != nil {
				creatorEmail = creator.Email
				creatorName = creator.Name
			}
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]any{
				"success":          false,
				"already_purchased": true,
				"error":            "Você já adquiriu este ebook com o CPF informado.",
				"creator_email":    creatorEmail,
				"creator_name":     creatorName,
			})
			return
		}
	}

	if h.rfService != nil && config.AppConfig.IsProduction() {
		response, err := h.rfService.ConsultaCPF(request.Name, request.CPF, request.Birthdate)
		if err != nil {
			log.Printf("Erro na consulta da Receita Federal: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"error":   "Erro na validação dos dados. Tente novamente.",
			})
			return
		}

		if !response.Status {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"error":   "Dados não conferem com a Receita Federal",
			})
			return
		}

		if !isNameSimilar(request.Name, response.Result.NomeDaPF) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"error":   "Nome não confere com os dados da Receita Federal",
			})
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Dados validados com sucesso",
	})
}

// CreateEbookCheckout cria uma sessão de checkout no Stripe para o ebook
func (h *CheckoutHandler) CreateEbookCheckout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stripe.Key = config.AppConfig.StripeSecretKey

	var request struct {
		Name      string `json:"name"`
		CPF       string `json:"cpf"`
		Birthdate string `json:"birthdate"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		EbookID   string `json:"ebookId"`
		CSRFToken string `json:"csrfToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Erro ao decodificar requisição: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Dados inválidos",
		})
		return
	}

	if request.EbookID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Ebook inválido",
		})
		return
	}

	ebook, err := h.ebookService.FindByPublicID(request.EbookID)
	if err != nil || ebook == nil || !ebook.Status {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Ebook não encontrado ou indisponível",
		})
		return
	}

	creator, err := h.creatorService.FindByID(ebook.CreatorID)
	if err != nil {
		log.Printf("Erro ao buscar criador: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Erro interno do servidor",
		})
		return
	}

	client, err := h.createOrFindClient(request, creator.ID)
	if err != nil {
		log.Printf("Erro ao criar/buscar cliente: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Erro ao processar dados do cliente",
		})
		return
	}

	purchase, err := h.purchaseService.CreatePurchaseWithResult(ebook.ID, client.ID)
	if err != nil {
		log.Printf("Erro ao criar/buscar compra pendente: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Erro ao processar compra",
		})
		return
	}

	if purchase != nil {
		log.Printf("Purchase processada com sucesso: ID=%d para EbookID=%d, ClientID=%d", purchase.ID, ebook.ID, client.ID)

		existingTransaction, _ := h.transactionService.FindTransactionByPurchaseID(purchase.ID)
		if existingTransaction == nil {
			transaction := salesmodel.NewTransaction(purchase.ID, creator.ID, salesmodel.SplitTypeFixedAmount)
			transaction.PlatformPercentage = config.Business.PlatformFeePercentage
			transaction.CalculateSplit(int64(ebook.GetFinalValue() * 100))
			transaction.Status = salesmodel.TransactionStatusPending

			err = h.transactionService.CreateDirectTransaction(transaction)
			if err != nil {
				log.Printf("Erro ao criar transação pendente: %v", err)
			} else {
				log.Printf("Transação pendente criada com sucesso: ID=%d, PurchaseID=%d", transaction.ID, purchase.ID)
			}
		} else {
			log.Printf("Transação já existe para PurchaseID=%d: ID=%d", purchase.ID, existingTransaction.ID)
		}
	}

	host := fmt.Sprintf("%s:%s", config.AppConfig.Host, config.AppConfig.Port)

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyBRL)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(ebook.Title),
						Description: stripe.String(ebook.Description),
					},
					UnitAmount: stripe.Int64(int64(ebook.GetFinalValue() * 100)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL:    stripe.String(host + "/purchase/success?session_id={CHECKOUT_SESSION_ID}&creator_id=" + strconv.FormatUint(uint64(creator.ID), 10)),
		CancelURL:     stripe.String(host + "/checkout/" + ebook.PublicID),
		CustomerEmail: stripe.String(request.Email),
		Metadata: map[string]string{
			"ebook_id":        strconv.FormatUint(uint64(ebook.ID), 10),
			"client_id":       strconv.FormatUint(uint64(client.ID), 10),
			"creator_id":      strconv.FormatUint(uint64(creator.ID), 10),
			"client_name":     request.Name,
			"client_cpf":      request.CPF,
			"ebook_title":     ebook.Title,
			"ebook_price":     strconv.FormatFloat(ebook.GetFinalValue(), 'f', 2, 64),
			"payment_version": "2.0",
		},
	}

	if purchase != nil && purchase.ID > 0 {
		params.Metadata["purchase_id"] = strconv.FormatUint(uint64(purchase.ID), 10)
	}

	if creator.StripeConnectAccountID != "" && creator.OnboardingCompleted && creator.ChargesEnabled {
		log.Printf("Criador tem conta Stripe Connect habilitada: ID=%d, Nome=%s, Conta=%s",
			creator.ID, creator.Name, creator.StripeConnectAccountID)

		platformFeeAmount := config.Business.GetPlatformFeeAmount(int64(ebook.GetFinalValue() * 100))
		creatorAmount := int64(ebook.GetFinalValue()*100) - platformFeeAmount

		log.Printf("Divisão do pagamento: Total=%d centavos | Plataforma=%d centavos | Criador=%d centavos",
			int64(ebook.GetFinalValue()*100), platformFeeAmount, creatorAmount)

		params.PaymentIntentData = &stripe.CheckoutSessionPaymentIntentDataParams{
			ApplicationFeeAmount: stripe.Int64(platformFeeAmount),
			Metadata: map[string]string{
				"fee_percent":     config.Business.PlatformFeePercentageDisplay,
				"payment_type":    "direct_to_creator",
				"creator_account": creator.StripeConnectAccountID,
				"platform_fee":    strconv.FormatInt(platformFeeAmount, 10),
				"creator_amount":  strconv.FormatInt(creatorAmount, 10),
			},
		}
	} else {
		log.Printf("Criador não tem conta Stripe Connect habilitada: ID=%d, Nome=%s, Conta=%s, OnboardingCompleted=%t, ChargesEnabled=%t",
			creator.ID, creator.Name, creator.StripeConnectAccountID, creator.OnboardingCompleted, creator.ChargesEnabled)

		params.Metadata["payment_type"] = "platform_only"
	}

	params.SetStripeAccount(creator.StripeConnectAccountID)
	s, err := session.New(params)
	if err != nil {
		log.Printf("Erro ao criar sessão do Stripe: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Erro ao processar pagamento",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"url":     s.URL,
	})
}

// PurchaseSuccessView exibe a página de sucesso da compra
func (h *CheckoutHandler) PurchaseSuccessView(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Sessão não encontrada", http.StatusBadRequest)
		return
	}

	creatorIDFromURL := r.URL.Query().Get("creator_id")
	creator, err := h.creatorService.FindByID(func() uint {
		if id, err := strconv.ParseUint(creatorIDFromURL, 10, 32); err == nil {
			return uint(id)
		}
		return 0
	}())
	if err != nil || creator == nil {
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	var sessionParams *stripe.CheckoutSessionParams
	if creator.StripeConnectAccountID != "" {
		sessionParams = &stripe.CheckoutSessionParams{}
		sessionParams.SetStripeAccount(creator.StripeConnectAccountID)
	}

	s, err := session.Get(sessionID, sessionParams)
	if err != nil {
		log.Printf("Erro ao buscar sessão do Stripe: %v", err)
		http.Error(w, "Sessão inválida", http.StatusBadRequest)
		return
	}

	if s.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		http.Error(w, "Pagamento não confirmado", http.StatusBadRequest)
		return
	}

	ebookIDStr := s.Metadata["ebook_id"]
	clientIDStr := s.Metadata["client_id"]

	if ebookIDStr == "" || clientIDStr == "" {
		http.Error(w, "Dados da compra inválidos", http.StatusBadRequest)
		return
	}

	ebookID, _ := strconv.ParseUint(ebookIDStr, 10, 32)
	ebook, err := h.ebookService.FindByID(uint(ebookID))
	if err != nil || ebook == nil {
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	clientID, _ := strconv.ParseUint(clientIDStr, 10, 32)
	clientRepo := salesrepogorm.NewClientGormRepository()
	client := &salesmodel.Client{}
	err = clientRepo.FindByIDAndCreators(client, uint(clientID), "")
	if err != nil || client.ID == 0 {
		http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		return
	}

	purchase, err := h.purchaseService.CreatePurchaseWithResult(uint(ebookID), uint(clientID))
	if err != nil {
		log.Printf("Erro ao criar/buscar compra: %v", err)
		purchase = salesmodel.NewPurchase(uint(ebookID), uint(clientID), utils.UuidV7())
		purchase.ExpiresAt = time.Now().AddDate(0, 0, 30)
	} else {
		log.Printf("Purchase processada com sucesso: ID=%d para EbookID=%d, ClientID=%d", purchase.ID, ebookID, clientID)
	}

	log.Printf("[checkout_handler] DADOS DA COMPRA: %+v", purchase)
	log.Printf("[checkout_handler] Enviando email para: %s", purchase.Client.Email)

	if purchase.ID > 0 {
		h.recordStripePayment(purchase.ID, creator.ID, ebook, s.PaymentIntent.ID)
	}

	if purchase.ID > 0 {
		go h.emailService.SendLinkToDownload([]*salesmodel.Purchase{purchase})
	}

	data := map[string]any{
		"Ebook":         ebook,
		"CustomerEmail": client.Email,
		"CreatorEmail":  creator.Email,
		"Purchase":      purchase,
	}

	h.templateRenderer.View(w, r, "purchase/purchase-success", data, "guest")
}

// createOrFindClient cria ou busca um cliente existente
func (h *CheckoutHandler) createOrFindClient(request struct {
	Name      string `json:"name"`
	CPF       string `json:"cpf"`
	Birthdate string `json:"birthdate"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	EbookID   string `json:"ebookId"`
	CSRFToken string `json:"csrfToken"`
}, creatorID uint) (*salesmodel.Client, error) {
	existingClient, err := h.clientRepo.FindByEmail(request.Email)
	if err == nil && existingClient != nil {
		log.Printf("Cliente existente encontrado: ID=%d, Email='%s'", existingClient.ID, existingClient.Email)

		if existingClient.Email != request.Email {
			log.Printf("Atualizando email do cliente: '%s' -> '%s'", existingClient.Email, request.Email)
			existingClient.Email = request.Email
			err = h.clientRepo.Save(existingClient)
			if err != nil {
				log.Printf("Erro ao atualizar email do cliente: %v", err)
			}
		}
		return existingClient, nil
	}

	creatorRepo := accountrepo.NewGormCreatorRepository(database.DB)
	creator, err := creatorRepo.FindByID(creatorID)
	if err != nil {
		log.Printf("Erro ao buscar criador: %v", err)
		return nil, fmt.Errorf("erro ao buscar criador: %v", err)
	}

	birthDate, err := time.Parse("02/01/2006", request.Birthdate)
	if err != nil {
		return nil, err
	}

	client := salesmodel.NewClient(request.Name, request.CPF, birthDate.Format("2006-01-02"), request.Email, request.Phone, creator)

	log.Printf("Criando novo cliente: Name='%s', Email='%s', Phone='%s', associado ao Creator ID=%d",
		client.Name, client.Email, client.Phone, creator.ID)

	err = h.clientRepo.Save(client)
	if err != nil {
		log.Printf("Erro ao salvar cliente: %v", err)
		return nil, err
	}

	log.Printf("Cliente criado com sucesso: ID=%d, Email='%s'", client.ID, client.Email)

	return client, nil
}

// recordStripePayment garante que o payment intent do Stripe seja registrado no banco.
// Se a purchase já tem uma transação completed com um payment intent diferente
// (pagamento duplicado), cria uma nova transação para trilha de auditoria.
// Se a transação está pendente, atualiza para completed.
func (h *CheckoutHandler) recordStripePayment(purchaseID uint, creatorID uint, ebook *librarymodel.Ebook, paymentIntentID string) {
	existingTx, _ := h.transactionService.FindTransactionByPurchaseID(purchaseID)
	if existingTx != nil && existingTx.Status == salesmodel.TransactionStatusCompleted && existingTx.StripePaymentIntentID != paymentIntentID {
		log.Printf("Alerta: payment intent %s para purchase_id=%d já tem transação completada (ID=%d, intent=%s). Criando nova transação.",
			paymentIntentID, purchaseID, existingTx.ID, existingTx.StripePaymentIntentID)
		newTx := salesmodel.NewTransaction(purchaseID, creatorID, salesmodel.SplitTypeFixedAmount)
		newTx.PlatformPercentage = config.Business.PlatformFeePercentage
		newTx.CalculateSplit(int64(ebook.GetFinalValue() * 100))
		newTx.Status = salesmodel.TransactionStatusCompleted
		newTx.StripePaymentIntentID = paymentIntentID
		now := time.Now()
		newTx.ProcessedAt = &now
		if err := h.transactionService.CreateDirectTransaction(newTx); err != nil {
			log.Printf("Erro crítico: não foi possível registrar transação para payment intent %s: %v", paymentIntentID, err)
		} else {
			log.Printf("Nova transação registrada: ID=%d, PaymentIntent=%s, PurchaseID=%d", newTx.ID, paymentIntentID, purchaseID)
		}
		return
	}
	err := h.transactionService.UpdateTransactionToCompleted(purchaseID, paymentIntentID)
	if err != nil {
		log.Printf("Erro crítico: Não foi possível atualizar transação para purchase_id=%d: %v", purchaseID, err)
	} else {
		log.Printf("Transação atualizada para completed: purchase_id=%d, payment_intent=%s", purchaseID, paymentIntentID)
	}
}

func isValidEmail(email string) bool {
	return len(email) > 3 && len(email) < 254
}

func isNameSimilar(name1, name2 string) bool {
	return len(name1) > 0 && len(name2) > 0
}
