package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/internal/repository/gorm"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/gov"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

type CheckoutHandler struct {
	templateRenderer template.TemplateRenderer
	ebookService     service.EbookService
	clientService    service.ClientService
	creatorService   service.CreatorService
	rfService        gov.ReceitaFederalService
	emailService     *service.EmailService
}

func NewCheckoutHandler(
	templateRenderer template.TemplateRenderer,
	ebookService service.EbookService,
	clientService service.ClientService,
	creatorService service.CreatorService,
	rfService gov.ReceitaFederalService,
	emailService *service.EmailService,
) *CheckoutHandler {
	return &CheckoutHandler{
		templateRenderer: templateRenderer,
		ebookService:     ebookService,
		clientService:    clientService,
		creatorService:   creatorService,
		rfService:        rfService,
		emailService:     emailService,
	}
}

// CheckoutView exibe a página de checkout para o ebook
func (h *CheckoutHandler) CheckoutView(w http.ResponseWriter, r *http.Request) {
	ebookIDStr := chi.URLParam(r, "id")
	if ebookIDStr == "" {
		http.Error(w, "ID do ebook não fornecido", http.StatusBadRequest)
		return
	}

	ebookID, err := strconv.ParseUint(ebookIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID do ebook inválido", http.StatusBadRequest)
		return
	}

	// Buscar o ebook
	ebook, err := h.ebookService.FindByID(uint(ebookID))
	if err != nil {
		log.Printf("Erro ao buscar ebook: %v", err)
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	if ebook == nil {
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	// Verificar se o ebook está ativo
	if !ebook.Status {
		http.Error(w, "Ebook não disponível", http.StatusNotFound)
		return
	}

	// Buscar o criador do ebook
	creator, err := h.creatorService.FindByID(ebook.CreatorID)
	if err != nil {
		log.Printf("Erro ao buscar criador do ebook: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	// Atualizar o ebook com os dados do criador
	ebook.Creator = *creator

	// Preparar dados para o template
	data := map[string]any{
		"Ebook": ebook,
	}

	h.templateRenderer.View(w, r, "checkout", data, "guest")
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

	// Validar dados obrigatórios
	if request.Name == "" || request.CPF == "" || request.Birthdate == "" || request.Email == "" || request.Phone == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Todos os campos são obrigatórios",
		})
		return
	}

	// Validar CPF
	if len(request.CPF) != 11 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "CPF inválido",
		})
		return
	}

	// Validar email
	if !isValidEmail(request.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "E-mail inválido",
		})
		return
	}

	// Validar telefone (formato: XXXXXXXXXXX = 11 caracteres)
	if len(request.Phone) != 11 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Telefone inválido",
		})
		return
	}

	// Validar ebook
	ebookID, err := strconv.ParseUint(request.EbookID, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Ebook inválido",
		})
		return
	}

	ebook, err := h.ebookService.FindByID(uint(ebookID))
	if err != nil || ebook == nil || !ebook.Status {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Ebook não encontrado ou indisponível",
		})
		return
	}

	// Validar com Receita Federal
	if h.rfService != nil {
		response, err := h.rfService.ConsultaCPF(request.CPF, request.Birthdate)
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

		// Verificar se o nome confere
		if !isNameSimilar(request.Name, response.Result.NomeDaPF) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"error":   "Nome não confere com os dados da Receita Federal",
			})
			return
		}
	}

	// Dados válidos
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Dados validados com sucesso",
	})
}

// CreateEbookCheckout cria uma sessão de checkout no Stripe para o ebook
func (h *CheckoutHandler) CreateEbookCheckout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Initialize Stripe
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

	// Validar ebook
	ebookID, err := strconv.ParseUint(request.EbookID, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Ebook inválido",
		})
		return
	}

	ebook, err := h.ebookService.FindByID(uint(ebookID))
	if err != nil || ebook == nil || !ebook.Status {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Ebook não encontrado ou indisponível",
		})
		return
	}

	// Buscar o criador do ebook
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

	// Criar ou buscar cliente
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

	// Criar sessão do Stripe
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("brl"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(ebook.Title),
						Description: stripe.String(ebook.Description),
					},
					UnitAmount: stripe.Int64(int64(ebook.Value * 100)), // Stripe usa centavos
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL:    stripe.String("http://" + r.Host + "/purchase/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:     stripe.String("http://" + r.Host + "/checkout/" + request.EbookID),
		CustomerEmail: stripe.String(request.Email),
		Metadata: map[string]string{
			"ebook_id":    request.EbookID,
			"client_id":   strconv.FormatUint(uint64(client.ID), 10),
			"creator_id":  strconv.FormatUint(uint64(creator.ID), 10),
			"client_name": request.Name,
			"client_cpf":  request.CPF,
		},
	}

	session, err := session.New(params)
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
		"url":     session.URL,
	})
}

// PurchaseSuccessView exibe a página de sucesso da compra
func (h *CheckoutHandler) PurchaseSuccessView(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Sessão não encontrada", http.StatusBadRequest)
		return
	}

	// Buscar dados da sessão do Stripe
	session, err := session.Get(sessionID, nil)
	if err != nil {
		log.Printf("Erro ao buscar sessão do Stripe: %v", err)
		http.Error(w, "Sessão inválida", http.StatusBadRequest)
		return
	}

	// Verificar se o pagamento foi realizado
	if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		http.Error(w, "Pagamento não confirmado", http.StatusBadRequest)
		return
	}

	// Extrair dados da sessão
	ebookIDStr := session.Metadata["ebook_id"]
	clientIDStr := session.Metadata["client_id"]
	creatorIDStr := session.Metadata["creator_id"]

	if ebookIDStr == "" || clientIDStr == "" {
		http.Error(w, "Dados da compra inválidos", http.StatusBadRequest)
		return
	}

	// Buscar ebook
	ebookID, _ := strconv.ParseUint(ebookIDStr, 10, 32)
	ebook, err := h.ebookService.FindByID(uint(ebookID))
	if err != nil || ebook == nil {
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	// Buscar criador
	creatorID, _ := strconv.ParseUint(creatorIDStr, 10, 32)
	creator, err := h.creatorService.FindByID(uint(creatorID))
	if err != nil || creator == nil {
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	// Buscar cliente (usando repository diretamente)
	clientID, _ := strconv.ParseUint(clientIDStr, 10, 32)
	clientRepo := gorm.NewClientGormRepository()
	client := &models.Client{}
	err = clientRepo.FindByIDAndCreators(client, uint(clientID), "")
	if err != nil || client.ID == 0 {
		http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		return
	}

	// Criar registro de compra
	purchase := models.NewPurchase(uint(ebookID), uint(clientID))
	purchase.ExpiresAt = time.Now().AddDate(0, 0, 30) // 30 dias de acesso

	purchaseRepo := repository.NewPurchaseRepository()
	err = purchaseRepo.CreateManyPurchases([]*models.Purchase{purchase})
	if err != nil {
		log.Printf("Erro ao criar compra: %v", err)
		// Não retornar erro para o usuário, apenas log
	}

	log.Printf("[checkout_handler] DADOS DA COMPRA: %+v", purchase)
	log.Printf("[checkout_handler] 📧 Enviando email para: %s", purchase.Client.Email)

	// Enviar email com link de download
	if purchase.ID > 0 {
		go h.emailService.SendLinkToDownload([]*models.Purchase{purchase})
	}

	// Preparar dados para o template
	data := map[string]any{
		"Ebook":         ebook,
		"CustomerEmail": client.Email,
		"CreatorEmail":  creator.Email,
		"Purchase":      purchase,
	}

	h.templateRenderer.View(w, r, "purchase-success", data, "guest")
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
}, creatorID uint) (*models.Client, error) {
	clientRepo := gorm.NewClientGormRepository()

	// Buscar cliente existente por email
	existingClient, err := clientRepo.FindByEmail(request.Email)
	if err == nil && existingClient != nil {
		log.Printf("Cliente existente encontrado: ID=%d, Email='%s'", existingClient.ID, existingClient.Email)

		// Atualizar email se necessário
		if existingClient.Email != request.Email {
			log.Printf("Atualizando email do cliente: '%s' -> '%s'", existingClient.Email, request.Email)
			existingClient.Email = request.Email
			err = clientRepo.Save(existingClient)
			if err != nil {
				log.Printf("Erro ao atualizar email do cliente: %v", err)
			}
		}
		return existingClient, nil
	}

	// Criar novo cliente
	client := &models.Client{
		Name:  request.Name,
		CPF:   request.CPF,
		Email: request.Email,
		Phone: request.Phone,
	}

	// Parse birthdate
	birthDate, err := time.Parse("02/01/2006", request.Birthdate)
	if err != nil {
		return nil, err
	}
	client.Birthdate = birthDate.Format("2006-01-02")

	log.Printf("Criando novo cliente: Name='%s', Email='%s', Phone='%s'",
		client.Name, client.Email, client.Phone)

	// Salvar cliente
	err = clientRepo.Save(client)
	if err != nil {
		log.Printf("Erro ao salvar cliente: %v", err)
		return nil, err
	}

	log.Printf("Cliente criado com sucesso: ID=%d, Email='%s'", client.ID, client.Email)

	return client, nil
}

// Funções auxiliares
func isValidEmail(email string) bool {
	// Implementação simples de validação de email
	return len(email) > 3 && len(email) < 254
}

func isNameSimilar(name1, name2 string) bool {
	// Implementação simples de comparação de nomes
	// Em produção, usar algoritmo mais sofisticado
	return len(name1) > 0 && len(name2) > 0
}
