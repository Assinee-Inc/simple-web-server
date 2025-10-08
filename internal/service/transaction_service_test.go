package service_test

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockTransactionRepository é um mock do repositório de transações
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) CreateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByID(id uint) (*models.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByCreatorID(creatorID uint, page, limit int) ([]*models.Transaction, int64, error) {
	args := m.Called(creatorID, page, limit)
	return args.Get(0).([]*models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) FindByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*models.Transaction, int64, error) {
	args := m.Called(creatorID, page, limit, search, status)
	return args.Get(0).([]*models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) FindByPurchaseID(purchaseID uint) (*models.Transaction, error) {
	args := m.Called(purchaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) UpdateTransactionStatus(id uint, status models.TransactionStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

// MockCreatorService é um mock do serviço de criadores
type MockCreatorService struct {
	mock.Mock
}

func (m *MockCreatorService) CreateCreator(input models.InputCreateCreator) (*models.Creator, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorService) FindCreatorByEmail(email string) (*models.Creator, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorService) FindCreatorByUserID(userID uint) (*models.Creator, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorService) FindByID(id uint) (*models.Creator, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Creator), args.Error(1)
}

func (m *MockCreatorService) UpdateCreator(creator *models.Creator) error {
	args := m.Called(creator)
	return args.Error(0)
}

// MockPurchaseService é um mock do serviço de compras
type MockPurchaseService struct {
	mock.Mock
}

func (m *MockPurchaseService) CreatePurchase(ebookId uint, clients []uint) error {
	args := m.Called(ebookId, clients)
	return args.Error(0)
}

func (m *MockPurchaseService) GetPurchaseByID(id uint) (*models.Purchase, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Purchase), args.Error(1)
}

func (m *MockPurchaseService) GetEbookFile(purchaseID int, fileID uint) (string, error) {
	args := m.Called(purchaseID, fileID)
	return args.String(0), args.Error(1)
}

func (m *MockPurchaseService) GetEbookFiles(purchaseID int) ([]*models.File, error) {
	args := m.Called(purchaseID)
	return args.Get(0).([]*models.File), args.Error(1)
}

// MockStripeService é um mock do serviço Stripe
type MockStripeService struct {
	mock.Mock
}

func TestCreateTransaction(t *testing.T) {
	// Configurar mocks
	mockTransactionRepo := new(MockTransactionRepository)
	mockCreatorService := new(MockCreatorService)

	// Para o PurchaseService, precisamos criar um stub com o repositório mockado
	purchaseRepo := &repository.PurchaseRepository{}
	emailService := &service.EmailService{}
	purchaseService := service.NewPurchaseService(purchaseRepo, emailService)

	// Mock do StripeService
	mockStripeService := &service.StripeService{}

	// Serviço a ser testado
	transactionService := service.NewTransactionService(
		mockTransactionRepo,
		purchaseService,
		mockCreatorService,
		mockStripeService,
	)

	// Configurar objetos de teste
	creator := &models.Creator{
		Model:                  gorm.Model{ID: 1},
		Name:                   "Test Creator",
		StripeConnectAccountID: "acct_123456",
		OnboardingCompleted:    true,
		ChargesEnabled:         true,
	}

	ebook := &models.Ebook{
		Model:     gorm.Model{ID: 1},
		Title:     "Test Ebook",
		Value:     100.00,
		CreatorID: 1,
		Creator:   *creator,
	}

	purchase := &models.Purchase{
		Model:    gorm.Model{ID: 1},
		EbookID:  1,
		ClientID: 1,
		Ebook:    *ebook,
	}

	// Configurar expectativas dos mocks
	mockCreatorService.On("FindByID", uint(1)).Return(creator, nil)
	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*models.Transaction")).Return(nil)

	// Executar teste
	transaction, err := transactionService.CreateTransaction(purchase, 10000)

	// Verificar resultados
	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, uint(1), transaction.PurchaseID)
	assert.Equal(t, uint(1), transaction.CreatorID)
	assert.Equal(t, models.SplitTypePercentage, transaction.SplitType)
	assert.Equal(t, models.TransactionStatusPending, transaction.Status)

	// Verificar chamadas de mock
	mockCreatorService.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

func TestCreateTransactionFailureCreatorNotFound(t *testing.T) {
	// Configurar mocks
	mockTransactionRepo := new(MockTransactionRepository)
	mockCreatorService := new(MockCreatorService)

	// Para o PurchaseService, precisamos criar um stub com o repositório mockado
	purchaseRepo := &repository.PurchaseRepository{}
	emailService := &service.EmailService{}
	purchaseService := service.NewPurchaseService(purchaseRepo, emailService)

	// Mock do StripeService
	mockStripeService := &service.StripeService{}

	// Serviço a ser testado
	transactionService := service.NewTransactionService(
		mockTransactionRepo,
		purchaseService,
		mockCreatorService,
		mockStripeService,
	)

	// Configurar objetos de teste
	ebook := &models.Ebook{
		Model:     gorm.Model{ID: 1},
		Title:     "Test Ebook",
		Value:     100.00,
		CreatorID: 1,
	}

	purchase := &models.Purchase{
		Model:    gorm.Model{ID: 1},
		EbookID:  1,
		ClientID: 1,
		Ebook:    *ebook,
	}

	// Configurar expectativas dos mocks - criador não encontrado
	mockCreatorService.On("FindByID", uint(1)).Return(nil, assert.AnError)

	// Executar teste
	transaction, err := transactionService.CreateTransaction(purchase, 10000)

	// Verificar resultados
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Contains(t, err.Error(), "erro ao buscar criador")

	// Verificar chamadas de mock
	mockCreatorService.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

func TestCreateTransactionFailureCreatorNotConnected(t *testing.T) {
	// Configurar mocks
	mockTransactionRepo := new(MockTransactionRepository)
	mockCreatorService := new(MockCreatorService)

	// Para o PurchaseService, precisamos criar um stub com o repositório mockado
	purchaseRepo := &repository.PurchaseRepository{}
	emailService := &service.EmailService{}
	purchaseService := service.NewPurchaseService(purchaseRepo, emailService)

	// Mock do StripeService
	mockStripeService := &service.StripeService{}

	// Serviço a ser testado
	transactionService := service.NewTransactionService(
		mockTransactionRepo,
		purchaseService,
		mockCreatorService,
		mockStripeService,
	)

	// Configurar objetos de teste - criador sem conta Stripe Connect
	creator := &models.Creator{
		Model:                  gorm.Model{ID: 1},
		Name:                   "Test Creator",
		StripeConnectAccountID: "acct_123456",
		OnboardingCompleted:    false, // onboarding não completo
		ChargesEnabled:         false,
	}

	ebook := &models.Ebook{
		Model:     gorm.Model{ID: 1},
		Title:     "Test Ebook",
		Value:     100.00,
		CreatorID: 1,
		Creator:   *creator,
	}

	purchase := &models.Purchase{
		Model:    gorm.Model{ID: 1},
		EbookID:  1,
		ClientID: 1,
		Ebook:    *ebook,
	}

	// Configurar expectativas dos mocks
	mockCreatorService.On("FindByID", uint(1)).Return(creator, nil)

	// Executar teste
	transaction, err := transactionService.CreateTransaction(purchase, 10000)

	// Verificar resultados
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Contains(t, err.Error(), "criador não está habilitado para receber pagamentos")

	// Verificar chamadas de mock
	mockCreatorService.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}
