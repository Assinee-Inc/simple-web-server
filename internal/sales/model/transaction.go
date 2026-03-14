package model

import (
	"fmt"
	"math"
	"time"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"gorm.io/gorm"
)

// SplitType representa o tipo de divisão aplicada
type SplitType string

const (
	SplitTypePercentage  SplitType = "percentage"
	SplitTypeFixedAmount SplitType = "fixed"
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

	PublicID              string `json:"public_id" gorm:"type:varchar(40);uniqueIndex"`
	StripePaymentIntentID string `json:"stripe_payment_intent_id"`
	StripeTransferID      string `json:"stripe_transfer_id"`

	TotalAmount         int64 `json:"total_amount"`
	PlatformAmount      int64 `json:"platform_amount"`
	CreatorAmount       int64 `json:"creator_amount"`
	StripeProcessingFee int64 `json:"stripe_processing_fee"`

	SplitType          SplitType `json:"split_type"`
	PlatformPercentage float64   `json:"platform_percentage"`
	PlatformFixedFee   int64     `json:"platform_fixed_fee"`

	PurchaseID uint                `json:"purchase_id"`
	Purchase   Purchase            `gorm:"foreignKey:PurchaseID"`
	CreatorID  uint                `json:"creator_id"`
	Creator    accountmodel.Creator `gorm:"foreignKey:CreatorID"`

	Status       TransactionStatus `json:"status"`
	ProcessedAt  *time.Time        `json:"processed_at"`
	ErrorMessage string            `json:"error_message"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.PublicID == "" {
		t.PublicID = utils.GeneratePublicID("txn_")
	}
	return nil
}

// CalculateSplit calcula os valores de split com base na configuração
func (t *Transaction) CalculateSplit(totalAmount int64) {
	t.TotalAmount = totalAmount

	t.StripeProcessingFee = config.Business.GetStripeProcessingFee(totalAmount)

	percentPart := int64(math.Round(float64(totalAmount) * t.PlatformPercentage))
	fixedPart := t.PlatformFixedFee
	t.PlatformAmount = percentPart + fixedPart

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

	t.PlatformPercentage = config.Business.PlatformFeePercentage
	t.PlatformFixedFee = config.Business.PlatformFixedFeeCents

	return t
}

func (t *Transaction) GetFormattedTotalAmount() string {
	return formatCentsToBRL(t.TotalAmount)
}

func (t *Transaction) GetFormattedPlatformAmount() string {
	return formatCentsToBRL(t.PlatformAmount + t.StripeProcessingFee)
}

func (t *Transaction) GetFormattedCreatorAmount() string {
	return formatCentsToBRL(t.CreatorAmount)
}

func (t *Transaction) GetFormattedProcessingFee() string {
	return formatCentsToBRL(t.StripeProcessingFee)
}

func formatCentsToBRL(cents int64) string {
	reais := float64(cents) / 100.0
	return fmt.Sprintf("R$ %.2f", reais)
}
