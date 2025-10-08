package handler

import (
	"fmt"
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository/gorm"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormsqlite "gorm.io/driver/sqlite"
	gormdb "gorm.io/gorm"
)

// TestClientCreatorAssociation_Integration testa todo o fluxo de associação cliente-creator
func TestClientCreatorAssociation_Integration(t *testing.T) {
	setupIntegrationTestDB(t)
	defer cleanupIntegrationTestDB(t)

	t.Run("Complete_Flow_Checkout_To_ClientListing", func(t *testing.T) {
		// 1. Configurar dados iniciais
		creator, ebook := setupCreatorAndEbook(t)

		// 2. Simular checkout completo com criação de cliente
		client := simulateCheckoutWithClientCreation(t, creator, ebook)

		// 3. Verificar que o cliente foi criado com associação ao creator
		verifyClientCreatorAssociation(t, client, creator)

		// 4. Simular listagem de clientes para o creator
		clients := simulateClientListing(t, creator)

		// 5. Verificar que o cliente aparece na listagem
		assert.Len(t, clients, 1, "Creator deve ter 1 cliente associado")
		assert.Equal(t, client.ID, clients[0].ID, "Cliente na listagem deve ser o mesmo criado no checkout")

		t.Logf("✅ Fluxo completo funcionando: cliente criado via checkout aparece na listagem")
	})

	t.Run("Multiple_Clients_Different_Ebooks", func(t *testing.T) {
		// 1. Configurar creator com múltiplos ebooks
		creator, ebook1 := setupCreatorAndEbook(t)
		ebook2 := createAdditionalEbook(t, creator.ID)

		// 2. Criar clientes através de diferentes ebooks
		client1 := simulateCheckoutWithClientCreation(t, creator, ebook1)
		client2 := simulateCheckoutWithClientCreation(t, creator, ebook2)

		// 3. Verificar que ambos os clientes estão associados ao creator
		verifyClientCreatorAssociation(t, client1, creator)
		verifyClientCreatorAssociation(t, client2, creator)

		// 4. Verificar listagem completa
		clients := simulateClientListing(t, creator)
		assert.Len(t, clients, 2, "Creator deve ter 2 clientes associados")

		clientIDs := []uint{clients[0].ID, clients[1].ID}
		assert.Contains(t, clientIDs, client1.ID, "Cliente 1 deve estar na listagem")
		assert.Contains(t, clientIDs, client2.ID, "Cliente 2 deve estar na listagem")

		t.Logf("✅ Múltiplos clientes via diferentes ebooks funcionando")
	})

	t.Run("Client_Appears_Via_Purchase_History", func(t *testing.T) {
		// Este teste verifica se clientes que não têm associação direta
		// mas fizeram compras aparecem na listagem

		// 1. Configurar dados
		creator, ebook := setupCreatorAndEbook(t)

		// 2. Criar cliente SEM associação direta (simular dados antigos)
		client := createClientWithoutAssociation(t)

		// 3. Criar purchase para simular compra
		purchase := createPurchaseForClient(t, client.ID, ebook.ID)

		// 4. Verificar que cliente aparece na listagem via histórico de compras
		clients := simulateClientListing(t, creator)
		assert.Len(t, clients, 1, "Cliente deve aparecer via histórico de compras")
		assert.Equal(t, client.ID, clients[0].ID, "Deve ser o cliente que fez a compra")

		t.Logf("✅ Cliente sem associação direta aparece via histórico de compras - Purchase ID: %d", purchase.ID)
	})

	t.Run("Client_Search_Functionality", func(t *testing.T) {
		// Testar busca por nome, email, CPF, telefone

		// 1. Configurar dados
		creator, ebook := setupCreatorAndEbook(t)

		// 2. Criar clientes com dados específicos para busca
		client1 := createClientWithSpecificData(t, creator, ebook, "João Silva", "joao@example.com", "11111111111", "11999999999")
		client2 := createClientWithSpecificData(t, creator, ebook, "Maria Santos", "maria@example.com", "22222222222", "11888888888")

		// 3. Testar busca por nome
		clients := simulateClientListingWithSearch(t, creator, "João")
		assert.Len(t, clients, 1, "Busca por 'João' deve retornar 1 cliente")
		assert.Equal(t, client1.ID, clients[0].ID)

		// 4. Testar busca por email
		clients = simulateClientListingWithSearch(t, creator, "maria@example.com")
		assert.Len(t, clients, 1, "Busca por email deve retornar 1 cliente")
		assert.Equal(t, client2.ID, clients[0].ID)

		// 5. Testar busca por CPF
		clients = simulateClientListingWithSearch(t, creator, "11111111111")
		assert.Len(t, clients, 1, "Busca por CPF deve retornar 1 cliente")
		assert.Equal(t, client1.ID, clients[0].ID)

		// 6. Testar busca que retorna múltiplos resultados
		clients = simulateClientListingWithSearch(t, creator, "example.com")
		assert.Len(t, clients, 2, "Busca por 'example.com' deve retornar 2 clientes")

		t.Logf("✅ Funcionalidade de busca funcionando para nome, email e CPF")
	})

	t.Run("Repository_Save_Method_Handles_Associations", func(t *testing.T) {
		// Testar especificamente o método Save do repositório

		// 1. Configurar creator
		creator, _ := setupCreatorAndEbook(t)

		// 2. Criar cliente usando construtor NewClient (com associação)
		client := models.NewClient("Teste Repository", "33333333333", "1990-01-01", "repo@test.com", "11555555555", creator)

		// 3. Salvar usando repositório
		clientRepo := gorm.NewClientGormRepository()
		err := clientRepo.Save(client)
		require.NoError(t, err)

		// 4. Verificar que cliente foi salvo com ID
		assert.NotZero(t, client.ID, "Cliente deve ter ID após Save")

		// 5. Verificar associação na tabela client_creators
		var count int64
		err = database.DB.Table("client_creators").
			Where("client_id = ? AND creator_id = ?", client.ID, creator.ID).
			Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Deve existir 1 registro na tabela client_creators")

		// 6. Testar atualização do mesmo cliente (não deve duplicar)
		client.Email = "updated@test.com"
		err = clientRepo.Save(client)
		require.NoError(t, err)

		// 7. Verificar que ainda há apenas 1 associação
		err = database.DB.Table("client_creators").
			Where("client_id = ? AND creator_id = ?", client.ID, creator.ID).
			Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Ainda deve haver apenas 1 registro após atualização")

		t.Logf("✅ Método Save do repositório mantém associações corretamente - Client ID: %d", client.ID)
	})

	t.Run("Client_Without_Association_Not_Listed", func(t *testing.T) {
		// Verificar que clientes sem qualquer relação não aparecem

		// 1. Configurar 2 creators diferentes
		creator1, _ := setupCreatorAndEbook(t)
		creator2, _ := setupCreatorAndEbook(t)

		// 2. Criar cliente associado apenas ao creator1 (vou criar um ebook para o creator1)
		ebook1 := createAdditionalEbook(t, creator1.ID)
		client := simulateCheckoutWithClientCreation(t, creator1, ebook1)

		// 3. Verificar que cliente aparece na listagem do creator1
		clients1 := simulateClientListing(t, creator1)
		assert.Len(t, clients1, 1, "Creator1 deve ver 1 cliente")
		assert.Equal(t, client.ID, clients1[0].ID)

		// 4. Verificar que cliente NÃO aparece na listagem do creator2
		clients2 := simulateClientListing(t, creator2)
		assert.Len(t, clients2, 0, "Creator2 não deve ver clientes de outros creators")

		t.Logf("✅ Isolamento entre creators funcionando corretamente")
	})

	t.Run("Checkout_Handler_Creates_Association", func(t *testing.T) {
		// Este teste seria mais complexo pois requer configurar todo o CheckoutHandler
		// Por enquanto, vamos focar nos testes de repositório e listagem
		// TODO: Implementar teste completo do CheckoutHandler quando necessário

		t.Skip("Teste do CheckoutHandler será implementado em próxima iteração")
	})
}

// Helper functions

func setupIntegrationTestDB(t *testing.T) {
	// Configurar banco SQLite em memória para testes de integração
	db, err := gormdb.Open(gormsqlite.Open(":memory:"), &gormdb.Config{})
	require.NoError(t, err)

	// Migrar todos os modelos
	err = db.AutoMigrate(
		&models.User{},
		&models.Creator{},
		&models.Client{},
		&models.Ebook{},
		&models.Purchase{},
		&models.Transaction{},
		&models.ClientCreator{},
	)
	require.NoError(t, err)

	database.DB = db
	t.Logf("Database integration test setup completed")
}

func cleanupIntegrationTestDB(t *testing.T) {
	database.DB.Exec("DELETE FROM client_creators")
	database.DB.Exec("DELETE FROM transactions")
	database.DB.Exec("DELETE FROM purchases")
	database.DB.Exec("DELETE FROM ebooks")
	database.DB.Exec("DELETE FROM clients")
	database.DB.Exec("DELETE FROM creators")
	database.DB.Exec("DELETE FROM users")
}

func setupCreatorAndEbook(t *testing.T) (*models.Creator, *models.Ebook) {
	// Criar usuário
	user := &models.User{
		Username: fmt.Sprintf("testuser_%d", time.Now().UnixNano()),
		Email:    fmt.Sprintf("user_%d@test.com", time.Now().UnixNano()),
		Password: "hashedpassword",
	}
	err := database.DB.Create(user).Error
	require.NoError(t, err)

	// Criar creator
	creator := &models.Creator{
		UserID:                 user.ID,
		Name:                   fmt.Sprintf("Test Creator %d", time.Now().UnixNano()),
		CPF:                    fmt.Sprintf("%011d", time.Now().UnixNano()%99999999999),
		Email:                  user.Email,
		OnboardingCompleted:    true,
		ChargesEnabled:         true,
		StripeConnectAccountID: fmt.Sprintf("acct_test_%d", time.Now().UnixNano()),
	}
	err = database.DB.Create(creator).Error
	require.NoError(t, err)

	// Criar ebook
	ebook := &models.Ebook{
		Title:       fmt.Sprintf("Test Ebook %d", time.Now().UnixNano()),
		Description: "Test Description",
		Value:       290.00,
		Status:      true,
		CreatorID:   creator.ID,
		Slug:        fmt.Sprintf("test-ebook-%d", time.Now().UnixNano()),
	}
	err = database.DB.Create(ebook).Error
	require.NoError(t, err)

	return creator, ebook
}

func createAdditionalEbook(t *testing.T, creatorID uint) *models.Ebook {
	ebook := &models.Ebook{
		Title:       fmt.Sprintf("Additional Ebook %d", time.Now().UnixNano()),
		Description: "Additional Description",
		Value:       190.00,
		Status:      true,
		CreatorID:   creatorID,
		Slug:        fmt.Sprintf("additional-ebook-%d", time.Now().UnixNano()),
	}
	err := database.DB.Create(ebook).Error
	require.NoError(t, err)
	return ebook
}

func simulateCheckoutWithClientCreation(t *testing.T, creator *models.Creator, ebook *models.Ebook) *models.Client {
	// Usar construtor que cria associação automaticamente
	client := models.NewClient(
		fmt.Sprintf("Cliente %d", time.Now().UnixNano()%10000),
		fmt.Sprintf("%011d", time.Now().UnixNano()%99999999999),
		"1990-01-01",
		fmt.Sprintf("client_%d@test.com", time.Now().UnixNano()%100000),
		"11999999999",
		creator,
	)

	// Salvar usando o repositório real
	clientRepo := gorm.NewClientGormRepository()
	err := clientRepo.Save(client)
	require.NoError(t, err)

	return client
}

func createClientWithoutAssociation(t *testing.T) *models.Client {
	// Criar cliente sem associação (simular dados antigos)
	client := &models.Client{
		Name:      fmt.Sprintf("Cliente Sem Associacao %d", time.Now().UnixNano()%10000),
		CPF:       fmt.Sprintf("%011d", time.Now().UnixNano()%99999999999),
		Email:     fmt.Sprintf("sem_associacao_%d@test.com", time.Now().UnixNano()%100000),
		Phone:     "11999999999",
		Birthdate: "1990-01-01",
	}
	err := database.DB.Create(client).Error
	require.NoError(t, err)
	return client
}

func createPurchaseForClient(t *testing.T, clientID, ebookID uint) *models.Purchase {
	purchase := models.NewPurchase(ebookID, clientID)
	err := database.DB.Create(purchase).Error
	require.NoError(t, err)
	return purchase
}

func createClientWithSpecificData(t *testing.T, creator *models.Creator, ebook *models.Ebook, name, email, cpf, phone string) *models.Client {
	client := models.NewClient(name, cpf, "1990-01-01", email, phone, creator)

	clientRepo := gorm.NewClientGormRepository()
	err := clientRepo.Save(client)
	require.NoError(t, err)

	return client
}

func simulateClientListing(t *testing.T, creator *models.Creator) []models.Client {
	return simulateClientListingWithSearch(t, creator, "")
}

func simulateClientListingWithSearch(t *testing.T, creator *models.Creator, searchTerm string) []models.Client {
	// Usar repositório real para buscar clientes
	clientRepo := gorm.NewClientGormRepository()

	pagination := models.NewPagination(1, 10)
	filter := models.ClientFilter{
		Term:       searchTerm,
		Pagination: pagination,
	}

	clients, err := clientRepo.FindClientsByCreator(creator, filter)
	require.NoError(t, err)
	require.NotNil(t, clients)

	return *clients
}

func verifyClientCreatorAssociation(t *testing.T, client *models.Client, creator *models.Creator) {
	// Verificar que existe registro na tabela client_creators
	var count int64
	err := database.DB.Table("client_creators").
		Where("client_id = ? AND creator_id = ?", client.ID, creator.ID).
		Count(&count).Error
	require.NoError(t, err)

	assert.Equal(t, int64(1), count,
		fmt.Sprintf("Cliente %d deve ter associação com Creator %d na tabela client_creators", client.ID, creator.ID))
}

func setupRealCheckoutHandler(t *testing.T) *CheckoutHandler {
	// Esta função seria mais complexa na implementação real
	// Por enquanto, retorna nil já que não está sendo usada
	return nil
}

// TestClientCreatorAssociation_EdgeCases testa casos extremos
func TestClientCreatorAssociation_EdgeCases(t *testing.T) {
	setupIntegrationTestDB(t)
	defer cleanupIntegrationTestDB(t)

	t.Run("Client_With_Same_CPF_Different_Creators", func(t *testing.T) {
		// Testar cenário onde o mesmo CPF tenta se associar a diferentes creators

		// 1. Configurar 2 creators
		creator1, _ := setupCreatorAndEbook(t)
		creator2, _ := setupCreatorAndEbook(t)

		sameCPF := "55555555555"

		// 2. Criar cliente associado ao creator1
		client1 := models.NewClient("Cliente Creator1", sameCPF, "1990-01-01", "client1@test.com", "11999999999", creator1)
		clientRepo := gorm.NewClientGormRepository()
		err := clientRepo.Save(client1)
		require.NoError(t, err)

		// 3. Tentar criar cliente com mesmo CPF para creator2
		client2 := models.NewClient("Cliente Creator2", sameCPF, "1990-01-01", "client2@test.com", "11888888888", creator2)
		err = clientRepo.Save(client2)
		require.NoError(t, err)

		// 4. Verificar que foi reutilizado o mesmo cliente mas com nova associação
		assert.Equal(t, client1.ID, client2.ID, "Mesmo CPF deve reutilizar mesmo cliente")

		// 5. Verificar que cliente agora tem associação com ambos creators
		verifyClientCreatorAssociation(t, client1, creator1)
		verifyClientCreatorAssociation(t, client2, creator2)

		// 6. Verificar que aparece na listagem de ambos
		clients1 := simulateClientListing(t, creator1)
		clients2 := simulateClientListing(t, creator2)

		assert.Len(t, clients1, 1, "Creator1 deve ver o cliente")
		assert.Len(t, clients2, 1, "Creator2 deve ver o cliente")
		assert.Equal(t, clients1[0].ID, clients2[0].ID, "Deve ser o mesmo cliente para ambos")

		t.Logf("✅ Cliente com mesmo CPF pode ser associado a múltiplos creators")
	})

	t.Run("Empty_Search_Returns_All_Clients", func(t *testing.T) {
		// Verificar que busca vazia retorna todos os clientes do creator

		creator, ebook := setupCreatorAndEbook(t)

		// Criar múltiplos clientes
		client1 := simulateCheckoutWithClientCreation(t, creator, ebook)
		client2 := simulateCheckoutWithClientCreation(t, creator, ebook)
		client3 := simulateCheckoutWithClientCreation(t, creator, ebook)

		// Buscar sem termo
		clients := simulateClientListingWithSearch(t, creator, "")

		assert.Len(t, clients, 3, "Busca vazia deve retornar todos os 3 clientes")

		clientIDs := []uint{clients[0].ID, clients[1].ID, clients[2].ID}
		assert.Contains(t, clientIDs, client1.ID)
		assert.Contains(t, clientIDs, client2.ID)
		assert.Contains(t, clientIDs, client3.ID)

		t.Logf("✅ Busca vazia retorna todos os clientes do creator")
	})

	t.Run("Special_Characters_In_Search", func(t *testing.T) {
		// Testar busca com caracteres especiais

		creator, ebook := setupCreatorAndEbook(t)

		// Criar cliente com caracteres especiais
		client := createClientWithSpecificData(t, creator, ebook,
			"José da Silva", "jose@test.com", "66666666666", "11999999999")

		// Testar busca com acentos
		clients := simulateClientListingWithSearch(t, creator, "José")
		assert.Len(t, clients, 1, "Busca por 'José' deve encontrar o cliente")
		assert.Equal(t, client.ID, clients[0].ID)

		// Testar busca parcial
		clients = simulateClientListingWithSearch(t, creator, "Silva")
		assert.Len(t, clients, 1, "Busca por 'Silva' deve encontrar o cliente")

		t.Logf("✅ Busca com caracteres especiais funcionando")
	})
}
