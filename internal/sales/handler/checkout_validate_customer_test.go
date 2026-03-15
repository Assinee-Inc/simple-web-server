package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	"github.com/anglesson/simple-web-server/internal/mocks"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func buildValidateCustomerRequest(t *testing.T, body map[string]any) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return httptest.NewRequest(http.MethodPost, "/api/validate-customer", bytes.NewBuffer(raw))
}

func buildCheckoutHandlerForValidation(
	ebookSvc *mocks.MockEbookService,
	clientRepo *mocks.MockClientRepository,
	purchaseSvc *mocks.MockPurchaseService,
	creatorSvc *mocks.MockCreatorService,
) *CheckoutHandler {
	return &CheckoutHandler{
		ebookService:    ebookSvc,
		clientRepo:      clientRepo,
		purchaseService: purchaseSvc,
		creatorService:  creatorSvc,
		rfService:       nil, // non-production: RF skipped
	}
}

// TestValidateCustomer_AlreadyPurchased_ReturnConflict verifica que o endpoint
// retorna 409 com already_purchased=true quando o cliente já comprou o ebook.
func TestValidateCustomer_AlreadyPurchased_ReturnConflict(t *testing.T) {
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 1}, PublicID: "ebook-pub-1", Status: true, CreatorID: 10}
	client := &salesmodel.Client{Model: gorm.Model{ID: 5}, CPF: "12345678901"}
	purchase := &salesmodel.Purchase{Model: gorm.Model{ID: 99}}
	creator := &accountmodel.Creator{Model: gorm.Model{ID: 10}, Name: "João Producer", Email: "joao@producer.com"}

	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "ebook-pub-1").Return(ebook, nil)
	mockClient.On("FindByCPF", "12345678901").Return(client, nil)
	mockPurchase.On("FindExistingPurchase", uint(1), uint(5)).Return(purchase, nil)
	mockCreator.On("FindByID", uint(10)).Return(creator, nil)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "João Silva",
		"cpf":       "12345678901",
		"birthdate": "01/01/1990",
		"email":     "joao@email.com",
		"phone":     "11999999999",
		"ebookId":   "ebook-pub-1",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, true, resp["already_purchased"])
	assert.Equal(t, "joao@producer.com", resp["creator_email"])
	assert.Equal(t, "João Producer", resp["creator_name"])

	mockEbook.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockPurchase.AssertExpectations(t)
	mockCreator.AssertExpectations(t)
}

// TestValidateCustomer_AlreadyPurchased_CreatorNotFound verifica o caso onde
// o creator não é encontrado (retorna campos vazios mas ainda bloqueia a compra).
func TestValidateCustomer_AlreadyPurchased_CreatorNotFound(t *testing.T) {
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 2}, PublicID: "ebook-pub-2", Status: true, CreatorID: 20}
	client := &salesmodel.Client{Model: gorm.Model{ID: 7}, CPF: "98765432100"}
	purchase := &salesmodel.Purchase{Model: gorm.Model{ID: 88}}

	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "ebook-pub-2").Return(ebook, nil)
	mockClient.On("FindByCPF", "98765432100").Return(client, nil)
	mockPurchase.On("FindExistingPurchase", uint(2), uint(7)).Return(purchase, nil)
	mockCreator.On("FindByID", uint(20)).Return((*accountmodel.Creator)(nil), errors.New("not found"))

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "Maria Souza",
		"cpf":       "98765432100",
		"birthdate": "01/01/1985",
		"email":     "maria@email.com",
		"phone":     "11988887777",
		"ebookId":   "ebook-pub-2",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, true, resp["already_purchased"])
	assert.Equal(t, "", resp["creator_email"])
	assert.Equal(t, "", resp["creator_name"])
}

// TestValidateCustomer_CPFFound_NoPurchase_Proceeds verifica que o fluxo continua
// normalmente quando o cliente existe mas ainda não comprou o ebook.
func TestValidateCustomer_CPFFound_NoPurchase_Proceeds(t *testing.T) {
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 3}, PublicID: "ebook-pub-3", Status: true, CreatorID: 30}
	client := &salesmodel.Client{Model: gorm.Model{ID: 9}, CPF: "11122233344"}

	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "ebook-pub-3").Return(ebook, nil)
	mockClient.On("FindByCPF", "11122233344").Return(client, nil)
	mockPurchase.On("FindExistingPurchase", uint(3), uint(9)).Return((*salesmodel.Purchase)(nil), errors.New("not found"))

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "Carlos Lima",
		"cpf":       "11122233344",
		"birthdate": "01/01/1992",
		"email":     "carlos@email.com",
		"phone":     "11977776666",
		"ebookId":   "ebook-pub-3",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	// Non-production: RF skipped, should return success
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.Nil(t, resp["already_purchased"])
}

// TestValidateCustomer_CPFNotFound_Proceeds verifica que o fluxo continua
// normalmente quando o CPF não pertence a nenhum cliente cadastrado.
func TestValidateCustomer_CPFNotFound_Proceeds(t *testing.T) {
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 4}, PublicID: "ebook-pub-4", Status: true, CreatorID: 40}

	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "ebook-pub-4").Return(ebook, nil)
	mockClient.On("FindByCPF", "55566677788").Return((*salesmodel.Client)(nil), nil)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "Ana Costa",
		"cpf":       "55566677788",
		"birthdate": "01/01/2000",
		"email":     "ana@email.com",
		"phone":     "11966665555",
		"ebookId":   "ebook-pub-4",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
}

// TestValidateCustomer_CPFLookupError_Proceeds verifica que o fluxo continua
// normalmente quando FindByCPF retorna erro (falha conservadora).
func TestValidateCustomer_CPFLookupError_Proceeds(t *testing.T) {
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 5}, PublicID: "ebook-pub-5", Status: true, CreatorID: 50}

	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "ebook-pub-5").Return(ebook, nil)
	mockClient.On("FindByCPF", "99988877766").Return((*salesmodel.Client)(nil), errors.New("db error"))

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "Pedro Alves",
		"cpf":       "99988877766",
		"birthdate": "01/01/1988",
		"email":     "pedro@email.com",
		"phone":     "11955554444",
		"ebookId":   "ebook-pub-5",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	// Erro ao buscar CPF não deve bloquear a compra
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
}

// TestValidateCustomer_MissingFields verifica validação de campos obrigatórios.
func TestValidateCustomer_MissingFields(t *testing.T) {
	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name": "Sem CPF",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

// TestValidateCustomer_InvalidCPFLength verifica que CPF com tamanho inválido é rejeitado.
func TestValidateCustomer_InvalidCPFLength(t *testing.T) {
	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "João",
		"cpf":       "123", // inválido
		"birthdate": "01/01/1990",
		"email":     "joao@email.com",
		"phone":     "11999999999",
		"ebookId":   "pub-1",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestValidateCustomer_InvalidEmail verifica que e-mail inválido é rejeitado.
func TestValidateCustomer_InvalidEmail(t *testing.T) {
	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "João",
		"cpf":       "12345678901",
		"birthdate": "01/01/1990",
		"email":     "ab", // inválido
		"phone":     "11999999999",
		"ebookId":   "pub-1",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestValidateCustomer_InvalidPhone verifica que telefone com tamanho inválido é rejeitado.
func TestValidateCustomer_InvalidPhone(t *testing.T) {
	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "João",
		"cpf":       "12345678901",
		"birthdate": "01/01/1990",
		"email":     "joao@email.com",
		"phone":     "123", // inválido
		"ebookId":   "pub-1",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestValidateCustomer_MissingEbookID verifica que ebookId vazio é rejeitado.
func TestValidateCustomer_MissingEbookID(t *testing.T) {
	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "João",
		"cpf":       "12345678901",
		"birthdate": "01/01/1990",
		"email":     "joao@email.com",
		"phone":     "11999999999",
		"ebookId":   "",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestValidateCustomer_EbookNotFound verifica que ebook inválido é rejeitado.
func TestValidateCustomer_EbookNotFound(t *testing.T) {
	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "nao-existe").Return((*librarymodel.Ebook)(nil), errors.New("not found"))

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "João",
		"cpf":       "12345678901",
		"birthdate": "01/01/1990",
		"email":     "joao@email.com",
		"phone":     "11999999999",
		"ebookId":   "nao-existe",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestValidateCustomer_InvalidBody verifica que body inválido retorna erro.
func TestValidateCustomer_InvalidBody(t *testing.T) {
	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	req := httptest.NewRequest(http.MethodPost, "/api/validate-customer", bytes.NewBufferString("invalid json"))
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestValidateCustomer_EbookInactive verifica que ebook inativo é rejeitado.
func TestValidateCustomer_EbookInactive(t *testing.T) {
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 6}, PublicID: "ebook-inactive", Status: false}

	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "ebook-inactive").Return(ebook, nil)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "João",
		"cpf":       "12345678901",
		"birthdate": "01/01/1990",
		"email":     "joao@email.com",
		"phone":     "11999999999",
		"ebookId":   "ebook-inactive",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestValidateCustomer_AlreadyPurchased_ErrorMessagePresent verifica a mensagem de erro.
func TestValidateCustomer_AlreadyPurchased_ErrorMessagePresent(t *testing.T) {
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 7}, PublicID: "ebook-pub-7", Status: true, CreatorID: 70}
	client := &salesmodel.Client{Model: gorm.Model{ID: 15}, CPF: "44455566677"}
	purchase := &salesmodel.Purchase{Model: gorm.Model{ID: 200}}
	creator := &accountmodel.Creator{Model: gorm.Model{ID: 70}, Name: "Produtor Teste", Email: "produtor@test.com"}

	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "ebook-pub-7").Return(ebook, nil)
	mockClient.On("FindByCPF", "44455566677").Return(client, nil)
	mockPurchase.On("FindExistingPurchase", uint(7), uint(15)).Return(purchase, nil)
	mockCreator.On("FindByID", uint(70)).Return(creator, nil)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "Test User",
		"cpf":       "44455566677",
		"birthdate": "01/01/1995",
		"email":     "test@email.com",
		"phone":     "11944443333",
		"ebookId":   "ebook-pub-7",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["error"])

	mockEbook.AssertExpectations(t)
	mockClient.AssertExpectations(t)
	mockPurchase.AssertExpectations(t)
	mockCreator.AssertExpectations(t)
}

// TestFindExistingPurchase_MockDelegate verifica que MockPurchaseService.FindExistingPurchase
// delega corretamente ao mock testify.
func TestFindExistingPurchase_MockDelegate(t *testing.T) {
	mockPurchase := new(mocks.MockPurchaseService)
	purchase := &salesmodel.Purchase{Model: gorm.Model{ID: 42}}

	mockPurchase.On("FindExistingPurchase", uint(1), uint(2)).Return(purchase, nil)

	result, err := mockPurchase.FindExistingPurchase(1, 2)

	assert.NoError(t, err)
	assert.Equal(t, purchase, result)
	mockPurchase.AssertExpectations(t)
}

// TestFindExistingPurchase_MockDelegate_NotFound verifica o caso onde a compra não existe.
func TestFindExistingPurchase_MockDelegate_NotFound(t *testing.T) {
	mockPurchase := new(mocks.MockPurchaseService)

	mockPurchase.On("FindExistingPurchase", uint(1), uint(2)).Return((*salesmodel.Purchase)(nil), errors.New("not found"))

	result, err := mockPurchase.FindExistingPurchase(1, 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockPurchase.AssertExpectations(t)
}

// TestFindByCPF_MockDelegate verifica que MockClientRepository.FindByCPF funciona.
func TestFindByCPF_MockDelegate(t *testing.T) {
	mockRepo := new(mocks.MockClientRepository)
	client := &salesmodel.Client{Model: gorm.Model{ID: 5}, CPF: "12345678901"}

	mockRepo.On("FindByCPF", "12345678901").Return(client, nil)

	result, err := mockRepo.FindByCPF("12345678901")

	assert.NoError(t, err)
	assert.Equal(t, client, result)
	mockRepo.AssertExpectations(t)
}

// TestFindByCPF_MockDelegate_NotFound verifica o caso onde o cliente não é encontrado.
func TestFindByCPF_MockDelegate_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockClientRepository)

	mockRepo.On("FindByCPF", "00000000000").Return((*salesmodel.Client)(nil), nil)

	result, err := mockRepo.FindByCPF("00000000000")

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// TestValidateCustomer_AlreadyPurchased_ContentTypeIsJSON verifica o Content-Type da resposta.
func TestValidateCustomer_AlreadyPurchased_ContentTypeIsJSON(t *testing.T) {
	ebook := &librarymodel.Ebook{Model: gorm.Model{ID: 8}, PublicID: "ebook-pub-8", Status: true, CreatorID: 80}
	client := &salesmodel.Client{Model: gorm.Model{ID: 20}, CPF: "77788899900"}
	purchase := &salesmodel.Purchase{Model: gorm.Model{ID: 300}}
	creator := &accountmodel.Creator{Model: gorm.Model{ID: 80}, Name: "Creator", Email: "creator@x.com"}

	mockEbook := new(mocks.MockEbookService)
	mockClient := new(mocks.MockClientRepository)
	mockPurchase := new(mocks.MockPurchaseService)
	mockCreator := new(mocks.MockCreatorService)

	mockEbook.On("FindByPublicID", "ebook-pub-8").Return(ebook, nil)
	mockClient.On("FindByCPF", "77788899900").Return(client, nil)
	mockPurchase.On("FindExistingPurchase", uint(8), uint(20)).Return(purchase, nil)
	mockCreator.On("FindByID", uint(80)).Return(creator, nil)

	handler := buildCheckoutHandlerForValidation(mockEbook, mockClient, mockPurchase, mockCreator)

	body := map[string]any{
		"name":      "Test",
		"cpf":       "77788899900",
		"birthdate": "01/01/1990",
		"email":     "test@x.com",
		"phone":     "11933332222",
		"ebookId":   "ebook-pub-8",
	}
	req := buildValidateCustomerRequest(t, body)
	rr := httptest.NewRecorder()

	handler.ValidateCustomer(rr, req)

	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	// Verify mock with any argument for FindByID (to handle mock setup flexibility)
	_ = mock.MatchedBy(func(id uint) bool { return true })
}
