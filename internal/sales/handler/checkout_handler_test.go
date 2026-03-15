package handler

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/mocks"
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	transaction := &salesmodel.Transaction{Model: gorm.Model{ID: 1}}

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

	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 1}, Title: "Test Ebook", Value: 10.0}

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

	creator := &accountmodel.Creator{Model: gorm.Model{ID: 2}, Name: "Test Creator"}

	// Configurar expectativa
	mockCreatorService.On("FindByID", uint(2)).Return(creator, nil)

	// Testar
	result, err := mockCreatorService.FindByID(uint(2))

	// Verificar
	assert.NoError(t, err)
	assert.Equal(t, creator, result)
	mockCreatorService.AssertExpectations(t)
}

type checkoutRequest struct {
	Name      string `json:"name"`
	CPF       string `json:"cpf"`
	Birthdate string `json:"birthdate"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	EbookID   string `json:"ebookId"`
	CSRFToken string `json:"csrfToken"`
}

// TestCreateOrFindClient_ExistingClientByCPF verifica que cliente existente é identificado pelo CPF
func TestCreateOrFindClient_ExistingClientByCPF(t *testing.T) {
	mockClientRepo := new(mocks.MockClientRepository)
	mockCreatorService := new(mocks.MockCreatorService)

	existing := &salesmodel.Client{Model: gorm.Model{ID: 10}, CPF: "12345678901", Email: "old@email.com"}
	mockClientRepo.On("FindByCPF", "12345678901").Return(existing, nil).Once()

	h := &CheckoutHandler{clientRepo: mockClientRepo, creatorService: mockCreatorService}
	req := checkoutRequest{
		Name: "João", CPF: "12345678901", Email: "new@email.com",
		Phone: "11999990000", Birthdate: "01/01/1990",
	}

	client, err := h.createOrFindClient(req, 1)

	assert.NoError(t, err)
	assert.Equal(t, uint(10), client.ID)
	mockClientRepo.AssertNotCalled(t, "FindByEmail", mock.Anything)
	mockClientRepo.AssertExpectations(t)
}

// TestCreateOrFindClient_NewClientCreatedByCPF verifica que novo cliente é criado quando CPF não existe
func TestCreateOrFindClient_NewClientCreatedByCPF(t *testing.T) {
	mockClientRepo := new(mocks.MockClientRepository)
	mockCreatorService := new(mocks.MockCreatorService)

	creator := &accountmodel.Creator{Model: gorm.Model{ID: 1}}
	mockClientRepo.On("FindByCPF", "12345678901").Return(nil, nil).Once()
	mockCreatorService.On("FindByID", uint(1)).Return(creator, nil).Once()
	mockClientRepo.On("Save", mock.AnythingOfType("*model.Client")).Return(nil).Once()

	h := &CheckoutHandler{clientRepo: mockClientRepo, creatorService: mockCreatorService}
	req := checkoutRequest{
		Name: "Maria", CPF: "12345678901", Email: "maria@email.com",
		Phone: "11988880000", Birthdate: "15/06/1985",
	}

	client, err := h.createOrFindClient(req, 1)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "12345678901", client.CPF)
	mockClientRepo.AssertNotCalled(t, "FindByEmail", mock.Anything)
	mockClientRepo.AssertExpectations(t)
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
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 1}, Title: "Test Ebook", Value: 29.99, Status: true, CreatorID: 2}
	creator := &accountmodel.Creator{Model: gorm.Model{ID: 2}, Name: "Test Creator", StripeConnectAccountID: "acct_test", OnboardingCompleted: true, ChargesEnabled: true}

	// Configurar expectativas
	mockEbookService.On("FindByID", uint(1)).Return(ebook, nil)
	mockCreatorService.On("FindByID", uint(2)).Return(creator, nil)
	mockPurchaseService.On("CreatePurchase", uint(1), []uint{uint(123)}).Return(nil)

	// Simular uma transação
	transaction := &salesmodel.Transaction{
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
