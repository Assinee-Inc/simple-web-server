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
			name:               "15% Percentage Split on R$100",
			splitType:          models.SplitTypePercentage,
			platformPercentage: 0.15, // 15%
			platformFixedFee:   0,
			totalAmount:        10000, // R$100,00
			expectedPlatform:   1441,  // R$14,41 (15% dos R$96,07 após taxa Stripe)
			expectedCreator:    8170,  // R$81,70
		},
		{
			name:               "Fixed Fee of R$5 on R$100",
			splitType:          models.SplitTypeFixedAmount,
			platformPercentage: 0,
			platformFixedFee:   500,   // R$5,00
			totalAmount:        10000, // R$100,00
			expectedPlatform:   500,   // R$5,00
			expectedCreator:    9111,  // R$91,11 (R$96,11 - R$5,00)
		},
		{
			name:               "15% Percentage Split on R$1000",
			splitType:          models.SplitTypePercentage,
			platformPercentage: 0.15, // 15%
			platformFixedFee:   0,
			totalAmount:        100000, // R$1000,00
			expectedPlatform:   14469,  // R$144,69 (15% dos R$964,61 após taxa Stripe)
			expectedCreator:    81992,  // R$819,92
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
