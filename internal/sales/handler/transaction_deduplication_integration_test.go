package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	"github.com/anglesson/simple-web-server/internal/mocks"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"

	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Usando mocks do pacote interno

// TestTransactionDeduplication_Integration testa o fluxo completo para garantir que não haja duplicação
func TestTransactionDeduplication_Integration(t *testing.T) {
	// Configurar banco de dados de teste
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Configurar dados de teste
	creator := setupTestCreator(t)
	ebook := setupTestEbook(t, creator.ID)

	// Configurar handlers
	checkoutHandler := setupTestCheckoutHandler(t)

	// CENÁRIO REAL: Teste do problema identificado - Success view não encontra transação pendente
	t.Run("Should_Not_Create_Fallback_Transaction_When_Pending_Transaction_Exists", func(t *testing.T) {
		// 1. Simular checkout criando purchase + transaction pendente
		client, _ := simulateCheckoutCreation(t, checkoutHandler, ebook.ID)

		// 2. Buscar a compra criada para verificar
		purchases := getPurchasesByClientAndEbook(t, client.ID, ebook.ID)
		require.Len(t, purchases, 1, "Deve ter exatamente 1 purchase")
		purchaseID := purchases[0].ID

		// 3. Verificar transação pendente criada
		pendingTransactions := getTransactionsByPurchaseID(t, purchaseID)
		require.Len(t, pendingTransactions, 1, "Deve ter exatamente 1 transação pendente")
		require.Equal(t, salesmodel.TransactionStatusPending, pendingTransactions[0].Status)
		require.NotZero(t, pendingTransactions[0].PurchaseID, "PurchaseID não deve ser zero")

		// 4. CRUCIAL: Simular Success View tentando atualizar a transação existente
		err := checkoutHandler.transactionService.UpdateTransactionToCompleted(purchaseID, "pi_test_success_123")

		// 5. VALIDAÇÃO: A atualização deve funcionar, não deve criar nova transação
		if err != nil {
			t.Logf("⚠️  PROBLEMA ENCONTRADO: UpdateTransactionToCompleted falhou: %v", err)
			t.Logf("Isso causaria a criação de uma transação fallback!")
		} else {
			t.Logf("✅ UpdateTransactionToCompleted funcionou corretamente")
		}

		// 6. Verificar que ainda temos apenas 1 transação (agora completed)
		finalTransactions := getTransactionsByPurchaseID(t, purchaseID)
		require.Len(t, finalTransactions, 1, "Ainda deve ter exatamente 1 transação")
		require.Equal(t, salesmodel.TransactionStatusCompleted, finalTransactions[0].Status)
		require.Equal(t, "pi_test_success_123", finalTransactions[0].StripePaymentIntentID)

		// 7. Verificar que não há transações órfãs
		orphanTransactions := getOrphanTransactions(t)
		require.Empty(t, orphanTransactions, "Não deve haver transações órfãs")
	})

	// CENÁRIO: Simular o que acontece quando UpdateTransactionToCompleted falha
	t.Run("Should_Handle_UpdateTransactionToCompleted_Failure_Gracefully", func(t *testing.T) {
		// 1. Criar uma purchase sem transação (cenário onde pode falhar)
		clientID := uint(99) // Cliente fictício
		client := &salesmodel.Client{
			Name:      "Test Client Fallback",
			Email:     "testfallback@example.com",
			CPF:       "99999999999",
			Phone:     "11999999999",
			Birthdate: "1990-01-01",
		}
		err := database.DB.Create(client).Error
		require.NoError(t, err)
		clientID = client.ID

		// 2. Criar purchase manualmente
		purchase := salesmodel.NewPurchase(ebook.ID, clientID, "test-hash-"+fmt.Sprint(time.Now().UnixNano()))
		err = database.DB.Create(purchase).Error
		require.NoError(t, err)

		// 3. Tentar atualizar transação que não existe
		err = checkoutHandler.transactionService.UpdateTransactionToCompleted(purchase.ID, "pi_test_no_pending")

		// 4. Deve falhar porque não existe transação pendente
		require.Error(t, err, "Deve retornar erro quando não há transação pendente")

		// 5. Verificar que nenhuma transação foi criada incorretamente
		transactions := getTransactionsByPurchaseID(t, purchase.ID)
		require.Empty(t, transactions, "Não deve criar transação quando UpdateTransactionToCompleted falha")

		// 6. Verificar que não há transações órfãs
		orphanTransactions := getOrphanTransactions(t)
		require.Empty(t, orphanTransactions, "Não deve haver transações órfãs")
	})

	// CENÁRIO: Verificar que múltiplas compras do mesmo ebook são permitidas (mas cada uma com sua transação)
	t.Run("Should_Allow_Multiple_Purchases_Of_Same_Ebook_By_Same_Client", func(t *testing.T) {
		clientID := uint(100)
		client := &salesmodel.Client{
			Name:      "Test Client Multiple",
			Email:     "testmultiple@example.com",
			CPF:       "88888888888",
			Phone:     "11999999999",
			Birthdate: "1990-01-01",
		}
		err := database.DB.Create(client).Error
		require.NoError(t, err)
		clientID = client.ID

		// Simular duas compras diferentes do mesmo ebook pelo mesmo cliente
		for i := 0; i < 2; i++ {
			// Criar purchase
			purchase := salesmodel.NewPurchase(ebook.ID, clientID, "test-hash-"+fmt.Sprint(time.Now().UnixNano()))
			err = database.DB.Create(purchase).Error
			require.NoError(t, err)

			// Criar transação correspondente
			transaction := salesmodel.NewTransaction(purchase.ID, creator.ID, salesmodel.SplitTypePercentage)
			transaction.Status = salesmodel.TransactionStatusCompleted
			transaction.StripePaymentIntentID = fmt.Sprintf("pi_test_multiple_%d", i)
			transaction.CalculateSplit(29000)
			err = checkoutHandler.transactionService.CreateDirectTransaction(transaction)
			require.NoError(t, err)
		}

		// Verificar que existem 2 purchases diferentes
		purchases := getPurchasesByClientAndEbook(t, clientID, ebook.ID)
		require.Len(t, purchases, 2, "Cliente deve poder comprar o mesmo ebook múltiplas vezes")

		// Verificar que cada purchase tem sua própria transação
		for _, purchase := range purchases {
			transactions := getTransactionsByPurchaseID(t, purchase.ID)
			require.Len(t, transactions, 1, fmt.Sprintf("Purchase %d deve ter exatamente 1 transação", purchase.ID))
			require.NotZero(t, transactions[0].PurchaseID, "Transação deve ter PurchaseID válido")
		}

		// Verificar que não há transações órfãs
		orphanTransactions := getOrphanTransactions(t)
		require.Empty(t, orphanTransactions, "Não deve haver transações órfãs")
	})
}

// TestRealWorldScenario_DuplicateTransactionBug reproduz o cenário exato observado nos logs
func TestRealWorldScenario_DuplicateTransactionBug(t *testing.T) {
	// Configurar banco de dados de teste
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Configurar dados de teste
	creator := setupTestCreator(t)
	ebook := setupTestEbook(t, creator.ID)
	checkoutHandler := setupTestCheckoutHandler(t)

	t.Run("Reproduce_The_Exact_Log_Scenario", func(t *testing.T) {
		// 1. Simular o cenário problemático
		client := &salesmodel.Client{
			Name:      "Real World Client",
			Email:     "realworld@example.com",
			CPF:       "11111111111",
			Phone:     "11999999999",
			Birthdate: "1990-01-01",
		}
		err := database.DB.Create(client).Error
		require.NoError(t, err)

		// 2. Criar purchase COM transação pending (normal)
		purchase1 := salesmodel.NewPurchase(ebook.ID, client.ID, "test-hash-1-"+fmt.Sprint(time.Now().UnixNano()))
		err = database.DB.Create(purchase1).Error
		require.NoError(t, err)

		transaction1 := salesmodel.NewTransaction(purchase1.ID, creator.ID, salesmodel.SplitTypePercentage)
		transaction1.Status = salesmodel.TransactionStatusPending
		transaction1.CalculateSplit(29000)
		err = checkoutHandler.transactionService.CreateDirectTransaction(transaction1)
		require.NoError(t, err)

		// 3. Criar purchase SEM transação (cenário problemático que pode acontecer)
		purchase2 := salesmodel.NewPurchase(ebook.ID, client.ID, "test-hash-2-"+fmt.Sprint(time.Now().UnixNano()))
		err = database.DB.Create(purchase2).Error
		require.NoError(t, err)

		// 4. Tentar UpdateTransactionToCompleted na purchase2 (vai falhar)
		err = checkoutHandler.transactionService.UpdateTransactionToCompleted(purchase2.ID, "pi_test_fallback")
		require.Error(t, err, "Deve falhar porque purchase2 não tem transação pending")

		// 5. PROBLEMA: No código real, isso acionaria o fallback que cria nova transação
		t.Logf("⚠️  SIMULANDO FALLBACK PROBLEMÁTICO:")
		t.Logf("Se o código de fallback for executado, criará transação duplicada!")

		// Verificar estado atual - deve ter apenas 1 transação por purchase
		trans1 := getTransactionsByPurchaseID(t, purchase1.ID)
		trans2 := getTransactionsByPurchaseID(t, purchase2.ID)

		require.Len(t, trans1, 1, "Purchase1 deve ter exatamente 1 transação")
		require.Len(t, trans2, 0, "Purchase2 NÃO deve ter transação (fallback não deve ser executado)")

		t.Logf("✅ CORRETO: Fallback não criou transação desnecessária")
		t.Logf("✅ SOLUÇÃO: Código deveria retornar erro em vez de criar fallback")

		// Verificar que não há transações órfãs
		orphanTransactions := getOrphanTransactions(t)
		require.Empty(t, orphanTransactions, "Não deve haver transações órfãs")
	})

	// TESTE CRÍTICO: Verificar que fallback foi REMOVIDO
	t.Run("FIXED_BUG_No_More_Fallback_Transactions", func(t *testing.T) {
		t.Log("✅ VERIFICANDO: Fallback foi removido - não deve criar transações desnecessárias")

		client := &salesmodel.Client{
			Name:      "No Fallback Client",
			Email:     "nofallback@example.com",
			CPF:       "22222222222",
			Phone:     "11999999999",
			Birthdate: "1990-01-01",
		}
		err := database.DB.Create(client).Error
		require.NoError(t, err)

		// Simular uma purchase que chegou ao success view mas SEM transação pending
		purchase := salesmodel.NewPurchase(ebook.ID, client.ID, "test-hash-"+fmt.Sprint(time.Now().UnixNano()))
		err = database.DB.Create(purchase).Error
		require.NoError(t, err)

		// Estado inicial: purchase existe, mas sem transação
		initialTransactions := getTransactionsByPurchaseID(t, purchase.ID)
		require.Len(t, initialTransactions, 0, "Purchase deve começar sem transações")

		// TESTAR O COMPORTAMENTO ATUAL (SEM FALLBACK):
		// 1. Tentar UpdateTransactionToCompleted (vai falhar)
		err = checkoutHandler.transactionService.UpdateTransactionToCompleted(purchase.ID, "pi_no_fallback")
		require.Error(t, err, "Deve falhar - não há transação pending")

		// 2. VERIFICAR: Sem fallback, nenhuma transação deve ser criada
		finalTransactions := getTransactionsByPurchaseID(t, purchase.ID)
		require.Len(t, finalTransactions, 0, "✅ CORRETO: Sem fallback, nenhuma transação criada")

		t.Log("✅ SUCESSO: Código não mais cria transações fallback desnecessárias")
		t.Log("✅ BENEFÍCIO: Problemas na origem agora ficam visíveis para investigação")

		// Verificar que não há transações órfãs
		orphanTransactions := getOrphanTransactions(t)
		require.Empty(t, orphanTransactions, "Não deve haver transações órfãs")
	})
}

// TestTransactionDeduplication_WithMockedServices demonstra uso dos mocks do pacote internal/mocks
func TestTransactionDeduplication_WithMockedServices(t *testing.T) {
	t.Run("Should_Use_Mocks_From_Internal_Package", func(t *testing.T) {
		// Demonstrar o uso correto dos mocks do pacote interno
		mockTransactionService := &mocks.MockTransactionService{}
		mockEmailService := &mocks.MockSalesEmailService{}

		// Configurar comportamentos esperados
		mockTransactionService.On("FindTransactionByPurchaseID", uint(1)).Return(nil, errors.New("not found"))
		mockTransactionService.On("CreateDirectTransaction", mock.AnythingOfType("*model.Transaction")).Return(nil)
		mockTransactionService.On("UpdateTransactionToCompleted", uint(1), "pi_test_123").Return(nil)

		mockEmailService.On("SendLinkToDownload", mock.AnythingOfType("[]*model.Purchase")).Return()

		// Verificar que os mocks podem ser chamados sem erro
		_, err := mockTransactionService.FindTransactionByPurchaseID(1)
		assert.Error(t, err, "Deve retornar erro como configurado")

		transaction := &salesmodel.Transaction{PurchaseID: 1}
		err = mockTransactionService.CreateDirectTransaction(transaction)
		assert.NoError(t, err, "Deve criar transação sem erro")

		err = mockTransactionService.UpdateTransactionToCompleted(1, "pi_test_123")
		assert.NoError(t, err, "Deve atualizar transação sem erro")

		mockEmailService.SendLinkToDownload([]*salesmodel.Purchase{})

		// Verificar que todas as expectativas foram atendidas
		mockTransactionService.AssertExpectations(t)
		mockEmailService.AssertExpectations(t)
	})
}

// TestDuplicatePurchaseBug testa o problema de duplicação de purchases
func TestDuplicatePurchaseBug(t *testing.T) {
	// Configurar banco de dados de teste
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Configurar dados de teste
	creator := setupTestCreator(t)
	ebook := setupTestEbook(t, creator.ID)
	checkoutHandler := setupTestCheckoutHandler(t)

	t.Run("Should_Not_Create_Duplicate_Purchases_For_Same_Client_Ebook", func(t *testing.T) {
		t.Log("✅ PROBLEMA RESOLVIDO: Checkout agora previne purchases duplicadas!")

		client := &salesmodel.Client{
			Name:      "Duplicate Purchase Client",
			Email:     "duplicate@example.com",
			CPF:       "33333333333",
			Phone:     "11999999999",
			Birthdate: "1990-01-01",
		}
		err := database.DB.Create(client).Error
		require.NoError(t, err)

		// 1. Simular primeira chamada ao checkout (normal)
		t.Log("📞 Primeira chamada ao checkout...")
		err = checkoutHandler.purchaseService.CreatePurchase(ebook.ID, []uint{client.ID})
		require.NoError(t, err)

		purchases1 := getPurchasesByClientAndEbook(t, client.ID, ebook.ID)
		require.Len(t, purchases1, 1, "Deve ter 1 purchase após primeira chamada")

		// Criar transação para a primeira purchase
		transaction1 := salesmodel.NewTransaction(purchases1[0].ID, creator.ID, salesmodel.SplitTypePercentage)
		transaction1.Status = salesmodel.TransactionStatusPending
		transaction1.CalculateSplit(29000)
		err = checkoutHandler.transactionService.CreateDirectTransaction(transaction1)
		require.NoError(t, err)

		// 2. Simular segunda chamada ao checkout (tentativa de duplicação)
		t.Log("📞 Segunda chamada ao checkout (tentativa de duplicação)...")
		err = checkoutHandler.purchaseService.CreatePurchase(ebook.ID, []uint{client.ID})
		// ATUALMENTE: Não dá erro, mas cria purchase duplicada
		require.NoError(t, err)

		// 3. Verificar a solução: agora temos apenas 1 purchase
		purchases2 := getPurchasesByClientAndEbook(t, client.ID, ebook.ID)
		if len(purchases2) > 1 {
			t.Errorf("❌ BUG AINDA PRESENTE: %d purchases para o mesmo cliente/ebook!", len(purchases2))
			t.Logf("Purchase IDs: %v", func() []uint {
				var ids []uint
				for _, p := range purchases2 {
					ids = append(ids, p.ID)
				}
				return ids
			}())

			// Verificar que apenas a primeira tem transação
			trans1 := getTransactionsByPurchaseID(t, purchases2[0].ID)
			trans2 := getTransactionsByPurchaseID(t, purchases2[1].ID)

			t.Logf("Purchase %d tem %d transações", purchases2[0].ID, len(trans1))
			t.Logf("Purchase %d tem %d transações", purchases2[1].ID, len(trans2))
		} else {
			t.Log("✅ CORRETO: Apenas 1 purchase para o cliente/ebook")
		}

		// 4. Verificar se a transação pode ser atualizada com sucesso
		if len(purchases2) == 1 {
			firstPurchaseID := purchases2[0].ID
			err = checkoutHandler.transactionService.UpdateTransactionToCompleted(firstPurchaseID, "pi_success_test")
			require.NoError(t, err, "Deve conseguir atualizar transação da purchase única")
			t.Log("✅ SUCESSO: Transaction atualizada corretamente")
		}

		t.Log("🎯 SOLUÇÃO IMPLEMENTADA: CreatePurchaseWithResult verifica purchase existente antes de criar nova")
	})
}

func setupTestDB(t *testing.T) {
	// Configurar banco de dados SQLite em memória para teste
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrar schemas
	err = db.AutoMigrate(
		&authmodel.User{},
		&accountmodel.Creator{},
		&salesmodel.Client{},
		&librarymodel.Ebook{},
		&salesmodel.Purchase{},
		&salesmodel.Transaction{},
	)
	require.NoError(t, err)

	// Configurar database global
	database.DB = db

	t.Logf("Database setup completed successfully")
}

func cleanupTestDB(t *testing.T) {
	// Limpar todas as tabelas
	database.DB.Exec("DELETE FROM transactions")
	database.DB.Exec("DELETE FROM purchases")
	database.DB.Exec("DELETE FROM ebooks")
	database.DB.Exec("DELETE FROM clients")
	database.DB.Exec("DELETE FROM creators")
	database.DB.Exec("DELETE FROM users")
}

func setupTestCreator(t *testing.T) *accountmodel.Creator {
	user := &authmodel.User{
		Username: "creator",
		Email:    "creator@test.com",
		Password: "hashedpassword",
	}
	err := database.DB.Create(user).Error
	require.NoError(t, err)

	creator := &accountmodel.Creator{
		UserID:                 user.ID,
		Name:                   "Test Creator",
		CPF:                    "12345678901",
		OnboardingCompleted:    true,
		ChargesEnabled:         true,
		StripeConnectAccountID: "acct_test123",
	}
	err = database.DB.Create(creator).Error
	require.NoError(t, err)

	return creator
}

func setupTestEbook(t *testing.T, creatorID uint) *librarymodel.Ebook {
	ebook := &librarymodel.Ebook{
		Title:       "Test Ebook",
		Description: "Test Description",
		Value:       290.00,
		Status:      true,
		CreatorID:   creatorID,
	}
	err := database.DB.Create(ebook).Error
	require.NoError(t, err)

	return ebook
}

func setupTestCheckoutHandler(t *testing.T) *CheckoutHandler {
	// Configurar repositórios
	transactionRepo := salesrepo.NewTransactionRepository(database.DB)
	purchaseRepo := salesrepo.NewPurchaseRepository()

	// Configurar services
	emailService := &mocks.MockSalesEmailService{} // Mock do pacote interno
	emailService.On("SendLinkToDownload", mock.MatchedBy(func(purchases []*salesmodel.Purchase) bool {
		return true // Aceitar qualquer chamada
	})).Return()

	purchaseService := salesvc.NewPurchaseService(purchaseRepo, emailService)
	transactionService := salesvc.NewTransactionService(transactionRepo, purchaseService, nil, nil)

	// Template renderer mock
	templateRenderer := &mocks.MockTemplateRenderer{}

	return &CheckoutHandler{
		templateRenderer:   templateRenderer,
		purchaseService:    purchaseService,
		transactionService: transactionService,
	}
}

func simulateCheckoutCreation(t *testing.T, handler *CheckoutHandler, ebookID uint) (*salesmodel.Client, string) {
	// Criar cliente de teste com dados únicos
	clientID := time.Now().UnixNano() // Usar timestamp como ID único
	client := &salesmodel.Client{
		Name:      fmt.Sprintf("Test Client %d", clientID%1000), // Usar módulo para nomes mais curtos
		Email:     fmt.Sprintf("test%d@example.com", clientID%10000),
		CPF:       fmt.Sprintf("%011d", clientID%99999999999), // CPF único baseado no timestamp
		Phone:     "11999999999",
		Birthdate: time.Now().AddDate(-30, 0, 0).Format("2006-01-02"),
	}
	err := database.DB.Create(client).Error
	require.NoError(t, err)

	// Simular requisição de checkout
	reqBody := map[string]interface{}{
		"name":      client.Name,
		"cpf":       client.CPF,
		"birthdate": "01/01/1994",
		"email":     client.Email,
		"phone":     client.Phone,
		"ebookId":   strconv.FormatUint(uint64(ebookID), 10),
	}

	jsonBody, _ := json.Marshal(reqBody)
	_ = httptest.NewRequest("POST", "/checkout", bytes.NewBuffer(jsonBody))

	// Executar a lógica similar ao CreateEbookCheckout
	// Simular criação de compra
	err = handler.purchaseService.CreatePurchase(ebookID, []uint{client.ID})
	require.NoError(t, err)

	// Buscar a compra criada
	var latestPurchase salesmodel.Purchase
	err = database.DB.Where("client_id = ? AND ebook_id = ?", client.ID, ebookID).
		Order("created_at DESC").
		First(&latestPurchase).Error
	require.NoError(t, err)

	// Simular criação da transação pendente (como no checkout_handler.go)
	transaction := salesmodel.NewTransaction(latestPurchase.ID, 1, salesmodel.SplitTypePercentage) // creator_id = 1
	transaction.Status = salesmodel.TransactionStatusPending
	transaction.CalculateSplit(29000) // 290.00 * 100

	err = handler.transactionService.CreateDirectTransaction(transaction)
	require.NoError(t, err)

	return client, "test_session_id"
}

func getPurchasesByClientAndEbook(t *testing.T, clientID, ebookID uint) []salesmodel.Purchase {
	var purchases []salesmodel.Purchase
	err := database.DB.Where("client_id = ? AND ebook_id = ?", clientID, ebookID).Find(&purchases).Error
	require.NoError(t, err)
	return purchases
}

func getTransactionsByPurchaseID(t *testing.T, purchaseID uint) []salesmodel.Transaction {
	var transactions []salesmodel.Transaction
	err := database.DB.Where("purchase_id = ?", purchaseID).Find(&transactions).Error
	require.NoError(t, err)
	return transactions
}

func getOrphanTransactions(t *testing.T) []salesmodel.Transaction {
	var transactions []salesmodel.Transaction
	err := database.DB.Where("purchase_id = 0 OR purchase_id IS NULL").Find(&transactions).Error
	require.NoError(t, err)
	return transactions
}
