package models

import (
	"fmt"
	"math"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
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

// CalculateSplit calcula os valores de split com base na configuração
// Regra atual: Docffy = (percentual sobre o valor total) + taxa fixa
func (t *Transaction) CalculateSplit(totalAmount int64) {
	t.TotalAmount = totalAmount

	// Taxa do Stripe baseada na configuração centralizada
	t.StripeProcessingFee = config.Business.GetStripeProcessingFee(totalAmount)

	// Calcular taxa da plataforma: percentual (sobre o total bruto) + parcela fixa, com arredondamento
	percentPart := int64(math.Round(float64(totalAmount) * t.PlatformPercentage))
	fixedPart := t.PlatformFixedFee
	t.PlatformAmount = percentPart + fixedPart

	// Valor restante vai para o criador
	remainingAmount := totalAmount - t.StripeProcessingFee
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

	// Valores padrão da taxa da plataforma (Docffy)
	// Sempre inicializamos ambos para permitir cálculo percentual + fixa
	t.PlatformPercentage = config.Business.PlatformFeePercentage
	t.PlatformFixedFee = config.Business.PlatformFixedFeeCents

	return t
}

// GetFormattedTotalAmount retorna o valor total formatado
func (t *Transaction) GetFormattedTotalAmount() string {
	return formatCentsToBRL(t.TotalAmount)
}

// GetFormattedPlatformAmount retorna o valor da plataforma formatado
func (t *Transaction) GetFormattedPlatformAmount() string {
	return formatCentsToBRL(t.PlatformAmount + t.StripeProcessingFee)
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
