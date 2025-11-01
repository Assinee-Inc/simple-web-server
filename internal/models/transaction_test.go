package models_test

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestTransactionCalculateSplit(t *testing.T) {
	tests := []struct {
		name               string
		splitType          models.SplitType
		platformPercentage float64
		platformFixedFee   int64
		totalAmount        int64
		expectedPlatform   int64
		expectedCreator    int64
	}{
		{
			name:               "Docffy 2.91% + 1.00 on R$30",
			splitType:          models.SplitTypePercentage,
			platformPercentage: 0.0291, // 2,91%
			platformFixedFee:   100,    // R$1,00
			totalAmount:        3000,   // R$30,00
			expectedPlatform:   187,    // 2,91% de 30,00 (R$0,87) + R$1,00 = R$1,87
			expectedCreator:    2654,   // 30,00 - Stripe(1,59) - Docffy(1,87) = 26,54
		},
		{
			name:               "Fixed Fee of R$5 on R$100",
			splitType:          models.SplitTypeFixedAmount,
			platformPercentage: 0,
			platformFixedFee:   500,   // R$5,00
			totalAmount:        10000, // R$100,00
			expectedPlatform:   500,   // R$5,00
			expectedCreator:    9062,  // 100,00 - Stripe(4,38) - 5,00 = 90,62
		},
		{
			name:               "5% Percentage Split on R$1000",
			splitType:          models.SplitTypePercentage,
			platformPercentage: 0.05, // 5%
			platformFixedFee:   0,
			totalAmount:        100000, // R$1000,00
			expectedPlatform:   5000,   // 5% de 1000,00
			expectedCreator:    90971,  // 1000,00 - Stripe(40,29) - 50,00 = 909,71
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction := &models.Transaction{
				SplitType:          tt.splitType,
				PlatformPercentage: tt.platformPercentage,
				PlatformFixedFee:   tt.platformFixedFee,
			}

			transaction.CalculateSplit(tt.totalAmount)

			// Verificar valor da plataforma
			assert.Equal(t, tt.expectedPlatform, transaction.PlatformAmount)

			// Verificar valor do criador
			assert.Equal(t, tt.expectedCreator, transaction.CreatorAmount)

			// Verificar se a soma está correta (total - taxa stripe = plataforma + criador)
			expectedTotal := transaction.PlatformAmount + transaction.CreatorAmount + transaction.StripeProcessingFee
			assert.Equal(t, tt.totalAmount, expectedTotal)
		})
	}
}
