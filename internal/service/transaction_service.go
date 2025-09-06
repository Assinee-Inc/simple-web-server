package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type TransactionService interface {
	CreateTransaction(purchase *models.Purchase, totalAmount int64) (*models.Transaction, error)
	ProcessPaymentWithSplit(transaction *models.Transaction) error
	GetTransactionByID(id uint) (*models.Transaction, error)
	GetTransactionsByCreatorID(creatorID uint, page, limit int) ([]*models.Transaction, int64, error)
	CreateDirectTransaction(transaction *models.Transaction) error
	FindTransactionByPurchaseID(purchaseID uint) (*models.Transaction, error)
	UpdateTransactionToCompleted(purchaseID uint, stripePaymentIntentID string) error
}

type transactionServiceImpl struct {
	transactionRepo repository.TransactionRepository
	purchaseService *PurchaseService
	creatorService  CreatorService
	stripeService   *StripeService
}

func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	purchaseService *PurchaseService,
	creatorService CreatorService,
	stripeService *StripeService,
) TransactionService {
	// Configurar Stripe apenas uma vez
	if stripe.Key == "" {
		stripe.Key = config.AppConfig.StripeSecretKey
	}

	return &transactionServiceImpl{
		transactionRepo: transactionRepo,
		purchaseService: purchaseService,
		creatorService:  creatorService,
		stripeService:   stripeService,
	}
}

// CreateTransaction cria uma transação para o split de pagamento
func (s *transactionServiceImpl) CreateTransaction(purchase *models.Purchase, totalAmount int64) (*models.Transaction, error) {
	// Validar entrada
	if purchase == nil || purchase.ID == 0 {
		return nil, fmt.Errorf("compra inválida")
	}

	if totalAmount <= 0 {
		return nil, fmt.Errorf("valor de transação inválido")
	}

	// Buscar o criador do ebook
	slog.Debug("Buscando criador para split de pagamento", "ebookID", purchase.EbookID, "creatorID", purchase.Ebook.CreatorID)
	creator, err := s.creatorService.FindByID(purchase.Ebook.CreatorID)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		return nil, fmt.Errorf("erro ao buscar criador: %v", err)
	}

	// Verificar se o criador está habilitado para Stripe Connect
	if !creator.OnboardingCompleted || !creator.ChargesEnabled {
		slog.Debug("Criador não está habilitado para receber pagamentos",
			"creatorID", creator.ID,
			"onboardingCompleted", creator.OnboardingCompleted,
			"chargesEnabled", creator.ChargesEnabled)
		return nil, fmt.Errorf("criador não está habilitado para receber pagamentos")
	}

	// Criar a transação
	transaction := models.NewTransaction(purchase.ID, creator.ID, models.SplitTypePercentage)
	transaction.PlatformPercentage = config.Business.PlatformFeePercentage // Usa configuração centralizada
	transaction.CalculateSplit(totalAmount)

	// Logar detalhes do split (sem dados sensíveis)
	slog.Info("Split de pagamento calculado",
		"transactionID", transaction.ID,
		"totalAmount", transaction.TotalAmount,
		"platformAmount", transaction.PlatformAmount,
		"creatorAmount", transaction.CreatorAmount,
		"processingFee", transaction.StripeProcessingFee,
		"splitType", transaction.SplitType)

	// Salvar a transação
	err = s.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		slog.Error("Erro ao criar transação", "error", err)
		return nil, fmt.Errorf("erro ao criar transação: %v", err)
	}

	return transaction, nil
}

// ProcessPaymentWithSplit processa um pagamento com split usando Stripe Connect
func (s *transactionServiceImpl) ProcessPaymentWithSplit(transaction *models.Transaction) error {
	// Buscar o criador
	creator, err := s.creatorService.FindByID(transaction.CreatorID)
	if err != nil {
		slog.Error("Erro ao buscar criador para processamento de pagamento", "error", err)
		return fmt.Errorf("erro ao buscar criador: %v", err)
	}

	// Verificar se o criador tem conta Stripe Connect
	if creator.StripeConnectAccountID == "" {
		slog.Debug("Criador não possui conta Stripe Connect", "creatorID", creator.ID)
		return fmt.Errorf("criador não possui conta Stripe Connect")
	}

	// Buscar a compra
	purchase, err := s.purchaseService.GetPurchaseByID(transaction.PurchaseID)
	if err != nil {
		slog.Error("Erro ao buscar compra para processamento de pagamento", "error", err)
		return fmt.Errorf("erro ao buscar compra: %v", err)
	}

	// Evitar reprocessamento
	if transaction.Status == models.TransactionStatusCompleted {
		slog.Debug("Transação já processada, ignorando", "transactionID", transaction.ID)
		return nil
	}

	// Verificar se o ID da conta do criador está presente e válido
	if creator.StripeConnectAccountID == "" {
		slog.Error("Conta Stripe Connect do criador não encontrada", "creatorID", creator.ID)
		return fmt.Errorf("criador não possui uma conta Stripe Connect válida")
	}

	// Verificar se o criador está habilitado para receber pagamentos
	if !creator.OnboardingCompleted || !creator.ChargesEnabled {
		slog.Error("Criador não está habilitado para receber pagamentos",
			"creatorID", creator.ID,
			"onboardingCompleted", creator.OnboardingCompleted,
			"chargesEnabled", creator.ChargesEnabled)
		return fmt.Errorf("criador não está habilitado para receber pagamentos (onboarding incompleto ou charges desabilitadas)")
	}

	// Log das informações de split para diagnóstico
	slog.Info("Processando split de pagamento",
		"transactionID", transaction.ID,
		"purchaseID", transaction.PurchaseID,
		"totalAmount", transaction.TotalAmount,
		"platformAmount", transaction.PlatformAmount,
		"creatorAmount", transaction.CreatorAmount,
		"creatorConnectAccount", creator.StripeConnectAccountID,
		"creatorID", creator.ID)

	// Criar um PaymentIntent para capturar o pagamento com Destination Charge
	piParams := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(transaction.TotalAmount),
		Currency:    stripe.String("brl"),
		Description: stripe.String(fmt.Sprintf("Compra do e-book %s", purchase.Ebook.Title)),
		Metadata: map[string]string{
			"purchase_id":    fmt.Sprintf("%d", purchase.ID),
			"transaction_id": fmt.Sprintf("%d", transaction.ID),
			"creator_id":     fmt.Sprintf("%d", creator.ID),
			"creator_email":  creator.Email,
			"split_type":     string(transaction.SplitType),
			"platform_fee":   fmt.Sprintf("%d", transaction.PlatformAmount),
		},
		TransferGroup: stripe.String(fmt.Sprintf("purchase_%d", purchase.ID)),
		// Configuração do split de pagamento diretamente no PaymentIntent
		// Isso garantirá que o dinheiro vá diretamente para a conta do vendedor
		TransferData: &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(creator.StripeConnectAccountID),
			Amount:      stripe.Int64(transaction.CreatorAmount),
		},
		// Método de pagamento seria fornecido pelo frontend ou pelo cliente via API
	}

	slog.Debug("Criando PaymentIntent",
		"amount", transaction.TotalAmount,
		"purchaseID", purchase.ID,
		"destination", creator.StripeConnectAccountID,
		"creatorAmount", transaction.CreatorAmount)

	pi, err := paymentintent.New(piParams)
	if err != nil {
		transaction.Status = models.TransactionStatusFailed
		transaction.ErrorMessage = fmt.Sprintf("Erro ao criar intent: %v", err)
		s.transactionRepo.UpdateTransaction(transaction)
		slog.Error("Erro ao criar payment intent", "error", err)
		return fmt.Errorf("erro ao criar payment intent: %v", err)
	}

	// Salvar ID do PaymentIntent
	transaction.StripePaymentIntentID = pi.ID

	// Atualizar transação no banco para evitar perda de dados em caso de falha
	err = s.transactionRepo.UpdateTransaction(transaction)
	if err != nil {
		slog.Error("Erro ao atualizar transação com ID do PaymentIntent", "error", err)
		// Continuar processamento mesmo com o erro para evitar inconsistências
	}

	// Como usamos TransferData no PaymentIntent, a transferência é feita automaticamente
	// quando o pagamento é confirmado. Não precisamos criar uma transferência separada.
	slog.Debug("Pagamento com split configurado",
		"totalAmount", transaction.TotalAmount,
		"platformAmount", transaction.PlatformAmount,
		"creatorAmount", transaction.CreatorAmount)

	// Atualizar dados da transação
	transaction.Status = models.TransactionStatusCompleted
	now := time.Now()
	transaction.ProcessedAt = &now
	transaction.StripeTransferID = pi.ID // Usamos o ID do PaymentIntent como referência
	transaction.ErrorMessage = ""        // Limpar qualquer mensagem de erro anterior

	// Salvar transação atualizada
	err = s.transactionRepo.UpdateTransaction(transaction)
	if err != nil {
		slog.Error("Erro ao atualizar transação final", "error", err, "transactionID", transaction.ID)
		return fmt.Errorf("erro ao atualizar transação: %v", err)
	}

	slog.Info("Split de pagamento processado com sucesso",
		"transactionID", transaction.ID,
		"paymentIntentID", maskStripeID(pi.ID),
		"totalAmount", transaction.TotalAmount,
		"platformAmount", transaction.PlatformAmount,
		"creatorAmount", transaction.CreatorAmount,
		"creatorID", creator.ID,
		"stripeConnectAccountID", maskStripeID(creator.StripeConnectAccountID))

	return nil
}

// GetTransactionByID busca uma transação pelo ID
func (s *transactionServiceImpl) GetTransactionByID(id uint) (*models.Transaction, error) {
	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		slog.Error("Erro ao buscar transação", "id", id, "error", err)
		return nil, err
	}
	return transaction, nil
}

// GetTransactionsByCreatorID busca transações por ID do criador
func (s *transactionServiceImpl) GetTransactionsByCreatorID(creatorID uint, page, limit int) ([]*models.Transaction, int64, error) {
	return s.transactionRepo.FindByCreatorID(creatorID, page, limit)
}

// CreateDirectTransaction cria uma transação diretamente sem processar pagamento
// Útil para registrar transações de pagamentos que já foram processados diretamente pelo Stripe
func (s *transactionServiceImpl) CreateDirectTransaction(transaction *models.Transaction) error {
	return s.transactionRepo.CreateTransaction(transaction)
}

// FindTransactionByPurchaseID busca uma transação pelo ID da compra
func (s *transactionServiceImpl) FindTransactionByPurchaseID(purchaseID uint) (*models.Transaction, error) {
	return s.transactionRepo.FindByPurchaseID(purchaseID)
}

// UpdateTransactionToCompleted atualiza uma transação para status completada
func (s *transactionServiceImpl) UpdateTransactionToCompleted(purchaseID uint, stripePaymentIntentID string) error {
	// Buscar transação existente
	transaction, err := s.transactionRepo.FindByPurchaseID(purchaseID)
	if err != nil {
		return fmt.Errorf("erro ao buscar transação: %v", err)
	}

	if transaction == nil {
		return fmt.Errorf("transação não encontrada para purchase_id: %d", purchaseID)
	}

	// Atualizar status se ainda não estiver completada
	if transaction.Status != models.TransactionStatusCompleted {
		transaction.Status = models.TransactionStatusCompleted
		transaction.StripePaymentIntentID = stripePaymentIntentID
		now := time.Now()
		transaction.ProcessedAt = &now

		return s.transactionRepo.UpdateTransaction(transaction)
	}

	// Se já estiver completada, apenas log e retorna sucesso
	slog.Debug("Transação já completada", "transactionID", transaction.ID)
	return nil
}

// maskStripeID mascara IDs sensíveis do Stripe para logging seguro
func maskStripeID(id string) string {
	if len(id) <= 8 {
		return "****"
	}
	return id[:4] + "****" + id[len(id)-4:]
}
