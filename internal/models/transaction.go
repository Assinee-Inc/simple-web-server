package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// SplitType representa o tipo de divisão aplicada
type SplitType string

const (
	SplitTypePercentage  SplitType = "percentage" // Divisão por porcentagem
	SplitTypeFixedAmount SplitType = "fixed"      // Valor fixo para plataforma
)

// TransactionStatus representa o status da transação
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

// Transaction representa uma transação financeira com split de pagamento
type Transaction struct {
	gorm.Model

	// Identificadores externos
	StripePaymentIntentID string `json:"stripe_payment_intent_id"`
	StripeTransferID      string `json:"stripe_transfer_id"`

	// Valores da transação
	TotalAmount         int64 `json:"total_amount"`          // Em centavos
	PlatformAmount      int64 `json:"platform_amount"`       // Valor para a plataforma
	CreatorAmount       int64 `json:"creator_amount"`        // Valor para o criador
	StripeProcessingFee int64 `json:"stripe_processing_fee"` // Taxa do Stripe

	// Configuração do split
	SplitType          SplitType `json:"split_type"`
	PlatformPercentage float64   `json:"platform_percentage"` // Se for percentual
	PlatformFixedFee   int64     `json:"platform_fixed_fee"`  // Se for taxa fixa

	// Relacionamentos
	PurchaseID uint     `json:"purchase_id"`
	Purchase   Purchase `gorm:"foreignKey:PurchaseID"`
	CreatorID  uint     `json:"creator_id"`
	Creator    Creator  `gorm:"foreignKey:CreatorID"`

	// Status e processamento
	Status       TransactionStatus `json:"status"`
	ProcessedAt  *time.Time        `json:"processed_at"`
	ErrorMessage string            `json:"error_message"`
}

// CalculateSplit calcula os valores de split com base no tipo configurado
func (t *Transaction) CalculateSplit(totalAmount int64) {
	t.TotalAmount = totalAmount

	// Calcular a taxa do Stripe (aproximadamente 3.5% + R$0.39)
	stripePercentage := 0.035
	stripeFixed := int64(39) // R$0.39 em centavos
	t.StripeProcessingFee = int64(float64(totalAmount)*stripePercentage) + stripeFixed

	remainingAmount := totalAmount - t.StripeProcessingFee

	if t.SplitType == SplitTypePercentage {
		t.PlatformAmount = int64(float64(remainingAmount) * t.PlatformPercentage)
	} else if t.SplitType == SplitTypeFixedAmount {
		t.PlatformAmount = t.PlatformFixedFee
	}

	// Valor restante vai para o criador
	t.CreatorAmount = remainingAmount - t.PlatformAmount
}

// NewTransaction cria uma nova transação com split
func NewTransaction(purchaseID, creatorID uint, splitType SplitType) *Transaction {
	t := &Transaction{
		PurchaseID: purchaseID,
		CreatorID:  creatorID,
		SplitType:  splitType,
		Status:     TransactionStatusPending,
	}

	// Valores padrão
	if splitType == SplitTypePercentage {
		t.PlatformPercentage = 0.15 // 15% por padrão
	} else {
		t.PlatformFixedFee = 500 // R$5,00 em centavos
	}

	return t
}

// GetFormattedTotalAmount retorna o valor total formatado
func (t *Transaction) GetFormattedTotalAmount() string {
	return formatCentsToBRL(t.TotalAmount)
}

// GetFormattedPlatformAmount retorna o valor da plataforma formatado
func (t *Transaction) GetFormattedPlatformAmount() string {
	return formatCentsToBRL(t.PlatformAmount)
}

// GetFormattedCreatorAmount retorna o valor do criador formatado
func (t *Transaction) GetFormattedCreatorAmount() string {
	return formatCentsToBRL(t.CreatorAmount)
}

// GetFormattedProcessingFee retorna a taxa de processamento formatada
func (t *Transaction) GetFormattedProcessingFee() string {
	return formatCentsToBRL(t.StripeProcessingFee)
}

// formatCentsToBRL formata um valor em centavos para BRL
func formatCentsToBRL(cents int64) string {
	reais := float64(cents) / 100.0
	return fmt.Sprintf("R$ %.2f", reais)
}
