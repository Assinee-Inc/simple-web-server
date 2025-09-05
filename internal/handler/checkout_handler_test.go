package handler

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestPurchaseServiceMock verifica se o mock do PurchaseService funciona
func TestPurchaseServiceMock(t *testing.T) {
	mockPurchaseService := new(mocks.MockPurchaseService)

	// Configurar expectativa
	mockPurchaseService.On("CreatePurchase", uint(1), []uint{uint(99)}).Return(nil)

	// Testar
	err := mockPurchaseService.CreatePurchase(uint(1), []uint{uint(99)})

	// Verificar
	assert.NoError(t, err)
	mockPurchaseService.AssertExpectations(t)
}

// TestTransactionServiceMock verifica se o mock do TransactionService funciona
func TestTransactionServiceMock(t *testing.T) {
	mockTransactionService := new(mocks.MockTransactionService)

	transaction := &models.Transaction{Model: gorm.Model{ID: 1}}

	// Configurar expectativa
	mockTransactionService.On("CreateDirectTransaction", transaction).Return(nil)

	// Testar
	err := mockTransactionService.CreateDirectTransaction(transaction)

	// Verificar
	assert.NoError(t, err)
	mockTransactionService.AssertExpectations(t)
}

// TestEbookServiceMock verifica se o mock do EbookService funciona
func TestEbookServiceMock(t *testing.T) {
	mockEbookService := new(mocks.MockEbookService)

	ebook := &models.Ebook{Model: gorm.Model{ID: 1}, Title: "Test Ebook", Value: 10.0}

	// Configurar expectativa
	mockEbookService.On("FindByID", uint(1)).Return(ebook, nil)

	// Testar
	result, err := mockEbookService.FindByID(uint(1))

	// Verificar
	assert.NoError(t, err)
	assert.Equal(t, ebook, result)
	mockEbookService.AssertExpectations(t)
}

// TestCreatorServiceMock verifica se o mock do CreatorService funciona
func TestCreatorServiceMock(t *testing.T) {
	mockCreatorService := new(mocks.MockCreatorService)

	creator := &models.Creator{Model: gorm.Model{ID: 2}, Name: "Test Creator"}

	// Configurar expectativa
	mockCreatorService.On("FindByID", uint(2)).Return(creator, nil)

	// Testar
	result, err := mockCreatorService.FindByID(uint(2))

	// Verificar
	assert.NoError(t, err)
	assert.Equal(t, creator, result)
	mockCreatorService.AssertExpectations(t)
}

// TestCheckoutHandlerMocksIntegration testa a integração entre os mocks
func TestCheckoutHandlerMocksIntegration(t *testing.T) {
	// Configurar mocks
	mockPurchaseService := new(mocks.MockPurchaseService)
	mockTransactionService := new(mocks.MockTransactionService)
	mockEbookService := new(mocks.MockEbookService)
	mockCreatorService := new(mocks.MockCreatorService)

	// Dados de teste
	ebook := &models.Ebook{Model: gorm.Model{ID: 1}, Title: "Test Ebook", Value: 29.99, Status: true, CreatorID: 2}
	creator := &models.Creator{Model: gorm.Model{ID: 2}, Name: "Test Creator", StripeConnectAccountID: "acct_test", OnboardingCompleted: true, ChargesEnabled: true}

	// Configurar expectativas
	mockEbookService.On("FindByID", uint(1)).Return(ebook, nil)
	mockCreatorService.On("FindByID", uint(2)).Return(creator, nil)
	mockPurchaseService.On("CreatePurchase", uint(1), []uint{uint(123)}).Return(nil)

	// Simular uma transação
	transaction := &models.Transaction{
		Model:          gorm.Model{ID: 1},
		Status:         "pending",
		CreatorID:      2,
		TotalAmount:    2999, // 29.99 em centavos
		PlatformAmount: 299,  // 10% de comissão
		CreatorAmount:  2700, // 90% para o criador
	}
	mockTransactionService.On("CreateDirectTransaction", transaction).Return(nil)

	// Executar operações simulando o fluxo do CheckoutHandler
	foundEbook, err := mockEbookService.FindByID(uint(1))
	assert.NoError(t, err)
	assert.Equal(t, ebook.Title, foundEbook.Title)

	foundCreator, err := mockCreatorService.FindByID(uint(2))
	assert.NoError(t, err)
	assert.Equal(t, creator.Name, foundCreator.Name)

	err = mockPurchaseService.CreatePurchase(uint(1), []uint{uint(123)})
	assert.NoError(t, err)

	err = mockTransactionService.CreateDirectTransaction(transaction)
	assert.NoError(t, err)

	// Verificar que todos os mocks foram chamados conforme esperado
	mockEbookService.AssertExpectations(t)
	mockCreatorService.AssertExpectations(t)
	mockPurchaseService.AssertExpectations(t)
	mockTransactionService.AssertExpectations(t)
}
