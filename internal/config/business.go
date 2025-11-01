package config

import "math"

// BusinessConfig contém todas as configurações e constantes de regras de negócio
type BusinessConfig struct {
	// Split de pagamentos
	PlatformFeePercentage float64 // Percentual da taxa da Docffy sobre vendas (ex.: 2,91% -> 0.0291)
	PlatformFixedFeeCents int64   // Parcela fixa da taxa da Docffy em centavos (ex.: R$ 1,00 -> 100)

	// Stripe
	StripeProcessingPercentage float64 // Taxa percentual do Stripe (ex.: 3,99% -> 0.0399)
	StripeProcessingFixedFee   int64   // Taxa fixa do Stripe em centavos (R$ 0,39)

	// Compras
	PurchaseExpirationDays int // Dias de acesso após compra

	// Valores formatados para uso em strings
	PlatformFeePercentageDisplay string // Exibição amigável da taxa Docffy (ex.: "2,91% + R$ 1,00")
}

// Instância global das configurações de negócio
var Business = BusinessConfig{
	// Taxa da Docffy: 2,91% + R$ 1,00
	PlatformFeePercentage: 0.0291,
	PlatformFixedFeeCents: 100, // R$ 1,00

	// Configurações do Stripe (baseadas na documentação oficial)
	StripeProcessingPercentage: 0.0399, // 3,99%
	StripeProcessingFixedFee:   39,     // R$ 0,39 em centavos

	// 30 dias de acesso após compra
	PurchaseExpirationDays: 30,

	// Display formatado
	PlatformFeePercentageDisplay: "2,91% + R$ 1,00",
}

// GetPlatformFeeAmount calcula o valor da taxa da plataforma (Docffy) em centavos
// Fórmula: (percentual sobre o valor total) + parcela fixa, com arredondamento para o centavo mais próximo
func (bc *BusinessConfig) GetPlatformFeeAmount(totalAmount int64) int64 {
	percent := int64(math.Round(float64(totalAmount) * bc.PlatformFeePercentage))
	return percent + bc.PlatformFixedFeeCents
}

// GetCreatorAmount calcula o valor que vai para o criador em centavos (sem considerar taxa do Stripe)
func (bc *BusinessConfig) GetCreatorAmount(totalAmount int64) int64 {
	return totalAmount - bc.GetPlatformFeeAmount(totalAmount)
}

// GetStripeProcessingFee calcula a taxa de processamento do Stripe
// Fórmula: (percentual sobre o valor total) + parcela fixa, com arredondamento para o centavo mais próximo
func (bc *BusinessConfig) GetStripeProcessingFee(totalAmount int64) int64 {
	percent := int64(math.Round(float64(totalAmount) * bc.StripeProcessingPercentage))
	return percent + bc.StripeProcessingFixedFee
}

// GetPlatformFeeFromNetAmount calcula a taxa da plataforma após descontar a taxa do Stripe
// Aplica o percentual sobre o valor líquido (total - taxa Stripe) e soma a parcela fixa da Docffy, com arredondamento
func (bc *BusinessConfig) GetPlatformFeeFromNetAmount(totalAmount int64) int64 {
	stripesFee := bc.GetStripeProcessingFee(totalAmount)
	netAmount := totalAmount - stripesFee
	percent := int64(math.Round(float64(netAmount) * bc.PlatformFeePercentage))
	return percent + bc.PlatformFixedFeeCents
}

// GetCreatorAmountFromNetAmount calcula o valor do criador após descontar Stripe e plataforma
func (bc *BusinessConfig) GetCreatorAmountFromNetAmount(totalAmount int64) int64 {
	stripesFee := bc.GetStripeProcessingFee(totalAmount)
	platformFee := bc.GetPlatformFeeFromNetAmount(totalAmount)
	return totalAmount - stripesFee - platformFee
}
