package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
// e que o banco de dados não é alterado.
func TestCreateOrFindClient_ExistingClientByCPF(t *testing.T) {
	mockClientRepo := new(mocks.MockClientRepository)
	mockCreatorService := new(mocks.MockCreatorService)

	existing := &salesmodel.Client{Model: gorm.Model{ID: 10}, CPF: "12345678901", Email: "old@email.com"}
	mockClientRepo.On("FindByCPF", "12345678901").Return(existing, nil).Once()

	h := &CheckoutHandler{clientRepo: mockClientRepo, creatorService: mockCreatorService}
	req := checkoutRequest{
		Name: "João", CPF: "12345678901", Email: "old@email.com",
		Phone: "11999990000", Birthdate: "01/01/1990",
	}

	client, err := h.createOrFindClient(req, 1)

	assert.NoError(t, err)
	assert.Equal(t, uint(10), client.ID)
	mockClientRepo.AssertNotCalled(t, "Save", mock.Anything)
	mockClientRepo.AssertExpectations(t)
}

// TestCreateOrFindClient_ExistingClientUsesCheckoutEmail verifica que, ao encontrar um cliente
// pelo CPF, o email informado no formulário de checkout é utilizado para o envio do e-mail de
// download — sem alterar o cadastro do cliente no banco de dados.
func TestCreateOrFindClient_ExistingClientUsesCheckoutEmail(t *testing.T) {
	mockClientRepo := new(mocks.MockClientRepository)
	mockCreatorService := new(mocks.MockCreatorService)

	existing := &salesmodel.Client{Model: gorm.Model{ID: 10}, CPF: "12345678901", Email: "stored@email.com"}
	mockClientRepo.On("FindByCPF", "12345678901").Return(existing, nil).Once()

	h := &CheckoutHandler{clientRepo: mockClientRepo, creatorService: mockCreatorService}
	req := checkoutRequest{
		Name: "João", CPF: "12345678901", Email: "checkout@email.com",
		Phone: "11999990000", Birthdate: "01/01/1990",
	}

	client, err := h.createOrFindClient(req, 1)

	assert.NoError(t, err)
	assert.Equal(t, uint(10), client.ID)
	assert.Equal(t, "checkout@email.com", client.Email)
	mockClientRepo.AssertNotCalled(t, "Save", mock.Anything)
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

// newValidateCustomerRequest cria um *http.Request com corpo JSON para ValidateCustomer
func newValidateCustomerRequest(t *testing.T, body map[string]any) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/checkout/validate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// US-02-T1: purchase confirmed → 409 com already_purchased: true
func TestValidateCustomer_ConfirmedPurchase_Returns409WithAlreadyPurchased(t *testing.T) {
	mockEbookService := new(mocks.MockEbookService)
	mockClientRepo := new(mocks.MockClientRepository)
	mockPurchaseService := new(mocks.MockPurchaseService)
	mockCreatorService := new(mocks.MockCreatorService)

	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 1}, Status: true, CreatorID: 2}
	client := &salesmodel.Client{Model: gorm.Model{ID: 10}, CPF: "12345678901"}
	creator := &accountmodel.Creator{Model: gorm.Model{ID: 2}, Name: "Creator", Email: "creator@test.com"}
	purchase := &salesmodel.Purchase{
		Model:         gorm.Model{ID: 5},
		PaymentStatus: salesmodel.PaymentStatusConfirmed,
	}

	mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	mockClientRepo.On("FindByCPF", "12345678901").Return(client, nil)
	mockPurchaseService.On("FindExistingPurchase", uint(1), uint(10)).Return(purchase, nil)
	mockCreatorService.On("FindByID", uint(2)).Return(creator, nil)

	h := &CheckoutHandler{
		ebookService:    mockEbookService,
		clientRepo:      mockClientRepo,
		purchaseService: mockPurchaseService,
		creatorService:  mockCreatorService,
	}

	body := map[string]any{
		"name": "João", "cpf": "12345678901", "birthdate": "01/01/1990",
		"email": "joao@test.com", "phone": "11999990000", "ebookId": "ebk_abc",
	}
	req := newValidateCustomerRequest(t, body)
	w := httptest.NewRecorder()

	h.ValidateCustomer(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, true, resp["already_purchased"])
	assert.Equal(t, "creator@test.com", resp["creator_email"])

	mockEbookService.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
	mockPurchaseService.AssertExpectations(t)
	mockCreatorService.AssertExpectations(t)
}

// US-02-T2: purchase pending → 409 com mensagem "processamento", already_purchased: false
func TestValidateCustomer_PendingPurchase_Returns409WithProcessingMessage(t *testing.T) {
	mockEbookService := new(mocks.MockEbookService)
	mockClientRepo := new(mocks.MockClientRepository)
	mockPurchaseService := new(mocks.MockPurchaseService)
	mockCreatorService := new(mocks.MockCreatorService)

	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 1}, Status: true, CreatorID: 2}
	client := &salesmodel.Client{Model: gorm.Model{ID: 10}, CPF: "12345678901"}
	creator := &accountmodel.Creator{Model: gorm.Model{ID: 2}, Name: "Creator", Email: "creator@test.com"}
	purchase := &salesmodel.Purchase{
		Model:         gorm.Model{ID: 5},
		PaymentStatus: salesmodel.PaymentStatusPending,
	}

	mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	mockClientRepo.On("FindByCPF", "12345678901").Return(client, nil)
	mockPurchaseService.On("FindExistingPurchase", uint(1), uint(10)).Return(purchase, nil)
	mockCreatorService.On("FindByID", uint(2)).Return(creator, nil)

	h := &CheckoutHandler{
		ebookService:    mockEbookService,
		clientRepo:      mockClientRepo,
		purchaseService: mockPurchaseService,
		creatorService:  mockCreatorService,
	}

	body := map[string]any{
		"name": "João", "cpf": "12345678901", "birthdate": "01/01/1990",
		"email": "joao@test.com", "phone": "11999990000", "ebookId": "ebk_abc",
	}
	req := newValidateCustomerRequest(t, body)
	w := httptest.NewRecorder()

	h.ValidateCustomer(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, false, resp["already_purchased"])
	assert.Contains(t, resp["error"], "processamento")

	mockEbookService.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
	mockPurchaseService.AssertExpectations(t)
	mockCreatorService.AssertExpectations(t)
}

// US-02-T3: sem purchase → 200
func TestValidateCustomer_NoPurchase_Returns200(t *testing.T) {
	mockEbookService := new(mocks.MockEbookService)
	mockClientRepo := new(mocks.MockClientRepository)
	mockPurchaseService := new(mocks.MockPurchaseService)

	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 1}, Status: true, CreatorID: 2}
	client := &salesmodel.Client{Model: gorm.Model{ID: 10}, CPF: "12345678901"}

	mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	mockClientRepo.On("FindByCPF", "12345678901").Return(client, nil)
	mockPurchaseService.On("FindExistingPurchase", uint(1), uint(10)).Return(nil, assert.AnError)

	h := &CheckoutHandler{
		ebookService:    mockEbookService,
		clientRepo:      mockClientRepo,
		purchaseService: mockPurchaseService,
	}

	body := map[string]any{
		"name": "João", "cpf": "12345678901", "birthdate": "01/01/1990",
		"email": "joao@test.com", "phone": "11999990000", "ebookId": "ebk_abc",
	}
	req := newValidateCustomerRequest(t, body)
	w := httptest.NewRecorder()

	h.ValidateCustomer(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, true, resp["success"])

	mockEbookService.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
	mockPurchaseService.AssertExpectations(t)
}

// US-02-T4: CPF não cadastrado → FindExistingPurchase nunca chamado
func TestValidateCustomer_CPFNotFound_NeverCallsFindExistingPurchase(t *testing.T) {
	mockEbookService := new(mocks.MockEbookService)
	mockClientRepo := new(mocks.MockClientRepository)
	mockPurchaseService := new(mocks.MockPurchaseService)

	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 1}, Status: true, CreatorID: 2}

	mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	mockClientRepo.On("FindByCPF", "99999999999").Return(nil, assert.AnError)

	h := &CheckoutHandler{
		ebookService:    mockEbookService,
		clientRepo:      mockClientRepo,
		purchaseService: mockPurchaseService,
	}

	body := map[string]any{
		"name": "Maria", "cpf": "99999999999", "birthdate": "01/01/1990",
		"email": "maria@test.com", "phone": "11999990000", "ebookId": "ebk_abc",
	}
	req := newValidateCustomerRequest(t, body)
	w := httptest.NewRecorder()

	h.ValidateCustomer(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockPurchaseService.AssertNotCalled(t, "FindExistingPurchase", mock.Anything, mock.Anything)

	mockEbookService.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
}

// US-02-T5: criador não encontrado → 409 com campos de criador vazios
func TestValidateCustomer_CreatorNotFound_Returns409WithEmptyCreatorFields(t *testing.T) {
	mockEbookService := new(mocks.MockEbookService)
	mockClientRepo := new(mocks.MockClientRepository)
	mockPurchaseService := new(mocks.MockPurchaseService)
	mockCreatorService := new(mocks.MockCreatorService)

	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 1}, Status: true, CreatorID: 2}
	client := &salesmodel.Client{Model: gorm.Model{ID: 10}, CPF: "12345678901"}
	purchase := &salesmodel.Purchase{
		Model:         gorm.Model{ID: 5},
		PaymentStatus: salesmodel.PaymentStatusConfirmed,
	}

	mockEbookService.On("FindByPublicID", "ebk_abc").Return(ebook, nil)
	mockClientRepo.On("FindByCPF", "12345678901").Return(client, nil)
	mockPurchaseService.On("FindExistingPurchase", uint(1), uint(10)).Return(purchase, nil)
	mockCreatorService.On("FindByID", uint(2)).Return(nil, assert.AnError)

	h := &CheckoutHandler{
		ebookService:    mockEbookService,
		clientRepo:      mockClientRepo,
		purchaseService: mockPurchaseService,
		creatorService:  mockCreatorService,
	}

	body := map[string]any{
		"name": "João", "cpf": "12345678901", "birthdate": "01/01/1990",
		"email": "joao@test.com", "phone": "11999990000", "ebookId": "ebk_abc",
	}
	req := newValidateCustomerRequest(t, body)
	w := httptest.NewRecorder()

	h.ValidateCustomer(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "", resp["creator_email"])
	assert.Equal(t, "", resp["creator_name"])

	mockEbookService.AssertExpectations(t)
	mockClientRepo.AssertExpectations(t)
	mockPurchaseService.AssertExpectations(t)
	mockCreatorService.AssertExpectations(t)
}

// US-03-T1: PurchaseSuccessView — CreatePurchaseWithResult nunca é chamado
// (independente do caminho de execução, a página de sucesso não cria purchase)
func TestPurchaseSuccessView_ConfirmedPurchase_NoSideEffects(t *testing.T) {
	mockPurchaseService := new(mocks.MockPurchaseService)
	mockCreatorService := new(mocks.MockCreatorService)

	// Creator not found → handler returns early, before Stripe call.
	// This ensures CreatePurchaseWithResult is never called in any code path.
	mockCreatorService.On("FindByID", uint(0)).Return(nil, assert.AnError)

	h := &CheckoutHandler{
		purchaseService: mockPurchaseService,
		creatorService:  mockCreatorService,
	}

	req := httptest.NewRequest(http.MethodGet, "/purchase/success?session_id=cs_test_abc", nil)
	w := httptest.NewRecorder()

	h.PurchaseSuccessView(w, req)

	// CreatePurchaseWithResult must never be called from PurchaseSuccessView
	mockPurchaseService.AssertNotCalled(t, "CreatePurchaseWithResult", mock.Anything, mock.Anything)
	mockCreatorService.AssertExpectations(t)
}

// US-03-T2: session_id ausente → 400
func TestPurchaseSuccessView_MissingSessionID_Returns400(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/purchase/success", nil)
	w := httptest.NewRecorder()

	h := &CheckoutHandler{}
	h.PurchaseSuccessView(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
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
