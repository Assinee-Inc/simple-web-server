package service

import (
	"fmt"
	"log/slog"
	"time"

	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	"github.com/anglesson/simple-web-server/internal/config"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	subscriptionservice "github.com/anglesson/simple-web-server/internal/subscription/service"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type TransactionService interface {
	CreateTransaction(purchase *salesmodel.Purchase, totalAmount int64) (*salesmodel.Transaction, error)
	ProcessPaymentWithSplit(transaction *salesmodel.Transaction) error
	GetTransactionByID(id uint) (*salesmodel.Transaction, error)
	GetTransactionsByCreatorID(creatorID uint, page, limit int) ([]*salesmodel.Transaction, int64, error)
	GetTransactionsByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*salesmodel.Transaction, int64, error)
	CreateDirectTransaction(transaction *salesmodel.Transaction) error
	FindTransactionByPurchaseID(purchaseID uint) (*salesmodel.Transaction, error)
	UpdateTransactionToCompleted(purchaseID uint, stripePaymentIntentID string) error
}

type transactionServiceImpl struct {
	transactionRepo salesrepo.TransactionRepository
	purchaseService PurchaseService
	creatorService  accountsvc.CreatorService
	stripeService   *subscriptionservice.StripeService
}

func NewTransactionService(
	transactionRepo salesrepo.TransactionRepository,
	purchaseService PurchaseService,
	creatorService accountsvc.CreatorService,
	stripeService *subscriptionservice.StripeService,
) TransactionService {
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

func (s *transactionServiceImpl) CreateTransaction(purchase *salesmodel.Purchase, totalAmount int64) (*salesmodel.Transaction, error) {
	if purchase == nil || purchase.ID == 0 {
		return nil, fmt.Errorf("compra inválida")
	}

	if totalAmount <= 0 {
		return nil, fmt.Errorf("valor de transação inválido")
	}

	slog.Debug("Buscando criador para split de pagamento", "ebookID", purchase.EbookID, "creatorID", purchase.Ebook.CreatorID)
	creator, err := s.creatorService.FindByID(purchase.Ebook.CreatorID)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		return nil, fmt.Errorf("erro ao buscar criador: %v", err)
	}

	if !creator.OnboardingCompleted || !creator.ChargesEnabled {
		slog.Debug("Criador não está habilitado para receber pagamentos",
			"creatorID", creator.ID,
			"onboardingCompleted", creator.OnboardingCompleted,
			"chargesEnabled", creator.ChargesEnabled)
		return nil, fmt.Errorf("criador não está habilitado para receber pagamentos")
	}

	transaction := salesmodel.NewTransaction(purchase.ID, creator.ID, salesmodel.SplitTypePercentage)
	transaction.PlatformPercentage = config.Business.PlatformFeePercentage
	transaction.PlatformFixedFee = config.Business.PlatformFixedFeeCents
	transaction.CalculateSplit(totalAmount)

	slog.Info("Split de pagamento calculado",
		"transactionID", transaction.ID,
		"totalAmount", transaction.TotalAmount,
		"platformAmount", transaction.PlatformAmount,
		"creatorAmount", transaction.CreatorAmount,
		"processingFee", transaction.StripeProcessingFee,
		"splitType", transaction.SplitType)

	err = s.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		slog.Error("Erro ao criar transação", "error", err)
		return nil, fmt.Errorf("erro ao criar transação: %v", err)
	}

	return transaction, nil
}

func (s *transactionServiceImpl) ProcessPaymentWithSplit(transaction *salesmodel.Transaction) error {
	creator, err := s.creatorService.FindByID(transaction.CreatorID)
	if err != nil {
		slog.Error("Erro ao buscar criador para processamento de pagamento", "error", err)
		return fmt.Errorf("erro ao buscar criador: %v", err)
	}

	if creator.StripeConnectAccountID == "" {
		slog.Debug("Criador não possui conta Stripe Connect", "creatorID", creator.ID)
		return fmt.Errorf("criador não possui conta Stripe Connect")
	}

	purchase, err := s.purchaseService.GetPurchaseByID(transaction.PurchaseID)
	if err != nil {
		slog.Error("Erro ao buscar compra para processamento de pagamento", "error", err)
		return fmt.Errorf("erro ao buscar compra: %v", err)
	}

	if transaction.Status == salesmodel.TransactionStatusCompleted {
		slog.Debug("Transação já processada, ignorando", "transactionID", transaction.ID)
		return nil
	}

	if creator.StripeConnectAccountID == "" {
		slog.Error("Conta Stripe Connect do criador não encontrada", "creatorID", creator.ID)
		return fmt.Errorf("criador não possui uma conta Stripe Connect válida")
	}

	if !creator.OnboardingCompleted || !creator.ChargesEnabled {
		slog.Error("Criador não está habilitado para receber pagamentos",
			"creatorID", creator.ID,
			"onboardingCompleted", creator.OnboardingCompleted,
			"chargesEnabled", creator.ChargesEnabled)
		return fmt.Errorf("criador não está habilitado para receber pagamentos (onboarding incompleto ou charges desabilitadas)")
	}

	slog.Info("Processando split de pagamento",
		"transactionID", transaction.ID,
		"purchaseID", transaction.PurchaseID,
		"totalAmount", transaction.TotalAmount,
		"platformAmount", transaction.PlatformAmount,
		"creatorAmount", transaction.CreatorAmount,
		"creatorConnectAccount", creator.StripeConnectAccountID,
		"creatorID", creator.ID)

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
		TransferData: &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(creator.StripeConnectAccountID),
			Amount:      stripe.Int64(transaction.CreatorAmount),
		},
	}

	slog.Debug("Criando PaymentIntent",
		"amount", transaction.TotalAmount,
		"purchaseID", purchase.ID,
		"destination", creator.StripeConnectAccountID,
		"creatorAmount", transaction.CreatorAmount)

	pi, err := paymentintent.New(piParams)
	if err != nil {
		transaction.Status = salesmodel.TransactionStatusFailed
		transaction.ErrorMessage = fmt.Sprintf("Erro ao criar intent: %v", err)
		s.transactionRepo.UpdateTransaction(transaction)
		slog.Error("Erro ao criar payment intent", "error", err)
		return fmt.Errorf("erro ao criar payment intent: %v", err)
	}

	transaction.StripePaymentIntentID = pi.ID

	err = s.transactionRepo.UpdateTransaction(transaction)
	if err != nil {
		slog.Error("Erro ao atualizar transação com ID do PaymentIntent", "error", err)
	}

	slog.Debug("Pagamento com split configurado",
		"totalAmount", transaction.TotalAmount,
		"platformAmount", transaction.PlatformAmount,
		"creatorAmount", transaction.CreatorAmount)

	transaction.Status = salesmodel.TransactionStatusCompleted
	now := time.Now()
	transaction.ProcessedAt = &now
	transaction.StripeTransferID = pi.ID
	transaction.ErrorMessage = ""

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

func (s *transactionServiceImpl) GetTransactionByID(id uint) (*salesmodel.Transaction, error) {
	transaction, err := s.transactionRepo.FindByID(id)
	if err != nil {
		slog.Error("Erro ao buscar transação", "id", id, "error", err)
		return nil, err
	}
	return transaction, nil
}

func (s *transactionServiceImpl) GetTransactionsByCreatorID(creatorID uint, page, limit int) ([]*salesmodel.Transaction, int64, error) {
	return s.transactionRepo.FindByCreatorID(creatorID, page, limit)
}

func (s *transactionServiceImpl) GetTransactionsByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*salesmodel.Transaction, int64, error) {
	return s.transactionRepo.FindByCreatorIDWithFilters(creatorID, page, limit, search, status)
}

func (s *transactionServiceImpl) CreateDirectTransaction(transaction *salesmodel.Transaction) error {
	return s.transactionRepo.CreateTransaction(transaction)
}

func (s *transactionServiceImpl) FindTransactionByPurchaseID(purchaseID uint) (*salesmodel.Transaction, error) {
	return s.transactionRepo.FindByPurchaseID(purchaseID)
}

func (s *transactionServiceImpl) UpdateTransactionToCompleted(purchaseID uint, stripePaymentIntentID string) error {
	transaction, err := s.transactionRepo.FindByPurchaseID(purchaseID)
	if err != nil {
		return fmt.Errorf("erro ao buscar transação: %v", err)
	}

	if transaction == nil {
		return fmt.Errorf("transação não encontrada para purchase_id: %d", purchaseID)
	}

	if transaction.Status != salesmodel.TransactionStatusCompleted {
		transaction.Status = salesmodel.TransactionStatusCompleted
		transaction.StripePaymentIntentID = stripePaymentIntentID
		now := time.Now()
		transaction.ProcessedAt = &now

		return s.transactionRepo.UpdateTransaction(transaction)
	}

	slog.Debug("Transação já completada", "transactionID", transaction.ID)
	return nil
}

func maskStripeID(id string) string {
	if len(id) <= 8 {
		return "****"
	}
	return id[:4] + "****" + id[len(id)-4:]
}
