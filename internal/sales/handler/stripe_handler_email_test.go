package handler

import (
	"testing"
	"time"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	"github.com/anglesson/simple-web-server/internal/mocks"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v76"
	"gorm.io/gorm"
)

// newTestStripeHandler cria um StripeHandler com os mocks fornecidos.
// purchaseRepository é nil intencionalmente: só é acessado quando CreatePurchaseWithResult
// retorna uma purchase sem Ebook ou Client carregados, o que os testes aqui evitam.
func newTestStripeHandler(
	purchaseService *mocks.MockPurchaseService,
	emailService *mocks.MockSalesEmailService,
	creatorService *mocks.MockCreatorService,
	transactionService *mocks.MockTransactionService,
) *StripeHandler {
	return &StripeHandler{
		purchaseService:    purchaseService,
		emailService:       emailService,
		creatorService:     creatorService,
		transactionService: transactionService,
	}
}

// fullyLoadedPurchase retorna uma purchase já com Ebook e Client populados,
// evitando que handleEbookPayment precise chamar purchaseRepository.FindByID.
func fullyLoadedPurchase(id uint, creatorID uint, clientEmail string) *salesmodel.Purchase {
	return &salesmodel.Purchase{
		Model:    gorm.Model{ID: id},
		EbookID:  1,
		ClientID: 1,
		Ebook:    librarymodel.Ebook{Model: gorm.Model{ID: 1}, Value: 100.0, CreatorID: creatorID},
		Client:   salesmodel.Client{Model: gorm.Model{ID: 1}, Email: clientEmail},
	}
}

// stripeSessionWithoutPaymentIntent cria uma CheckoutSession de teste sem payment intent,
// evitando chamadas reais ao Stripe API.
func stripeSessionWithoutPaymentIntent(ebookID, clientID string) stripe.CheckoutSession {
	return stripe.CheckoutSession{
		PaymentIntent: &stripe.PaymentIntent{},
		Metadata: map[string]string{
			"ebook_id":  ebookID,
			"client_id": clientID,
		},
	}
}

// TestHandleEbookPayment_SendsEmailExactlyOnce garante que o webhook envia o e-mail uma única vez.
func TestHandleEbookPayment_SendsEmailExactlyOnce(t *testing.T) {
	mockPurchaseService := new(mocks.MockPurchaseService)
	mockEmailService := new(mocks.MockSalesEmailService)
	mockCreatorService := new(mocks.MockCreatorService)
	mockTransactionService := new(mocks.MockTransactionService)

	purchase := fullyLoadedPurchase(1, 1, "buyer@email.com")
	creator := &accountmodel.Creator{Model: gorm.Model{ID: 1}} // sem StripeConnect → nenhuma transação

	mockPurchaseService.On("CreatePurchaseWithResult", uint(1), uint(1)).Return(purchase, nil).Once()
	mockCreatorService.On("FindByID", uint(1)).Return(creator, nil).Once()
	mockEmailService.On("SendLinkToDownload", mock.Anything).Return().Once()

	h := newTestStripeHandler(mockPurchaseService, mockEmailService, mockCreatorService, mockTransactionService)

	err := h.handleEbookPayment(stripeSessionWithoutPaymentIntent("1", "1"))
	time.Sleep(50 * time.Millisecond) // aguarda goroutine do email

	assert.NoError(t, err)
	mockEmailService.AssertNumberOfCalls(t, "SendLinkToDownload", 1)
	mockEmailService.AssertExpectations(t)
}

// TestHandleEbookPayment_DoesNotSendEmailWhenClientEmailIsEmpty garante que nenhum e-mail é
// enviado (e um erro é retornado) quando o cliente não tem e-mail cadastrado.
func TestHandleEbookPayment_DoesNotSendEmailWhenClientEmailIsEmpty(t *testing.T) {
	mockPurchaseService := new(mocks.MockPurchaseService)
	mockEmailService := new(mocks.MockSalesEmailService)
	mockCreatorService := new(mocks.MockCreatorService)
	mockTransactionService := new(mocks.MockTransactionService)

	purchase := fullyLoadedPurchase(2, 1, "") // email vazio
	creator := &accountmodel.Creator{Model: gorm.Model{ID: 1}}

	mockPurchaseService.On("CreatePurchaseWithResult", uint(1), uint(2)).Return(purchase, nil).Once()
	mockCreatorService.On("FindByID", uint(1)).Return(creator, nil).Once()

	h := newTestStripeHandler(mockPurchaseService, mockEmailService, mockCreatorService, mockTransactionService)

	err := h.handleEbookPayment(stripeSessionWithoutPaymentIntent("1", "2"))

	assert.Error(t, err)
	mockEmailService.AssertNotCalled(t, "SendLinkToDownload", mock.Anything)
}
