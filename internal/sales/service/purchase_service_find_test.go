package service_test

import (
	"testing"

	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
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

func setupPurchaseServiceTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(
		&authmodel.User{},
		&accountmodel.Creator{},
		&salesmodel.Client{},
		&librarymodel.Ebook{},
		&salesmodel.Purchase{},
	)
	require.NoError(t, err)
	database.DB = db
}

func newPurchaseServiceForTest(t *testing.T) salesvc.PurchaseService {
	t.Helper()
	purchaseRepo := salesrepo.NewPurchaseRepository()
	emailMock := &mocks.MockSalesEmailService{}
	emailMock.On("SendLinkToDownload", mock.Anything).Return()
	return salesvc.NewPurchaseService(purchaseRepo, emailMock)
}

// TestPurchaseService_FindExistingPurchase_Found verifica que uma purchase existente é retornada.
func TestPurchaseService_FindExistingPurchase_Found(t *testing.T) {
	setupPurchaseServiceTestDB(t)

	client := &salesmodel.Client{CPF: "11122233344", Email: "c@test.com", Phone: "11999999999"}
	require.NoError(t, database.DB.Create(client).Error)

	ebook := &librarymodel.Ebook{Title: "Ebook X", Value: 50, Status: true}
	require.NoError(t, database.DB.Create(ebook).Error)

	purchase := &salesmodel.Purchase{EbookID: ebook.ID, ClientID: client.ID, HashID: "hash-abc", DownloadLimit: -1}
	require.NoError(t, database.DB.Create(purchase).Error)

	svc := newPurchaseServiceForTest(t)

	result, err := svc.FindExistingPurchase(ebook.ID, client.ID)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, purchase.ID, result.ID)
}

// TestPurchaseService_FindExistingPurchase_NotFound verifica que erro é retornado quando
// não existe purchase para o par ebook+cliente.
func TestPurchaseService_FindExistingPurchase_NotFound(t *testing.T) {
	setupPurchaseServiceTestDB(t)

	svc := newPurchaseServiceForTest(t)

	result, err := svc.FindExistingPurchase(9999, 9999)

	assert.Error(t, err)
	assert.Nil(t, result)
}
