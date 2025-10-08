package config

// BusinessConfig contém todas as configurações e constantes de regras de negócio
type BusinessConfig struct {
	// Split de pagamentos
	PlatformFeePercentage float64 // Percentual da plataforma sobre vendas

	// Stripe
	StripeProcessingPercentage float64 // Taxa percentual do Stripe (3.5%)
	StripeProcessingFixedFee   int64   // Taxa fixa do Stripe em centavos (R$ 0,39)

	// Compras
	PurchaseExpirationDays int // Dias de acesso após compra

	// Valores formatados para uso em strings
	PlatformFeePercentageDisplay string
}

// Instância global das configurações de negócio
var Business = BusinessConfig{
	// 5% para a plataforma - centralizado aqui!
	PlatformFeePercentage: 0.05,

	// Configurações do Stripe (baseadas na documentação oficial)
	StripeProcessingPercentage: 0.035,
	StripeProcessingFixedFee:   39, // R$ 0,39 em centavos

	// 30 dias de acesso após compra
	PurchaseExpirationDays: 30,

	// Display formatado
	PlatformFeePercentageDisplay: "5",
}

// GetPlatformFeeAmount calcula o valor da taxa da plataforma em centavos
func (bc *BusinessConfig) GetPlatformFeeAmount(totalAmount int64) int64 {
	return int64(float64(totalAmount) * bc.PlatformFeePercentage)
}

// GetCreatorAmount calcula o valor que vai para o criador em centavos
func (bc *BusinessConfig) GetCreatorAmount(totalAmount int64) int64 {
	return totalAmount - bc.GetPlatformFeeAmount(totalAmount)
}

// GetStripeProcessingFee calcula a taxa de processamento do Stripe
func (bc *BusinessConfig) GetStripeProcessingFee(totalAmount int64) int64 {
	return int64(float64(totalAmount)*bc.StripeProcessingPercentage) + bc.StripeProcessingFixedFee
}

// GetPlatformFeeFromNetAmount calcula a taxa da plataforma após descontar a taxa do Stripe
func (bc *BusinessConfig) GetPlatformFeeFromNetAmount(totalAmount int64) int64 {
	stripesFee := bc.GetStripeProcessingFee(totalAmount)
	netAmount := totalAmount - stripesFee
	return int64(float64(netAmount) * bc.PlatformFeePercentage)
}

// GetCreatorAmountFromNetAmount calcula o valor do criador após descontar Stripe e plataforma
func (bc *BusinessConfig) GetCreatorAmountFromNetAmount(totalAmount int64) int64 {
	stripesFee := bc.GetStripeProcessingFee(totalAmount)
	platformFee := bc.GetPlatformFeeFromNetAmount(totalAmount)
	return totalAmount - stripesFee - platformFee
}
